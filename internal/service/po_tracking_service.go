package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type POTrackingService struct {
	repos    *repo.Repos
	oc       *ordercore.Client
	tenantID uint64
}

func NewPOTrackingService(repos *repo.Repos, oc *ordercore.Client) *POTrackingService {
	return &POTrackingService{repos: repos, oc: oc}
}

func (s *POTrackingService) ForTenant(tenantID uint64) *POTrackingService {
	return &POTrackingService{repos: s.repos, oc: s.oc, tenantID: repo.NormalizeTenantID(tenantID)}
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

// SyncShipmentsFromOrders 从订单中心拉取快递单号/发货状态，写入代发采购单物流。
func (s *POTrackingService) SyncShipmentsFromOrders(ctx context.Context, poID uint64, bearerToken string, in *dto.SyncShipmentsFromOrdersInput) (*dto.SyncShipmentsFromOrdersResult, error) {
	if s.oc == nil {
		return nil, fmt.Errorf("OrderCore 未配置")
	}
	po, err := s.ensurePOTrackable(poID)
	if err != nil {
		return nil, err
	}
	if po.FulfillmentType != model.POFulfillmentDropship {
		return nil, fmt.Errorf("仅代发采购单支持从订单中心同步物流")
	}
	full, err := s.repos.PurchaseOrder.ForTenant(s.tenantID).GetWithItems(poID)
	if err != nil {
		return nil, err
	}

	type soGroup struct {
		refSoID uint64
		orderNo string
		items   []model.PurchaseOrderItem
	}
	groups := map[uint64]*soGroup{}
	filterSoID := uint64(0)
	if in != nil {
		filterSoID = in.RefSoID
	}
	for _, it := range full.Items {
		if it.Cancelled {
			continue
		}
		refSoID := it.RefSoID
		orderNo := strings.TrimSpace(it.RefOrderNo)
		if refSoID == 0 {
			refSoID = full.RefSoID
		}
		if orderNo == "" {
			trace := strings.TrimSpace(full.RefTraceID)
			if trace != "" && !strings.Contains(trace, ",") {
				orderNo = trace
			}
		}
		if refSoID == 0 {
			continue
		}
		if filterSoID > 0 && refSoID != filterSoID {
			continue
		}
		g := groups[refSoID]
		if g == nil {
			g = &soGroup{refSoID: refSoID, orderNo: orderNo}
			groups[refSoID] = g
		}
		g.items = append(g.items, it)
		if g.orderNo == "" && orderNo != "" {
			g.orderNo = orderNo
		}
	}

	out := &dto.SyncShipmentsFromOrdersResult{}
	if len(groups) == 0 {
		out.Skipped = 1
		out.Errors = append(out.Errors, "没有可同步的销售单明细")
		return out, nil
	}

	sr := s.repos.Shipment.ForTenant(s.tenantID)
	for _, g := range groups {
		order, gerr := s.oc.GetOrder(ctx, bearerToken, g.refSoID)
		if gerr != nil {
			out.Skipped++
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", coalesceOrderNo(g.orderNo, g.refSoID), gerr))
			continue
		}
		if g.orderNo == "" {
			g.orderNo = order.OrderNo
		}
		receiverName := strings.TrimSpace(order.BuyerName)
		receiverPhone := strings.TrimSpace(order.BuyerPhone)
		receiverAddr := ordercore.FormatReceiverAddress(order.Address)
		if order.Address != nil {
			if n := strings.TrimSpace(order.Address.Name); n != "" {
				receiverName = n
			}
			if p := strings.TrimSpace(order.Address.Phone); p != "" {
				receiverPhone = p
			}
		}

		type logistics struct {
			trackingNo string
			carrier    string
			shippedAt  *time.Time
			remark     string
		}
		logs := make([]logistics, 0)
		seenTrack := map[string]struct{}{}
		for _, sh := range order.Shipments {
			tn := strings.TrimSpace(sh.ExpressNo)
			if tn == "" {
				continue
			}
			if _, ok := seenTrack[tn]; ok {
				continue
			}
			seenTrack[tn] = struct{}{}
			var shippedAt *time.Time
			if sh.ShippedAt != nil && strings.TrimSpace(*sh.ShippedAt) != "" {
				if t := parseDateTime(*sh.ShippedAt); t != nil {
					shippedAt = t
				}
			}
			logs = append(logs, logistics{
				trackingNo: tn,
				carrier:    strings.TrimSpace(sh.ExpressCompany),
				shippedAt:  shippedAt,
				remark:     fmt.Sprintf("同步自订单 %s", order.OrderNo),
			})
		}
		if len(logs) == 0 && order.ShipStatus == "shipped" {
			logs = append(logs, logistics{
				trackingNo: fmt.Sprintf("SYNC-%s", order.OrderNo),
				carrier:    "订单中心已发货",
				remark:     fmt.Sprintf("同步自订单 %s（无快递单号）", order.OrderNo),
			})
		}
		if len(logs) == 0 {
			out.Skipped++
			continue
		}

		// 该销售单剩余可发数量：全部挂到第一条物流（代发通常一单一票）
		primary := logs[0]
		existing, ferr := sr.FindByTrackingNo(poID, primary.trackingNo)
		if ferr != nil && !errors.Is(ferr, gorm.ErrRecordNotFound) {
			out.Skipped++
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", order.OrderNo, ferr))
			continue
		}
		if existing != nil {
			changed := false
			if strings.TrimSpace(existing.ReceiverName) == "" && receiverName != "" {
				existing.ReceiverName = receiverName
				changed = true
			}
			if strings.TrimSpace(existing.ReceiverPhone) == "" && receiverPhone != "" {
				existing.ReceiverPhone = receiverPhone
				changed = true
			}
			if strings.TrimSpace(existing.ReceiverAddress) == "" && receiverAddr != "" {
				existing.ReceiverAddress = receiverAddr
				changed = true
			}
			if strings.TrimSpace(existing.CarrierName) == "" && primary.carrier != "" {
				existing.CarrierName = primary.carrier
				changed = true
			}
			if existing.Status == model.ShipmentStatusPending {
				existing.Status = model.ShipmentStatusShipped
				now := time.Now()
				if primary.shippedAt != nil {
					existing.ShippedAt = primary.shippedAt
				} else {
					existing.ShippedAt = &now
				}
				changed = true
			}
			if changed {
				if err := sr.Save(existing); err != nil {
					out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", order.OrderNo, err))
					out.Skipped++
					continue
				}
				out.Updated++
			} else {
				out.Skipped++
			}
			continue
		}

		inputs := make([]dto.ShipmentItemInput, 0, len(g.items))
		for _, it := range g.items {
			if it.ID == 0 {
				continue
			}
			inputs = append(inputs, dto.ShipmentItemInput{POItemID: it.ID, Qty: it.Qty})
		}
		// buildShipmentItems 会按剩余量校验；先算剩余
		remainInputs := make([]dto.ShipmentItemInput, 0, len(inputs))
		shippedQty := map[uint64]int{}
		allShip, _ := sr.ListByPO(poID)
		for _, sh := range allShip {
			for _, it := range sh.Items {
				shippedQty[it.POItemID] += it.Qty
			}
		}
		for _, inItem := range inputs {
			poItemQty := 0
			for _, it := range g.items {
				if it.ID == inItem.POItemID {
					poItemQty = it.Qty
					break
				}
			}
			remain := poItemQty - shippedQty[inItem.POItemID]
			if remain <= 0 {
				continue
			}
			remainInputs = append(remainInputs, dto.ShipmentItemInput{POItemID: inItem.POItemID, Qty: remain})
		}
		if len(remainInputs) == 0 {
			out.Skipped++
			continue
		}
		items, berr := s.buildShipmentItems(poID, remainInputs)
		if berr != nil {
			out.Skipped++
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", order.OrderNo, berr))
			continue
		}
		no, nerr := sr.NextShipmentNo()
		if nerr != nil {
			out.Skipped++
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", order.OrderNo, nerr))
			continue
		}
		now := time.Now()
		shippedAt := &now
		if primary.shippedAt != nil {
			shippedAt = primary.shippedAt
		}
		sh := &model.PurchaseShipment{
			POID:            poID,
			ShipmentNo:      no,
			Status:          model.ShipmentStatusShipped,
			CarrierName:     primary.carrier,
			TrackingNo:      primary.trackingNo,
			ShippedAt:       shippedAt,
			ReceiverName:    receiverName,
			ReceiverPhone:   receiverPhone,
			ReceiverAddress: receiverAddr,
			Remark:          primary.remark,
		}
		if err := sr.Create(sh, items); err != nil {
			out.Skipped++
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", order.OrderNo, err))
			continue
		}
		out.Created++

		// 额外快递单号：仅登记单号（明细已挂在首票）
		for i := 1; i < len(logs); i++ {
			extra := logs[i]
			if ex, _ := sr.FindByTrackingNo(poID, extra.trackingNo); ex != nil {
				continue
			}
			eno, _ := sr.NextShipmentNo()
			esh := &model.PurchaseShipment{
				POID:            poID,
				ShipmentNo:      eno,
				Status:          model.ShipmentStatusShipped,
				CarrierName:     extra.carrier,
				TrackingNo:      extra.trackingNo,
				ShippedAt:       shippedAt,
				ReceiverName:    receiverName,
				ReceiverPhone:   receiverPhone,
				ReceiverAddress: receiverAddr,
				Remark:          extra.remark + "（附加运单）",
			}
			if err := sr.Create(esh, nil); err != nil {
				out.Errors = append(out.Errors, fmt.Sprintf("%s extra: %v", order.OrderNo, err))
				continue
			}
			out.Created++
		}
	}

	if err := s.syncShipmentStatus(poID); err != nil {
		out.Errors = append(out.Errors, err.Error())
	}
	return out, nil
}

func coalesceOrderNo(orderNo string, soID uint64) string {
	if strings.TrimSpace(orderNo) != "" {
		return orderNo
	}
	return fmt.Sprintf("#%d", soID)
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
		POID: poID, PaymentID: in.PaymentID, ShipmentID: in.ShipmentID,
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
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetWithItems(poID)
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
	shippedQty := map[uint64]int{}
	hasInTransit, hasShipped, allDelivered := false, false, true
	for _, sh := range list {
		switch sh.Status {
		case model.ShipmentStatusInTransit:
			hasInTransit = true
			hasShipped = true
		case model.ShipmentStatusShipped, model.ShipmentStatusPending:
			hasShipped = true
		case model.ShipmentStatusDelivered:
			hasShipped = true
		}
		if sh.Status != model.ShipmentStatusDelivered {
			allDelivered = false
		}
		for _, it := range sh.Items {
			shippedQty[it.POItemID] += it.Qty
		}
	}
	fullyShipped := true
	activeLines := 0
	for _, it := range po.Items {
		if it.Cancelled {
			continue
		}
		activeLines++
		if shippedQty[it.ID] < it.Qty {
			fullyShipped = false
			break
		}
	}
	if activeLines == 0 {
		fullyShipped = false
	}

	switch {
	case allDelivered && fullyShipped:
		if po.FulfillmentType == model.POFulfillmentDropship {
			po.Status = model.POStatusCompleted
		} else {
			po.Status = model.POStatusPartialReceived
		}
	case hasInTransit || (fullyShipped && hasShipped):
		// 明细已全部发出：标运输中，不再标「部分发货」
		po.Status = model.POStatusInTransit
	case hasShipped:
		po.Status = model.POStatusPartialShipped
	}
	return pr.Save(po)
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
		ID: a.ID, PoID: a.POID, PaymentID: a.PaymentID, ShipmentID: a.ShipmentID,
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
