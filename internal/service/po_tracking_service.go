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

type POTrackingService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewPOTrackingService(repos *repo.Repos) *POTrackingService {
	return &POTrackingService{repos: repos}
}

func (s *POTrackingService) ForTenant(tenantID uint64) *POTrackingService {
	return &POTrackingService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *POTrackingService) ensurePOTrackable(poID uint64) (*model.PurchaseOrder, error) {
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetByID(poID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if po.Status == model.POStatusDraft || po.Status == model.POStatusCancelled {
		return nil, ErrInvalidStatus
	}
	return po, nil
}

// --- Shipments ---

func (s *POTrackingService) ListShipments(poID uint64) ([]dto.ShipmentDetail, error) {
	if _, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetByID(poID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	list, err := s.repos.Shipment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ShipmentDetail, 0, len(list))
	for i := range list {
		out = append(out, s.toShipmentDetail(&list[i]))
	}
	return out, nil
}

func (s *POTrackingService) CreateShipment(poID uint64, in *dto.ShipmentInput) (*dto.ShipmentDetail, error) {
	if _, err := s.ensurePOTrackable(poID); err != nil {
		return nil, err
	}
	if len(in.Items) == 0 {
		return nil, ErrBadRequest
	}
	if strings.TrimSpace(in.TrackingNo) == "" {
		return nil, ErrBadRequest
	}
	no, err := s.repos.Shipment.ForTenant(s.tenantID).NextShipmentNo()
	if err != nil {
		return nil, err
	}
	sh := &model.PurchaseShipment{
		POID: poID, ShipmentNo: no, Status: model.ShipmentStatusPending,
		CarrierCode: in.CarrierCode, CarrierName: in.CarrierName,
		TrackingNo: in.TrackingNo, ShipFromAddressID: in.ShipFromAddressID,
		ReceiverName: in.ReceiverName, ReceiverPhone: in.ReceiverPhone,
		ReceiverAddress: in.ReceiverAddress, Remark: in.Remark,
	}
	if d := parseDate(in.ExpectedArrivalDate); d != nil {
		sh.ExpectedArrivalDate = d
	}
	items, err := s.buildShipmentItems(poID, in.Items)
	if err != nil {
		return nil, err
	}
	if err := s.repos.Shipment.ForTenant(s.tenantID).Create(sh, items); err != nil {
		return nil, err
	}
	if err := s.syncShipmentStatus(poID); err != nil {
		return nil, err
	}
	detail := s.toShipmentDetail(sh)
	return &detail, nil
}

func (s *POTrackingService) UpdateShipmentStatus(poID, shipmentID uint64, status string) (*dto.ShipmentDetail, error) {
	if _, err := s.ensurePOTrackable(poID); err != nil {
		return nil, err
	}
	sr := s.repos.Shipment.ForTenant(s.tenantID)
	sh, err := sr.GetByID(poID, shipmentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !isValidShipmentStatus(status) {
		return nil, ErrBadRequest
	}
	now := time.Now()
	sh.Status = status
	switch status {
	case model.ShipmentStatusShipped:
		if sh.ShippedAt == nil {
			sh.ShippedAt = &now
		}
	case model.ShipmentStatusInTransit:
		if sh.ShippedAt == nil {
			sh.ShippedAt = &now
		}
	case model.ShipmentStatusDelivered:
		sh.DeliveredAt = &now
	}
	if err := sr.Save(sh); err != nil {
		return nil, err
	}
	if err := s.syncShipmentStatus(poID); err != nil {
		return nil, err
	}
	detail := s.toShipmentDetail(sh)
	return &detail, nil
}

func (s *POTrackingService) DeleteShipment(poID, shipmentID uint64) error {
	if _, err := s.ensurePOTrackable(poID); err != nil {
		return err
	}
	if err := s.repos.Shipment.ForTenant(s.tenantID).Delete(poID, shipmentID); err != nil {
		return err
	}
	return s.syncShipmentStatus(poID)
}

// --- Payments ---

func (s *POTrackingService) ListPayments(poID uint64) ([]dto.PaymentDetail, error) {
	if _, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetByID(poID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	list, err := s.repos.Payment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PaymentDetail, 0, len(list))
	for i := range list {
		out = append(out, s.toPaymentDetail(&list[i]))
	}
	return out, nil
}

func (s *POTrackingService) CreatePayment(poID uint64, in *dto.PaymentInput) (*dto.PaymentDetail, error) {
	po, err := s.ensurePOTrackable(poID)
	if err != nil {
		return nil, err
	}
	pay := &model.PurchasePayment{
		POID: poID, PayAmount: in.PayAmount,
		PayMethod: in.PayMethod, PayAccount: in.PayAccount,
		PayeeAccount: in.PayeeAccount, PayeeName: in.PayeeName,
		PayStatus: defaultPayRecordStatus(in.PayStatus),
		Remark: in.Remark,
	}
	if in.PaidAt != "" {
		if t := parseDateTime(in.PaidAt); t != nil {
			pay.PaidAt = t
		}
	} else if pay.PayStatus == model.POPayStatusPaid {
		now := time.Now()
		pay.PaidAt = &now
	}
	if err := s.repos.Payment.ForTenant(s.tenantID).Create(pay); err != nil {
		return nil, err
	}
	if err := s.syncPayStatus(po); err != nil {
		return nil, err
	}
	detail := s.toPaymentDetail(pay)
	return &detail, nil
}

func (s *POTrackingService) UpdatePayment(poID, paymentID uint64, in *dto.PaymentInput) (*dto.PaymentDetail, error) {
	po, err := s.ensurePOTrackable(poID)
	if err != nil {
		return nil, err
	}
	pr := s.repos.Payment.ForTenant(s.tenantID)
	pay, err := pr.GetByID(poID, paymentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	pay.PayAmount = in.PayAmount
	pay.PayMethod = in.PayMethod
	pay.PayAccount = in.PayAccount
	pay.PayeeAccount = in.PayeeAccount
	pay.PayeeName = in.PayeeName
	pay.PayStatus = defaultPayRecordStatus(in.PayStatus)
	pay.Remark = in.Remark
	if in.PaidAt != "" {
		pay.PaidAt = parseDateTime(in.PaidAt)
	}
	if err := pr.Save(pay); err != nil {
		return nil, err
	}
	if err := s.syncPayStatus(po); err != nil {
		return nil, err
	}
	detail := s.toPaymentDetail(pay)
	return &detail, nil
}

func (s *POTrackingService) DeletePayment(poID, paymentID uint64) error {
	po, err := s.ensurePOTrackable(poID)
	if err != nil {
		return err
	}
	atts, err := s.repos.Attachment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil {
		return err
	}
	for _, a := range atts {
		if a.PaymentID == paymentID {
			if err := s.repos.Attachment.ForTenant(s.tenantID).Delete(poID, a.ID); err != nil {
				return err
			}
		}
	}
	if err := s.repos.Payment.ForTenant(s.tenantID).Delete(poID, paymentID); err != nil {
		return err
	}
	return s.syncPayStatus(po)
}

// --- Attachments ---

func (s *POTrackingService) ListAttachments(poID uint64) ([]dto.AttachmentDetail, error) {
	if _, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetByID(poID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	list, err := s.repos.Attachment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.AttachmentDetail, 0, len(list))
	for i := range list {
		out = append(out, s.toAttachmentDetail(&list[i]))
	}
	return out, nil
}

func (s *POTrackingService) CreateAttachment(poID, uploadedBy uint64, in *dto.AttachmentInput) (*dto.AttachmentDetail, error) {
	if _, err := s.ensurePOTrackable(poID); err != nil {
		return nil, err
	}
	a := &model.PurchaseAttachment{
		POID: poID, PaymentID: in.PaymentID,
		FileType: in.FileType, FileName: in.FileName,
		FileURL: in.FileURL, UploadedBy: uploadedBy, Remark: in.Remark,
	}
	if err := s.repos.Attachment.ForTenant(s.tenantID).Create(a); err != nil {
		return nil, err
	}
	detail := s.toAttachmentDetail(a)
	return &detail, nil
}

func (s *POTrackingService) DeleteAttachment(poID, attachmentID uint64) error {
	if _, err := s.ensurePOTrackable(poID); err != nil {
		return err
	}
	return s.repos.Attachment.ForTenant(s.tenantID).Delete(poID, attachmentID)
}

// --- sync ---

func (s *POTrackingService) syncPayStatus(po *model.PurchaseOrder) error {
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	fresh, err := pr.GetByID(po.ID)
	if err != nil {
		return err
	}
	sum, err := s.repos.Payment.ForTenant(s.tenantID).SumPaid(fresh.ID)
	if err != nil {
		return err
	}
	switch {
	case sum <= 0:
		fresh.PayStatus = model.POPayStatusUnpaid
	case sum+0.001 < fresh.TotalAmount:
		fresh.PayStatus = model.POPayStatusPartial
	default:
		// 付款累计金额 >= 采购总额时自动标记已付清
		fresh.PayStatus = model.POPayStatusPaid
		if fresh.Status == model.POStatusOrdered {
			fresh.Status = model.POStatusPaid
		}
	}
	*po = *fresh
	return pr.Save(fresh)
}

func (s *POTrackingService) syncShipmentStatus(poID uint64) error {
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetByID(poID)
	if err != nil {
		return err
	}
	if po.Status == model.POStatusCompleted || po.Status == model.POStatusCancelled || po.Status == model.POStatusDraft {
		return nil
	}
	list, err := s.repos.Shipment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil || len(list) == 0 {
		return err
	}
	hasInTransit, hasShipped, allDelivered := false, false, true
	for _, sh := range list {
		switch sh.Status {
		case model.ShipmentStatusInTransit:
			hasInTransit = true
		case model.ShipmentStatusShipped:
			hasShipped = true
		}
		if sh.Status != model.ShipmentStatusDelivered {
			allDelivered = false
		}
	}
	switch {
	case hasInTransit:
		po.Status = model.POStatusInTransit
	case hasShipped:
		po.Status = model.POStatusPartialShipped
	case allDelivered && len(list) > 0:
		if po.FulfillmentType == model.POFulfillmentStockIn {
			po.Status = model.POStatusPartialReceived
		}
	}
	return s.repos.PurchaseOrder.ForTenant(s.tenantID).Save(po)
}

func (s *POTrackingService) buildShipmentItems(poID uint64, inputs []ShipmentItemInput) ([]model.PurchaseShipmentItem, error) {
	po, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(poID)
	if err != nil {
		return nil, err
	}
	itemMap := map[uint64]model.PurchaseOrderItem{}
	for _, it := range po.Items {
		itemMap[it.ID] = it
	}
	shippedQty := map[uint64]int{}
	existing, err := s.repos.Shipment.ForTenant(s.tenantID).ListByPO(poID)
	if err != nil {
		return nil, err
	}
	for _, sh := range existing {
		for _, it := range sh.Items {
			shippedQty[it.POItemID] += it.Qty
		}
	}
	seen := map[uint64]struct{}{}
	items := make([]model.PurchaseShipmentItem, 0, len(inputs))
	for _, in := range inputs {
		if _, dup := seen[in.POItemID]; dup {
			return nil, ErrBadRequest
		}
		seen[in.POItemID] = struct{}{}
		poItem, ok := itemMap[in.POItemID]
		if !ok {
			return nil, ErrNotFound
		}
		remain := poItem.Qty - shippedQty[in.POItemID]
		if in.Qty <= 0 || in.Qty > remain {
			return nil, ErrBadRequest
		}
		items = append(items, model.PurchaseShipmentItem{
			POItemID: in.POItemID, Qty: in.Qty,
		})
	}
	return items, nil
}

type ShipmentItemInput = dto.ShipmentItemInput

func (s *POTrackingService) toShipmentDetail(sh *model.PurchaseShipment) dto.ShipmentDetail {
	d := dto.ShipmentDetail{
		ID: sh.ID, PoID: sh.POID, ShipmentNo: sh.ShipmentNo, Status: sh.Status,
		CarrierCode: sh.CarrierCode, CarrierName: sh.CarrierName, TrackingNo: sh.TrackingNo,
		ReceiverName: sh.ReceiverName, ReceiverPhone: sh.ReceiverPhone,
		ReceiverAddress: sh.ReceiverAddress, Remark: sh.Remark,
		CreatedAt: formatTime(sh.CreatedAt),
		Items: make([]dto.ShipmentItemDetail, 0, len(sh.Items)),
	}
	if sh.ShippedAt != nil {
		d.ShippedAt = formatTimePtr(sh.ShippedAt)
	}
	if sh.ExpectedArrivalDate != nil {
		d.ExpectedArrivalDate = sh.ExpectedArrivalDate.Format("2006-01-02")
	}
	if sh.DeliveredAt != nil {
		d.DeliveredAt = formatTimePtr(sh.DeliveredAt)
	}
	po, _ := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(sh.POID)
	skuByItem := map[uint64]uint64{}
	if po != nil {
		for _, it := range po.Items {
			skuByItem[it.ID] = it.SkuID
		}
	}
	for _, it := range sh.Items {
		d.Items = append(d.Items, dto.ShipmentItemDetail{
			ID: it.ID, POItemID: it.POItemID, SkuID: skuByItem[it.POItemID], Qty: it.Qty,
		})
	}
	return d
}

func (s *POTrackingService) toPaymentDetail(p *model.PurchasePayment) dto.PaymentDetail {
	d := dto.PaymentDetail{
		ID: p.ID, PoID: p.POID, PayAmount: p.PayAmount,
		PayMethod: p.PayMethod, PayAccount: p.PayAccount,
		PayeeAccount: p.PayeeAccount, PayeeName: p.PayeeName,
		PayStatus: p.PayStatus, Remark: p.Remark,
		CreatedAt: formatTime(p.CreatedAt),
	}
	if p.PaidAt != nil {
		d.PaidAt = formatTimePtr(p.PaidAt)
	}
	return d
}

func (s *POTrackingService) toAttachmentDetail(a *model.PurchaseAttachment) dto.AttachmentDetail {
	return dto.AttachmentDetail{
		ID: a.ID, PoID: a.POID, PaymentID: a.PaymentID,
		FileType: a.FileType, FileName: a.FileName, FileURL: a.FileURL,
		UploadedBy: a.UploadedBy, Remark: a.Remark,
		CreatedAt: formatTime(a.CreatedAt),
	}
}

func isValidShipmentStatus(s string) bool {
	switch s {
	case model.ShipmentStatusPending, model.ShipmentStatusShipped,
		model.ShipmentStatusInTransit, model.ShipmentStatusDelivered, model.ShipmentStatusException:
		return true
	}
	return false
}

func defaultPayRecordStatus(s string) string {
	if s == "" {
		return model.POPayStatusPaid
	}
	return s
}

func parseDateTime(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t
		}
		if t, err := time.Parse(layout, s); err == nil {
			local := t.In(time.Local)
			return &local
		}
	}
	return nil
}
