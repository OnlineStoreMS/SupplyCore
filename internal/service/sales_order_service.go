package service

import (
	"errors"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type SalesOrderService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewSalesOrderService(repos *repo.Repos) *SalesOrderService {
	return &SalesOrderService{repos: repos}
}

func (s *SalesOrderService) ForTenant(tenantID uint64) *SalesOrderService {
	return &SalesOrderService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *SalesOrderService) List(f repo.SOListFilter) ([]dto.SalesOrderListItem, int64, error) {
	list, total, err := s.repos.SalesOrder.ForTenant(s.tenantID).List(f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.SalesOrderListItem, 0, len(list))
	for _, so := range list {
		item := dto.SalesOrderListItem{
			ID: so.ID, SoNo: so.SoNo, TraceID: so.TraceID,
			Status: so.Status, SourceChannel: so.SourceChannel,
			CreatedAt: formatTime(so.CreatedAt),
		}
		if n, err := s.repos.SalesOrder.ForTenant(s.tenantID).CountItems(so.ID); err == nil {
			item.ItemCount = int(n)
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *SalesOrderService) Get(id uint64) (*dto.SalesOrderDetail, error) {
	so, err := s.repos.SalesOrder.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.toDetail(so), nil
}

func (s *SalesOrderService) Create(in *dto.SalesOrderInput) (*dto.SalesOrderDetail, error) {
	sr := s.repos.SalesOrder.ForTenant(s.tenantID)
	soNo, err := sr.NextSoNo()
	if err != nil {
		return nil, err
	}
	so := &model.SalesOrder{
		SoNo: soNo, TraceID: sr.NewTraceID(),
		Status: model.SOStatusDraft,
		SourceChannel: defaultSourceChannel(in.SourceChannel),
		ReceiverName: in.ReceiverName, ReceiverPhone: in.ReceiverPhone,
		Province: in.Province, City: in.City, District: in.District,
		ReceiverAddress: in.ReceiverAddress, Remark: in.Remark,
	}
	items := make([]model.SalesOrderItem, 0, len(in.Items))
	for _, line := range in.Items {
		mode := line.FulfillmentMode
		if mode == "" {
			mode = model.SOFulfillmentDropship
		}
		items = append(items, model.SalesOrderItem{
			SkuID: line.SkuID, Qty: line.Qty,
			FulfillmentMode: mode, Remark: line.Remark,
		})
	}
	if err := sr.Create(so, items); err != nil {
		return nil, err
	}
	return s.Get(so.ID)
}

func (s *SalesOrderService) Update(id uint64, in *dto.SalesOrderInput) (*dto.SalesOrderDetail, error) {
	sr := s.repos.SalesOrder.ForTenant(s.tenantID)
	so, err := sr.GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if so.Status != model.SOStatusDraft {
		return nil, ErrImmutable
	}
	so.SourceChannel = defaultSourceChannel(in.SourceChannel)
	so.ReceiverName = in.ReceiverName
	so.ReceiverPhone = in.ReceiverPhone
	so.Province = in.Province
	so.City = in.City
	so.District = in.District
	so.ReceiverAddress = in.ReceiverAddress
	so.Remark = in.Remark
	if err := sr.Save(so); err != nil {
		return nil, err
	}
	// replace items: delete old + create new via transaction in repo - simplify by updating in place count match
	if len(in.Items) != len(so.Items) {
		return nil, errors.New("暂不支持修改明细行数，请删除后重建")
	}
	for i, line := range in.Items {
		so.Items[i].SkuID = line.SkuID
		so.Items[i].Qty = line.Qty
		if line.FulfillmentMode != "" {
			so.Items[i].FulfillmentMode = line.FulfillmentMode
		}
		so.Items[i].Remark = line.Remark
		if err := sr.SaveItem(&so.Items[i]); err != nil {
			return nil, err
		}
	}
	return s.Get(id)
}

func (s *SalesOrderService) Delete(id uint64) error {
	sr := s.repos.SalesOrder.ForTenant(s.tenantID)
	so, err := sr.GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if so.Status != model.SOStatusDraft {
		return ErrImmutable
	}
	for _, item := range so.Items {
		if item.LinkedPOID > 0 {
			return errors.New("已生成采购单，无法删除")
		}
	}
	so.Status = model.SOStatusCancelled
	return sr.Save(so)
}

func (s *SalesOrderService) Confirm(id uint64) (*dto.SalesOrderDetail, error) {
	sr := s.repos.SalesOrder.ForTenant(s.tenantID)
	so, err := sr.GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if so.Status != model.SOStatusDraft {
		return nil, ErrInvalidStatus
	}
	so.Status = model.SOStatusConfirmed
	if err := sr.Save(so); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *SalesOrderService) toDetail(so *model.SalesOrder) *dto.SalesOrderDetail {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	detail := &dto.SalesOrderDetail{
		ID: so.ID, SoNo: so.SoNo, TraceID: so.TraceID,
		Status: so.Status, SourceChannel: so.SourceChannel,
		ReceiverName: so.ReceiverName, ReceiverPhone: so.ReceiverPhone,
		Province: so.Province, City: so.City, District: so.District,
		ReceiverAddress: so.ReceiverAddress, Remark: so.Remark,
		CreatedAt: formatTime(so.CreatedAt),
		Items: make([]dto.SalesOrderItemDetail, 0, len(so.Items)),
	}
	for _, item := range so.Items {
		line := dto.SalesOrderItemDetail{
			ID: item.ID, SkuID: item.SkuID, Qty: item.Qty,
			FulfillmentMode: item.FulfillmentMode,
			SelectedOfferID: item.SelectedOfferID,
			LinkedPOID: item.LinkedPOID, Remark: item.Remark,
		}
		if item.LinkedPOID > 0 {
			if po, err := pr.GetByID(item.LinkedPOID); err == nil {
				line.LinkedPoNo = po.PoNo
			}
		}
		detail.Items = append(detail.Items, line)
	}
	return detail
}

func defaultSourceChannel(v string) string {
	if v == "" {
		return "manual"
	}
	return v
}
