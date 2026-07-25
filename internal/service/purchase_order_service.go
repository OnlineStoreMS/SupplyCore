package service

import (
	"errors"
	"strings"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type PurchaseOrderService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewPurchaseOrderService(repos *repo.Repos) *PurchaseOrderService {
	return &PurchaseOrderService{repos: repos}
}

func (s *PurchaseOrderService) ForTenant(tenantID uint64) *PurchaseOrderService {
	return &PurchaseOrderService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *PurchaseOrderService) List(f repo.POListFilter) ([]dto.PurchaseOrderListItem, int64, error) {
	list, total, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).List(f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseOrderListItem, 0, len(list))
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	for _, po := range list {
		item := dto.PurchaseOrderListItem{
			ID: po.ID, PoNo: po.PoNo, SupplierID: po.SupplierID,
			Status: po.Status, PayStatus: po.PayStatus,
			FulfillmentType: po.FulfillmentType,
			TotalAmount: po.TotalAmount, Currency: po.Currency,
			RefSoID: po.RefSoID, RefTraceID: po.RefTraceID,
			CreatedAt: formatTime(po.CreatedAt),
		}
		if po.OrderedAt != nil {
			item.OrderedAt = formatTimePtr(po.OrderedAt)
		}
		if sup, err := sr.GetByID(po.SupplierID); err == nil {
			item.SupplierName = sup.Name
		}
		if n, err := pr.CountItems(po.ID); err == nil {
			item.ItemCount = int(n)
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *PurchaseOrderService) Get(id uint64) (*dto.PurchaseOrderDetail, error) {
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.toDetail(po), nil
}

func (s *PurchaseOrderService) Create(in *dto.PurchaseOrderInput, buyerID uint64, buyerName string) (*dto.PurchaseOrderDetail, error) {
	if _, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(in.SupplierID); errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	ft := defaultFulfillment(in.FulfillmentType)
	items, total, err := s.buildItems(in.SupplierID, ft, in.Items)
	if err != nil {
		return nil, err
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	const maxAttempts = 5
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		poNo, err := pr.NextPoNo()
		if err != nil {
			return nil, err
		}
		po := &model.PurchaseOrder{
			PoNo: poNo, SupplierID: in.SupplierID,
			Status: model.POStatusDraft, TotalAmount: total,
			Currency: defaultCurrency(in.Currency),
			FulfillmentType: ft,
			RefSoID: in.RefSoID, RefTraceID: in.RefTraceID,
			BuyerID: buyerID, BuyerName: buyerName,
			PayStatus: model.POPayStatusUnpaid, Remark: in.Remark,
		}
		if d := parseDate(in.ExpectedArrivalDate); d != nil {
			po.ExpectedArrivalDate = d
		}
		lineItems := make([]model.PurchaseOrderItem, len(items))
		copy(lineItems, items)
		for i := range lineItems {
			lineItems[i].ID = 0
			lineItems[i].POID = 0
		}
		if err := pr.Create(po, lineItems); err != nil {
			lastErr = err
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		return s.Get(po.ID)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("创建采购单失败：单号冲突")
}

func (s *PurchaseOrderService) Update(id uint64, in *dto.PurchaseOrderInput) (*dto.PurchaseOrderDetail, error) {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if po.Status != model.POStatusDraft {
		return nil, ErrImmutable
	}
	if _, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(in.SupplierID); errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	ft := defaultFulfillment(in.FulfillmentType)
	items, total, err := s.buildItems(in.SupplierID, ft, in.Items)
	if err != nil {
		return nil, err
	}
	po.SupplierID = in.SupplierID
	po.TotalAmount = total
	po.Currency = defaultCurrency(in.Currency)
	po.FulfillmentType = ft
	po.WarehouseID = in.WarehouseID
	po.RefSoID = in.RefSoID
	po.RefTraceID = in.RefTraceID
	po.Remark = in.Remark
	po.ExpectedArrivalDate = parseDate(in.ExpectedArrivalDate)
	if err := pr.Save(po); err != nil {
		return nil, err
	}
	if err := pr.ReplaceItems(id, items); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *PurchaseOrderService) Delete(id uint64) error {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if po.Status != model.POStatusDraft {
		return ErrImmutable
	}
	return pr.Delete(id)
}

func (s *PurchaseOrderService) Submit(id uint64) (*dto.PurchaseOrderDetail, error) {
	return s.transition(id, model.POStatusDraft, model.POStatusOrdered, func(po *model.PurchaseOrder) {
		now := time.Now()
		po.OrderedAt = &now
	})
}

func (s *PurchaseOrderService) MarkPaid(id uint64) (*dto.PurchaseOrderDetail, error) {
	return s.transition(id, model.POStatusOrdered, model.POStatusPaid, func(po *model.PurchaseOrder) {
		po.PayStatus = model.POPayStatusPaid
	})
}

func (s *PurchaseOrderService) Complete(id uint64) (*dto.PurchaseOrderDetail, error) {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{
		model.POStatusPaid: true, model.POStatusPartialShipped: true,
		model.POStatusInTransit: true, model.POStatusPartialReceived: true,
	}
	if !allowed[po.Status] {
		return nil, ErrInvalidStatus
	}
	now := time.Now()
	po.Status = model.POStatusCompleted
	po.CompletedAt = &now
	if err := pr.Save(po); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *PurchaseOrderService) Cancel(id uint64) (*dto.PurchaseOrderDetail, error) {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if po.Status != model.POStatusDraft && po.Status != model.POStatusOrdered {
		return nil, ErrInvalidStatus
	}
	po.Status = model.POStatusCancelled
	if err := pr.Save(po); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *PurchaseOrderService) transition(id uint64, from, to string, apply func(*model.PurchaseOrder)) (*dto.PurchaseOrderDetail, error) {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if po.Status != from {
		return nil, ErrInvalidStatus
	}
	po.Status = to
	if apply != nil {
		apply(po)
	}
	if err := pr.Save(po); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *PurchaseOrderService) buildItems(supplierID uint64, fulfillmentType string, inputs []dto.PurchaseOrderItemInput) ([]model.PurchaseOrderItem, float64, error) {
	or := s.repos.Offer.ForTenant(s.tenantID)
	dropship := fulfillmentType == model.POFulfillmentDropship
	items := make([]model.PurchaseOrderItem, 0, len(inputs))
	var total float64
	for _, in := range inputs {
		if in.Qty <= 0 {
			return nil, 0, ErrBadRequest
		}
		if !dropship && in.SkuID == 0 && in.OfferID == 0 {
			return nil, 0, errors.New("请选择 SKU 或供货报价")
		}
		item := model.PurchaseOrderItem{
			SkuID: in.SkuID, OfferID: in.OfferID,
			ProductName:     strings.TrimSpace(in.ProductName),
			SupplierSkuCode: in.SupplierSkuCode,
			Qty: in.Qty, Remark: in.Remark,
		}
		if in.OfferID > 0 {
			offer, err := or.GetByID(in.OfferID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, ErrNotFound
			}
			if err != nil {
				return nil, 0, err
			}
			if offer.SupplierID != supplierID {
				return nil, 0, errors.New("供货报价与采购单供应商不一致")
			}
			if item.SkuID == 0 {
				item.SkuID = offer.SkuID
			} else if item.SkuID != offer.SkuID {
				return nil, 0, errors.New("SKU 与供货报价不匹配")
			}
			if item.SupplierSkuCode == "" {
				item.SupplierSkuCode = offer.SupplierSkuCode
			}
			if in.UnitPrice <= 0 {
				item.UnitPrice = offer.SupplyPrice
			} else {
				item.UnitPrice = in.UnitPrice
			}
		} else {
			// 代发草稿允许单价为 0（OMS 推送后在 SupplyCore 补价）
			if in.UnitPrice < 0 || (!dropship && in.UnitPrice <= 0) {
				return nil, 0, errors.New("请填写单价或选择供货报价")
			}
			item.UnitPrice = in.UnitPrice
		}
		item.LineAmount = float64(item.Qty) * item.UnitPrice
		total += item.LineAmount
		items = append(items, item)
	}
	return items, total, nil
}

func (s *PurchaseOrderService) toDetail(po *model.PurchaseOrder) *dto.PurchaseOrderDetail {
	detail := &dto.PurchaseOrderDetail{
		ID: po.ID, PoNo: po.PoNo, SupplierID: po.SupplierID,
		Status: po.Status, TotalAmount: po.TotalAmount, Currency: po.Currency,
		WarehouseID: po.WarehouseID, FulfillmentType: po.FulfillmentType,
		RefSoID: po.RefSoID, RefTraceID: po.RefTraceID, BuyerID: po.BuyerID, BuyerName: po.BuyerName,
		PayStatus: po.PayStatus, Remark: po.Remark,
		CreatedAt: formatTime(po.CreatedAt),
		Items: make([]dto.PurchaseOrderItemDetail, 0, len(po.Items)),
	}
	if po.ExpectedArrivalDate != nil {
		detail.ExpectedArrivalDate = po.ExpectedArrivalDate.Format("2006-01-02")
	}
	if po.OrderedAt != nil {
		detail.OrderedAt = formatTimePtr(po.OrderedAt)
	}
	if po.CompletedAt != nil {
		detail.CompletedAt = formatTimePtr(po.CompletedAt)
	}
	if sup, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(po.SupplierID); err == nil {
		detail.SupplierName = sup.Name
		detail.SupplierCode = sup.Code
	}
	for _, it := range po.Items {
		detail.Items = append(detail.Items, dto.PurchaseOrderItemDetail{
			ID: it.ID, SkuID: it.SkuID, OfferID: it.OfferID,
			ProductName: it.ProductName, SupplierSkuCode: it.SupplierSkuCode, Qty: it.Qty,
			UnitPrice: it.UnitPrice, LineAmount: it.LineAmount,
			ReceivedQty: it.ReceivedQty, Remark: it.Remark,
		})
	}
	return detail
}

func defaultCurrency(c string) string {
	if c == "" {
		return "CNY"
	}
	return c
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "23505")
}

func defaultFulfillment(t string) string {
	if t == "" {
		return model.POFulfillmentStockIn
	}
	return t
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return nil
	}
	return &t
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTime(*t)
}
