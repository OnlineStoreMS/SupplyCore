package service

import (
	"errors"
	"fmt"
	"strings"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type SourcingService struct {
	repos    *repo.Repos
	poSvc    *PurchaseOrderService
	tenantID uint64
}

func NewSourcingService(repos *repo.Repos, poSvc *PurchaseOrderService) *SourcingService {
	return &SourcingService{repos: repos, poSvc: poSvc}
}

func (s *SourcingService) ForTenant(tenantID uint64) *SourcingService {
	return &SourcingService{
		repos: s.repos, poSvc: s.poSvc.ForTenant(tenantID),
		tenantID: repo.NormalizeTenantID(tenantID),
	}
}

func (s *SourcingService) Evaluate(soID uint64) (*dto.SourcingEvaluateResp, error) {
	so, err := s.repos.SalesOrder.ForTenant(s.tenantID).GetWithItems(soID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	resp := &dto.SourcingEvaluateResp{
		SoID: so.ID, TraceID: so.TraceID,
		Lines: make([]dto.SourcingLinePlan, 0, len(so.Items)),
	}
	or := s.repos.Offer.ForTenant(s.tenantID)
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	for _, item := range so.Items {
		line := dto.SourcingLinePlan{
			SoItemID: item.ID, SkuID: item.SkuID, Qty: item.Qty,
			FulfillmentMode: item.FulfillmentMode,
		}
		if item.FulfillmentMode == model.SOFulfillmentSelf {
			line.NeedsPO = false
			line.Message = "自发货，无需采购单"
			resp.Lines = append(resp.Lines, line)
			continue
		}
		if item.LinkedPOID > 0 {
			line.NeedsPO = false
			line.Message = "已关联采购单"
			resp.Lines = append(resp.Lines, line)
			continue
		}
		line.NeedsPO = true
		offers, err := or.ListBySku(item.SkuID, true)
		if err != nil {
			return nil, err
		}
		for _, offer := range offers {
			if !offer.SupportsDropship {
				continue
			}
			opt := dto.SupplyOptionOffer{
				OfferID: offer.ID, SupplierID: offer.SupplierID,
				SupplierSkuCode: offer.SupplierSkuCode, SupplyPrice: offer.SupplyPrice,
				Currency: offer.Currency, SupportsDropship: offer.SupportsDropship,
				SupportsSelfStock: offer.SupportsSelfStock, LeadTimeDays: offer.LeadTimeDays,
				IsPrimary: offer.IsPrimary, Priority: offer.Priority,
			}
			if sup, err := sr.GetByID(offer.SupplierID); err == nil {
				opt.SupplierName = sup.Name
				opt.SupplierCode = sup.Code
			}
			line.Offers = append(line.Offers, opt)
			if line.RecommendedID == 0 && offer.IsPrimary {
				line.RecommendedID = offer.ID
			}
		}
		if line.RecommendedID == 0 && len(line.Offers) > 0 {
			line.RecommendedID = line.Offers[0].OfferID
		}
		if len(line.Offers) == 0 {
			line.Message = "无可用代发供货报价"
		}
		resp.Lines = append(resp.Lines, line)
	}
	return resp, nil
}

func (s *SourcingService) CreatePOsFromSO(soID uint64, in *dto.SourcingCreatePOInput, buyerID uint64, buyerName string) (*dto.SourcingCreatePOResp, error) {
	so, err := s.repos.SalesOrder.ForTenant(s.tenantID).GetWithItems(soID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if so.Status == model.SOStatusCancelled {
		return nil, ErrInvalidStatus
	}
	in.RefSoID = so.ID
	if in.RefTraceID == "" {
		in.RefTraceID = so.TraceID
	}
	return s.createPOsFromSelections(so, in, buyerID, buyerName)
}

func (s *SourcingService) CreateDropshipPO(in *dto.SourcingCreatePOInput, buyerID uint64, buyerName string) (*dto.SourcingCreatePOResp, error) {
	if in.RefTraceID == "" {
		return nil, errors.New("refTraceId 不能为空")
	}
	var so *model.SalesOrder
	if in.RefSoID > 0 {
		var err error
		so, err = s.repos.SalesOrder.ForTenant(s.tenantID).GetWithItems(in.RefSoID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	return s.createPOsFromSelections(so, in, buyerID, buyerName)
}

type poGroup struct {
	supplierID uint64
	items      []dto.PurchaseOrderItemInput
	soItemIDs  []uint64
	offerIDs   []uint64
}

func (s *SourcingService) createPOsFromSelections(so *model.SalesOrder, in *dto.SourcingCreatePOInput, buyerID uint64, buyerName string) (*dto.SourcingCreatePOResp, error) {
	if len(in.Selections) == 0 {
		return nil, errors.New("请选择供货报价")
	}
	itemByID := map[uint64]model.SalesOrderItem{}
	if so != nil {
		for _, item := range so.Items {
			itemByID[item.ID] = item
		}
	}
	or := s.repos.Offer.ForTenant(s.tenantID)
	groups := map[uint64]*poGroup{}
	for _, sel := range in.Selections {
		qty := sel.Qty
		if soItem, ok := itemByID[sel.SoItemID]; ok {
			if soItem.LinkedPOID > 0 {
				return nil, fmt.Errorf("行 %d 已关联采购单", sel.SoItemID)
			}
			if soItem.FulfillmentMode == model.SOFulfillmentSelf {
				return nil, fmt.Errorf("行 %d 为自发货，无需采购", sel.SoItemID)
			}
			if qty <= 0 {
				qty = soItem.Qty
			}
		} else if so != nil {
			return nil, fmt.Errorf("销售单行 %d 不存在", sel.SoItemID)
		} else if qty <= 0 {
			return nil, errors.New("外部代发需指定 qty")
		}
		offer, err := or.GetByID(sel.OfferID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		if !offer.SupportsDropship {
			return nil, fmt.Errorf("报价 %d 不支持代发", sel.OfferID)
		}
		g := groups[offer.SupplierID]
		if g == nil {
			g = &poGroup{supplierID: offer.SupplierID}
			groups[offer.SupplierID] = g
		}
		g.items = append(g.items, dto.PurchaseOrderItemInput{
			SkuID: offer.SkuID, OfferID: offer.ID,
			SupplierSkuCode: offer.SupplierSkuCode, Qty: qty,
			UnitPrice: offer.SupplyPrice,
		})
		g.soItemIDs = append(g.soItemIDs, sel.SoItemID)
		g.offerIDs = append(g.offerIDs, sel.OfferID)
	}
	receiverRemark := buildReceiverRemark(so, in.Receiver)
	resp := &dto.SourcingCreatePOResp{
		RefSoID: in.RefSoID, RefTraceID: in.RefTraceID,
		PoIDs: make([]uint64, 0), PoNos: make([]string, 0),
	}
	sr := s.repos.SalesOrder.ForTenant(s.tenantID)
	for _, g := range groups {
		remark := receiverRemark
		if so != nil && so.Remark != "" {
			remark = strings.TrimSpace(so.Remark + "\n" + remark)
		}
		poIn := &dto.PurchaseOrderInput{
			SupplierID: g.supplierID,
			FulfillmentType: model.POFulfillmentDropship,
			RefSoID: in.RefSoID, RefTraceID: in.RefTraceID,
			Remark: remark, Items: g.items,
		}
		po, err := s.poSvc.Create(poIn, buyerID, buyerName)
		if err != nil {
			return nil, err
		}
		resp.PoIDs = append(resp.PoIDs, po.ID)
		resp.PoNos = append(resp.PoNos, po.PoNo)
		if so != nil {
			for i, soItemID := range g.soItemIDs {
				if item, ok := itemByID[soItemID]; ok {
					item.LinkedPOID = po.ID
					item.SelectedOfferID = g.offerIDs[i]
					_ = sr.SaveItem(&item)
					itemByID[soItemID] = item
				}
			}
		}
		if in.AutoSubmit {
			if _, err := s.poSvc.Submit(po.ID); err != nil {
				return nil, err
			}
		}
	}
	if so != nil {
		allSourced := true
		for _, item := range so.Items {
			if item.FulfillmentMode == model.SOFulfillmentDropship {
				updated := itemByID[item.ID]
				if updated.LinkedPOID == 0 {
					allSourced = false
					break
				}
			}
		}
		if allSourced {
			so.Status = model.SOStatusSourced
			_ = sr.Save(so)
		}
	}
	return resp, nil
}

func buildReceiverRemark(so *model.SalesOrder, override *dto.SourcingReceiver) string {
	name, phone, province, city, district, address := "", "", "", "", "", ""
	if override != nil {
		name, phone = override.Name, override.Phone
		province, city, district, address = override.Province, override.City, override.District, override.Address
	}
	if so != nil {
		if name == "" {
			name = so.ReceiverName
		}
		if phone == "" {
			phone = so.ReceiverPhone
		}
		if province == "" {
			province = so.Province
		}
		if city == "" {
			city = so.City
		}
		if district == "" {
			district = so.District
		}
		if address == "" {
			address = so.ReceiverAddress
		}
	}
	if name == "" && phone == "" && address == "" {
		return "代发采购"
	}
	return fmt.Sprintf("代发收货：%s %s %s%s%s %s", name, phone, province, city, district, address)
}
