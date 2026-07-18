package service

import (
	"errors"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type PurchaseExtService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewPurchaseExtService(repos *repo.Repos) *PurchaseExtService {
	return &PurchaseExtService{repos: repos}
}

func (s *PurchaseExtService) ForTenant(tenantID uint64) *PurchaseExtService {
	return &PurchaseExtService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

// ---- Purchase Accounts ----

func (s *PurchaseExtService) ListAccounts(keyword string, page, pageSize int) ([]dto.PurchaseAccountDTO, int64, error) {
	list, total, err := s.repos.PurchaseAccount.ForTenant(s.tenantID).List(keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseAccountDTO, 0, len(list))
	for i := range list {
		out = append(out, s.toAccountDTO(&list[i]))
	}
	return out, total, nil
}

func (s *PurchaseExtService) CreateAccount(in *dto.PurchaseAccountDTO) (*dto.PurchaseAccountDTO, error) {
	if in.Channel == "" {
		in.Channel = "other"
	}
	if in.Status == "" {
		in.Status = model.PurchaseAccountStatusActive
	}
	if in.AuthStatus == "" {
		in.AuthStatus = "unauthorized"
	}
	m := &model.PurchaseAccount{
		Channel: in.Channel, AccountAlias: in.AccountAlias, AccountName: in.AccountName,
		IsPrimary: in.IsPrimary, Status: in.Status, AuthStatus: in.AuthStatus,
		OperatorName: in.OperatorName, Remark: in.Remark,
	}
	if err := s.repos.PurchaseAccount.ForTenant(s.tenantID).Create(m); err != nil {
		return nil, err
	}
	d := s.toAccountDTO(m)
	return &d, nil
}

func (s *PurchaseExtService) UpdateAccount(id uint64, in *dto.PurchaseAccountDTO) (*dto.PurchaseAccountDTO, error) {
	m, err := s.repos.PurchaseAccount.ForTenant(s.tenantID).GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if in.Channel != "" {
		m.Channel = in.Channel
	}
	if in.AccountAlias != "" {
		m.AccountAlias = in.AccountAlias
	}
	if in.AccountName != "" {
		m.AccountName = in.AccountName
	}
	m.IsPrimary = in.IsPrimary
	if in.Status != "" {
		m.Status = in.Status
	}
	if in.AuthStatus != "" {
		m.AuthStatus = in.AuthStatus
	}
	m.OperatorName = in.OperatorName
	m.Remark = in.Remark
	if err := s.repos.PurchaseAccount.ForTenant(s.tenantID).Save(m); err != nil {
		return nil, err
	}
	d := s.toAccountDTO(m)
	return &d, nil
}

func (s *PurchaseExtService) DeleteAccount(id uint64) error {
	return s.repos.PurchaseAccount.ForTenant(s.tenantID).Delete(id)
}

func (s *PurchaseExtService) toAccountDTO(m *model.PurchaseAccount) dto.PurchaseAccountDTO {
	d := dto.PurchaseAccountDTO{
		ID: m.ID, Channel: m.Channel, AccountAlias: m.AccountAlias, AccountName: m.AccountName,
		IsPrimary: m.IsPrimary, Status: m.Status, AuthStatus: m.AuthStatus,
		OperatorName: m.OperatorName, Remark: m.Remark,
		CreatedAt: formatTime(m.CreatedAt), UpdatedAt: formatTime(m.UpdatedAt),
	}
	if m.LastSyncAt != nil {
		t := formatTime(*m.LastSyncAt)
		d.LastSyncAt = &t
	}
	return d
}

// ---- Inbound ----

func (s *PurchaseExtService) ListInbounds(status, keyword string, page, pageSize int) ([]dto.PurchaseInboundListItem, int64, error) {
	list, total, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).List(status, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseInboundListItem, 0, len(list))
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	for _, m := range list {
		item := dto.PurchaseInboundListItem{
			ID: m.ID, InboundNo: m.InboundNo, Status: m.Status, POID: m.POID, PoNo: m.PoNo,
			SupplierID: m.SupplierID, WarehouseName: m.WarehouseName, TotalQty: m.TotalQty,
			TotalAmount: m.TotalAmount, TrackingNo: m.TrackingNo, CreatorName: m.CreatorName,
			CreatedAt: formatTime(m.CreatedAt),
		}
		if m.SupplierID > 0 {
			if sup, err := sr.GetByID(m.SupplierID); err == nil {
				item.SupplierName = sup.Name
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *PurchaseExtService) GetInbound(id uint64) (*dto.PurchaseInboundDetail, error) {
	m, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.toInboundDetail(m), nil
}

func (s *PurchaseExtService) CreateInbound(in *dto.PurchaseInboundInput, creatorName string) (*dto.PurchaseInboundDetail, error) {
	if len(in.Items) == 0 {
		return nil, errors.New("入库明细不能为空")
	}
	var poNo string
	supplierID := in.SupplierID
	if in.POID > 0 {
		po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(in.POID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		poNo = po.PoNo
		if supplierID == 0 {
			supplierID = po.SupplierID
		}
	}
	items := make([]model.PurchaseInboundItem, 0, len(in.Items))
	totalQty := 0
	totalAmount := 0.0
	for _, it := range in.Items {
		if it.InboundQty <= 0 {
			continue
		}
		qc := it.QCQty
		if qc == 0 {
			qc = it.InboundQty
		}
		line := float64(it.InboundQty) * it.UnitPrice
		items = append(items, model.PurchaseInboundItem{
			POItemID: it.POItemID, SkuID: it.SkuID, SkuCode: it.SkuCode, SkuName: it.SkuName,
			PurchaseQty: it.PurchaseQty, QCQty: qc, RejectQty: it.RejectQty, InboundQty: it.InboundQty,
			UnitPrice: it.UnitPrice, LineAmount: line, LocationCode: it.LocationCode, Remark: it.Remark,
		})
		totalQty += it.InboundQty
		totalAmount += line
	}
	if len(items) == 0 {
		return nil, errors.New("有效入库数量不能为空")
	}
	no, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).NextNo()
	if err != nil {
		return nil, err
	}
	m := &model.PurchaseInbound{
		InboundNo: no, Status: model.InboundStatusDraft, POID: in.POID, PoNo: poNo,
		SupplierID: supplierID, WarehouseID: in.WarehouseID, WarehouseName: in.WarehouseName,
		TrackingNo: in.TrackingNo, PlatformOrderNo: in.PlatformOrderNo,
		TotalQty: totalQty, TotalAmount: totalAmount, CreatorName: creatorName,
		BuyerName: in.BuyerName, Remark: in.Remark,
	}
	if err := s.repos.PurchaseInbound.ForTenant(s.tenantID).Create(m, items); err != nil {
		return nil, err
	}
	return s.GetInbound(m.ID)
}

func (s *PurchaseExtService) ApproveInboundWH(id uint64, auditor string) (*dto.PurchaseInboundDetail, error) {
	m, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if m.Status != model.InboundStatusDraft && m.Status != model.InboundStatusPendingWH {
		return nil, errors.New("当前状态不可入库审核")
	}
	now := time.Now()
	m.Status = model.InboundStatusPendingFinance
	m.WHAuditorName = auditor
	m.WHAuditedAt = &now
	if err := s.repos.PurchaseInbound.ForTenant(s.tenantID).Save(m); err != nil {
		return nil, err
	}
	// 回写采购单已收数量
	if m.POID > 0 {
		_ = s.applyInboundToPO(m)
	}
	return s.GetInbound(id)
}

func (s *PurchaseExtService) ApproveInboundFinance(id uint64, auditor string) (*dto.PurchaseInboundDetail, error) {
	m, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if m.Status != model.InboundStatusPendingFinance {
		return nil, errors.New("当前状态不可财务审核")
	}
	now := time.Now()
	m.Status = model.InboundStatusCompleted
	m.FinAuditorName = auditor
	m.FinAuditedAt = &now
	if err := s.repos.PurchaseInbound.ForTenant(s.tenantID).Save(m); err != nil {
		return nil, err
	}
	return s.GetInbound(id)
}

func (s *PurchaseExtService) VoidInbound(id uint64) error {
	m, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if m.Status == model.InboundStatusCompleted {
		return errors.New("已完成入库单不可作废")
	}
	m.Status = model.InboundStatusVoid
	return s.repos.PurchaseInbound.ForTenant(s.tenantID).Save(m)
}

func (s *PurchaseExtService) applyInboundToPO(m *model.PurchaseInbound) error {
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(m.POID)
	if err != nil {
		return err
	}
	itemMap := map[uint64]*model.PurchaseOrderItem{}
	for i := range po.Items {
		itemMap[po.Items[i].ID] = &po.Items[i]
	}
	allReceived := true
	anyReceived := false
	for _, it := range m.Items {
		if it.POItemID == 0 {
			continue
		}
		poi, ok := itemMap[it.POItemID]
		if !ok {
			continue
		}
		poi.ReceivedQty += it.InboundQty
		if poi.ReceivedQty > poi.Qty {
			poi.ReceivedQty = poi.Qty
		}
		if err := s.repos.PurchaseOrder.ForTenant(s.tenantID).SaveItem(poi); err != nil {
			return err
		}
	}
	for _, poi := range po.Items {
		if poi.ReceivedQty > 0 {
			anyReceived = true
		}
		if poi.ReceivedQty < poi.Qty {
			allReceived = false
		}
	}
	if allReceived {
		po.Status = model.POStatusCompleted
		now := time.Now()
		po.CompletedAt = &now
	} else if anyReceived {
		po.Status = model.POStatusPartialReceived
	}
	return s.repos.PurchaseOrder.ForTenant(s.tenantID).Save(po)
}

func (s *PurchaseExtService) toInboundDetail(m *model.PurchaseInbound) *dto.PurchaseInboundDetail {
	list := dto.PurchaseInboundListItem{
		ID: m.ID, InboundNo: m.InboundNo, Status: m.Status, POID: m.POID, PoNo: m.PoNo,
		SupplierID: m.SupplierID, WarehouseName: m.WarehouseName, TotalQty: m.TotalQty,
		TotalAmount: m.TotalAmount, TrackingNo: m.TrackingNo, CreatorName: m.CreatorName,
		CreatedAt: formatTime(m.CreatedAt),
	}
	if m.SupplierID > 0 {
		if sup, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(m.SupplierID); err == nil {
			list.SupplierName = sup.Name
		}
	}
	items := make([]dto.PurchaseInboundItemDTO, 0, len(m.Items))
	for _, it := range m.Items {
		items = append(items, dto.PurchaseInboundItemDTO{
			ID: it.ID, POItemID: it.POItemID, SkuID: it.SkuID, SkuCode: it.SkuCode, SkuName: it.SkuName,
			PurchaseQty: it.PurchaseQty, QCQty: it.QCQty, RejectQty: it.RejectQty, InboundQty: it.InboundQty,
			UnitPrice: it.UnitPrice, LineAmount: it.LineAmount, LocationCode: it.LocationCode, Remark: it.Remark,
		})
	}
	d := &dto.PurchaseInboundDetail{
		PurchaseInboundListItem: list,
		PlatformOrderNo: m.PlatformOrderNo, BuyerName: m.BuyerName, Remark: m.Remark,
		WHAuditorName: m.WHAuditorName, FinAuditorName: m.FinAuditorName, Items: items,
	}
	if m.WHAuditedAt != nil {
		d.WHAuditedAt = formatTime(*m.WHAuditedAt)
	}
	if m.FinAuditedAt != nil {
		d.FinAuditedAt = formatTime(*m.FinAuditedAt)
	}
	return d
}

// ---- Package receive ----

func (s *PurchaseExtService) ListPackageReceives(keyword string, page, pageSize int) ([]dto.PackageReceiveDTO, int64, error) {
	list, total, err := s.repos.PackageReceive.ForTenant(s.tenantID).List(keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PackageReceiveDTO, 0, len(list))
	for _, m := range list {
		out = append(out, dto.PackageReceiveDTO{
			ID: m.ID, WarehouseName: m.WarehouseName, Carrier: m.Carrier, TrackingNo: m.TrackingNo,
			PackageType: m.PackageType, POID: m.POID, PoNo: m.PoNo, InboundID: m.InboundID,
			InboundNo: m.InboundNo, ScannerName: m.ScannerName, Remark: m.Remark, CreatedAt: formatTime(m.CreatedAt),
		})
	}
	return out, total, nil
}

func (s *PurchaseExtService) ScanPackage(in *dto.PackageReceiveInput, scanner string) (*dto.PackageReceiveDTO, error) {
	if in.TrackingNo == "" {
		return nil, errors.New("物流单号不能为空")
	}
	pkgType := in.PackageType
	if pkgType == "" {
		pkgType = "normal"
	}
	m := &model.PackageReceiveRecord{
		WarehouseID: in.WarehouseID, WarehouseName: in.WarehouseName, Carrier: in.Carrier,
		TrackingNo: in.TrackingNo, PackageType: pkgType, POID: in.POID, PoNo: in.PoNo,
		ScannerName: scanner, Remark: in.Remark,
	}
	if in.POID > 0 && in.PoNo == "" {
		if po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(in.POID); err == nil {
			m.PoNo = po.PoNo
		}
	}
	if err := s.repos.PackageReceive.ForTenant(s.tenantID).Create(m); err != nil {
		return nil, err
	}
	d := dto.PackageReceiveDTO{
		ID: m.ID, WarehouseName: m.WarehouseName, Carrier: m.Carrier, TrackingNo: m.TrackingNo,
		PackageType: m.PackageType, POID: m.POID, PoNo: m.PoNo, ScannerName: m.ScannerName,
		Remark: m.Remark, CreatedAt: formatTime(m.CreatedAt),
	}
	return &d, nil
}

func (s *PurchaseExtService) CreateInboundFromPackage(receiveID uint64, creator string) (*dto.PurchaseInboundDetail, error) {
	rec, err := s.repos.PackageReceive.ForTenant(s.tenantID).GetByID(receiveID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if rec.POID == 0 {
		return nil, errors.New("收货记录未关联采购单，无法生成入库单")
	}
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(rec.POID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PurchaseInboundItemInput, 0, len(po.Items))
	for _, it := range po.Items {
		remain := it.Qty - it.ReceivedQty
		if remain <= 0 {
			continue
		}
		items = append(items, dto.PurchaseInboundItemInput{
			POItemID: it.ID, SkuID: it.SkuID, SkuCode: it.SupplierSkuCode,
			PurchaseQty: it.Qty, InboundQty: remain, UnitPrice: it.UnitPrice,
		})
	}
	detail, err := s.CreateInbound(&dto.PurchaseInboundInput{
		POID: po.ID, SupplierID: po.SupplierID, WarehouseName: rec.WarehouseName,
		TrackingNo: rec.TrackingNo, BuyerName: po.BuyerName, Items: items,
	}, creator)
	if err != nil {
		return nil, err
	}
	rec.InboundID = detail.ID
	rec.InboundNo = detail.InboundNo
	_ = s.repos.PackageReceive.ForTenant(s.tenantID).Save(rec)
	return detail, nil
}

// ---- Returns ----

func (s *PurchaseExtService) ListReturns(status, keyword string, page, pageSize int) ([]dto.PurchaseReturnListItem, int64, error) {
	list, total, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).List(status, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.PurchaseReturnListItem, 0, len(list))
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	for _, m := range list {
		item := dto.PurchaseReturnListItem{
			ID: m.ID, ReturnNo: m.ReturnNo, Status: m.Status, InboundID: m.InboundID, InboundNo: m.InboundNo,
			SupplierID: m.SupplierID, WarehouseName: m.WarehouseName, TotalQty: m.TotalQty,
			TotalAmount: m.TotalAmount, ActualAmount: m.ActualAmount, CreatorName: m.CreatorName,
			CreatedAt: formatTime(m.CreatedAt),
		}
		if m.SupplierID > 0 {
			if sup, err := sr.GetByID(m.SupplierID); err == nil {
				item.SupplierName = sup.Name
			}
		}
		out = append(out, item)
	}
	return out, total, nil
}

func (s *PurchaseExtService) GetReturn(id uint64) (*dto.PurchaseReturnDetail, error) {
	m, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.toReturnDetail(m), nil
}

func (s *PurchaseExtService) CreateReturn(in *dto.PurchaseReturnInput, creator string) (*dto.PurchaseReturnDetail, error) {
	if len(in.Items) == 0 {
		return nil, errors.New("退回明细不能为空")
	}
	inboundNo := ""
	supplierID := in.SupplierID
	if in.InboundID > 0 {
		ib, err := s.repos.PurchaseInbound.ForTenant(s.tenantID).GetWithItems(in.InboundID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		inboundNo = ib.InboundNo
		if supplierID == 0 {
			supplierID = ib.SupplierID
		}
	}
	items := make([]model.PurchaseReturnItem, 0, len(in.Items))
	totalQty := 0
	totalAmount := 0.0
	actualAmount := 0.0
	for _, it := range in.Items {
		if it.ReturnQty <= 0 {
			continue
		}
		amt := float64(it.ReturnQty) * it.UnitPrice
		act := it.ActualAmount
		if act == 0 {
			act = amt
		}
		items = append(items, model.PurchaseReturnItem{
			InboundItemID: it.InboundItemID, SkuID: it.SkuID, SkuCode: it.SkuCode, SkuName: it.SkuName,
			OriginalQty: it.OriginalQty, ReturnQty: it.ReturnQty, UnitPrice: it.UnitPrice,
			ReturnAmount: amt, ActualAmount: act, Remark: it.Remark,
		})
		totalQty += it.ReturnQty
		totalAmount += amt
		actualAmount += act
	}
	if len(items) == 0 {
		return nil, errors.New("有效退回数量不能为空")
	}
	no, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).NextNo()
	if err != nil {
		return nil, err
	}
	m := &model.PurchaseReturn{
		ReturnNo: no, Status: model.ReturnStatusDraft, InboundID: in.InboundID, InboundNo: inboundNo,
		SupplierID: supplierID, WarehouseID: in.WarehouseID, WarehouseName: in.WarehouseName,
		TrackingNo: in.TrackingNo, TotalQty: totalQty, TotalAmount: totalAmount, ActualAmount: actualAmount,
		CreatorName: creator, BuyerName: in.BuyerName, Remark: in.Remark,
	}
	if err := s.repos.PurchaseReturn.ForTenant(s.tenantID).Create(m, items); err != nil {
		return nil, err
	}
	return s.GetReturn(m.ID)
}

func (s *PurchaseExtService) ApproveReturn(id uint64, auditor string) (*dto.PurchaseReturnDetail, error) {
	m, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if m.Status != model.ReturnStatusDraft && m.Status != model.ReturnStatusPendingReturn {
		return nil, errors.New("当前状态不可退回审核")
	}
	now := time.Now()
	m.Status = model.ReturnStatusPendingFinance
	m.AuditorName = auditor
	m.AuditedAt = &now
	if err := s.repos.PurchaseReturn.ForTenant(s.tenantID).Save(m); err != nil {
		return nil, err
	}
	return s.GetReturn(id)
}

func (s *PurchaseExtService) ApproveReturnFinance(id uint64, auditor string) (*dto.PurchaseReturnDetail, error) {
	m, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if m.Status != model.ReturnStatusPendingFinance {
		return nil, errors.New("当前状态不可财务审核")
	}
	now := time.Now()
	m.Status = model.ReturnStatusCompleted
	m.FinAuditorName = auditor
	m.FinAuditedAt = &now
	if err := s.repos.PurchaseReturn.ForTenant(s.tenantID).Save(m); err != nil {
		return nil, err
	}
	return s.GetReturn(id)
}

func (s *PurchaseExtService) VoidReturn(id uint64) error {
	m, err := s.repos.PurchaseReturn.ForTenant(s.tenantID).GetWithItems(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if m.Status == model.ReturnStatusCompleted {
		return errors.New("已完成退回单不可作废")
	}
	m.Status = model.ReturnStatusVoid
	return s.repos.PurchaseReturn.ForTenant(s.tenantID).Save(m)
}

func (s *PurchaseExtService) toReturnDetail(m *model.PurchaseReturn) *dto.PurchaseReturnDetail {
	list := dto.PurchaseReturnListItem{
		ID: m.ID, ReturnNo: m.ReturnNo, Status: m.Status, InboundID: m.InboundID, InboundNo: m.InboundNo,
		SupplierID: m.SupplierID, WarehouseName: m.WarehouseName, TotalQty: m.TotalQty,
		TotalAmount: m.TotalAmount, ActualAmount: m.ActualAmount, CreatorName: m.CreatorName,
		CreatedAt: formatTime(m.CreatedAt),
	}
	if m.SupplierID > 0 {
		if sup, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(m.SupplierID); err == nil {
			list.SupplierName = sup.Name
		}
	}
	items := make([]dto.PurchaseReturnItemDTO, 0, len(m.Items))
	for _, it := range m.Items {
		items = append(items, dto.PurchaseReturnItemDTO{
			ID: it.ID, InboundItemID: it.InboundItemID, SkuID: it.SkuID, SkuCode: it.SkuCode, SkuName: it.SkuName,
			OriginalQty: it.OriginalQty, ReturnQty: it.ReturnQty, UnitPrice: it.UnitPrice,
			ReturnAmount: it.ReturnAmount, ActualAmount: it.ActualAmount, Remark: it.Remark,
		})
	}
	d := &dto.PurchaseReturnDetail{
		PurchaseReturnListItem: list, TrackingNo: m.TrackingNo, BuyerName: m.BuyerName, Remark: m.Remark,
		AuditorName: m.AuditorName, FinAuditorName: m.FinAuditorName, Items: items,
	}
	if m.AuditedAt != nil {
		d.AuditedAt = formatTime(*m.AuditedAt)
	}
	if m.FinAuditedAt != nil {
		d.FinAuditedAt = formatTime(*m.FinAuditedAt)
	}
	return d
}

// ---- Suggestions (缺货/预警/无库存) ----

func (s *PurchaseExtService) ListSuggestions(source string) ([]dto.StockoutSuggestionItem, error) {
	// 采购建议依赖 WarehouseCore 库存上下限；对接前返回空列表占位。
	_ = source
	return []dto.StockoutSuggestionItem{}, nil
}
