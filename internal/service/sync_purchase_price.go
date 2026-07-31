package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"

	"supplycore/internal/dto"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/model"
)

var remarkAmountRe = regexp.MustCompile(`\d+(?:\.\d+)?`)

// ParseRemarkPurchaseAmount 从备注中解析采购金额。
// 优先整段为数字；否则取备注中最后一个数字（如「货款70」）。
func ParseRemarkPurchaseAmount(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil && v >= 0 {
		return roundMoney2(v), true
	}
	// 去掉常见后缀再试
	trimmed := strings.TrimRight(s, "元块￥$ ")
	trimmed = strings.TrimSpace(trimmed)
	if v, err := strconv.ParseFloat(trimmed, 64); err == nil && v >= 0 {
		return roundMoney2(v), true
	}
	matches := remarkAmountRe.FindAllString(s, -1)
	if len(matches) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(matches[len(matches)-1], 64)
	if err != nil || v < 0 {
		return 0, false
	}
	return roundMoney2(v), true
}

func remarkTextBySource(o *ordercore.OrderBrief, source string) string {
	if o == nil {
		return ""
	}
	switch source {
	case model.SyncPurchasePriceFenFa:
		return o.FenFaRemark
	case model.SyncPurchasePriceAlloc:
		return o.AllocRemark
	case model.SyncPurchasePriceSeller:
		return o.SellerRemark
	case model.SyncPurchasePricePrinter:
		return o.PrinterRemark
	default:
		return ""
	}
}

func roundMoney2(v float64) float64 {
	return math.Round(v*100) / 100
}

// SyncDropshipPurchasePricesFromOrders 按供应商配置，从订单备注同步采购小计并反推单价。
// source 为空则跳过。按 RefSoID 聚合订单金额，再按数量分摊到明细。
func (s *PurchaseOrderService) SyncDropshipPurchasePricesFromOrders(
	ctx context.Context,
	oc *ordercore.Client,
	bearerToken string,
	poIDs []uint64,
	source string,
) (updated int, err error) {
	source = strings.TrimSpace(source)
	if source == "" || len(poIDs) == 0 {
		return 0, nil
	}
	if oc == nil {
		return 0, fmt.Errorf("OrderCore 未配置")
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	orderCache := map[uint64]*ordercore.OrderBrief{}

	for _, poID := range poIDs {
		if poID == 0 {
			continue
		}
		po, gerr := pr.GetWithItems(poID)
		if gerr != nil {
			return updated, gerr
		}
		if po.PayStatus == model.POPayStatusPaid || po.PayStatus == model.POPayStatusPartial {
			continue
		}
		if po.Status == model.POStatusCompleted || po.Status == model.POStatusCancelled {
			continue
		}

		// 按销售单分组明细
		bySo := map[uint64][]*model.PurchaseOrderItem{}
		for i := range po.Items {
			it := &po.Items[i]
			if it.Cancelled || it.RefSoID == 0 {
				continue
			}
			bySo[it.RefSoID] = append(bySo[it.RefSoID], it)
		}
		priceRows := make([]dto.UpdatePOItemPriceInput, 0)
		for soID, items := range bySo {
			order, ok := orderCache[soID]
			if !ok {
				order, err = oc.GetOrder(ctx, bearerToken, soID)
				if err != nil {
					log.Printf("[sync-purchase-price] get order %d: %v", soID, err)
					continue
				}
				orderCache[soID] = order
			}
			amount, ok := ParseRemarkPurchaseAmount(remarkTextBySource(order, source))
			if !ok {
				continue
			}
			priceRows = append(priceRows, allocateOrderPurchaseToItems(items, amount)...)
		}
		if len(priceRows) == 0 {
			continue
		}
		if _, uerr := s.UpdateItemPrices(poID, &dto.UpdatePOItemPricesInput{Items: priceRows}); uerr != nil {
			return updated, uerr
		}
		updated++
	}
	return updated, nil
}

// SyncPurchasePricesForPOIfConfigured 读取供应商配置，若开启则对该代发单同步采购价。
func (s *PurchaseOrderService) SyncPurchasePricesForPOIfConfigured(
	ctx context.Context,
	oc *ordercore.Client,
	bearerToken string,
	poID uint64,
) (int, error) {
	if poID == 0 || oc == nil {
		return 0, nil
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	po, err := pr.GetByID(poID)
	if err != nil {
		return 0, err
	}
	sup, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(po.SupplierID)
	if err != nil {
		return 0, err
	}
	return s.SyncDropshipPurchasePricesFromOrders(ctx, oc, bearerToken, []uint64{poID}, sup.SyncPurchasePriceFrom)
}

func allocateOrderPurchaseToItems(items []*model.PurchaseOrderItem, orderAmount float64) []dto.UpdatePOItemPriceInput {
	if len(items) == 0 || orderAmount < 0 {
		return nil
	}
	totalQty := 0
	for _, it := range items {
		q := it.Qty
		if q <= 0 {
			q = 1
		}
		totalQty += q
	}
	if totalQty <= 0 {
		totalQty = len(items)
	}
	out := make([]dto.UpdatePOItemPriceInput, 0, len(items))
	var allocated float64
	for i, it := range items {
		q := it.Qty
		if q <= 0 {
			q = 1
		}
		var lineAmt float64
		if i == len(items)-1 {
			lineAmt = roundMoney2(orderAmount - allocated)
		} else {
			lineAmt = roundMoney2(orderAmount * float64(q) / float64(totalQty))
			allocated += lineAmt
		}
		if lineAmt < 0 {
			lineAmt = 0
		}
		unit := roundMoney2(lineAmt / float64(q))
		out = append(out, dto.UpdatePOItemPriceInput{ItemID: it.ID, UnitPrice: unit})
	}
	return out
}
