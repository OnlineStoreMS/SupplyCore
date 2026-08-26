package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
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

type soPurchaseMeta struct {
	id     uint64
	order  *ordercore.OrderBrief
	amount float64
	hasAmt bool
	expr   string
	items  []*model.PurchaseOrderItem
}

// SyncDropshipPurchasePricesFromOrders 按供应商配置，从订单备注同步采购小计并反推单价。
// source 为空则跳过。
// 快递助手合单发货时分发备注常复制到各子单：同运单号且金额相同/仅一侧填写时，整包只计一次。
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

		bySo := map[uint64][]*model.PurchaseOrderItem{}
		for i := range po.Items {
			it := &po.Items[i]
			if it.Cancelled || it.RefSoID == 0 {
				continue
			}
			bySo[it.RefSoID] = append(bySo[it.RefSoID], it)
		}
		if len(bySo) == 0 {
			continue
		}

		metas := make([]soPurchaseMeta, 0, len(bySo))
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
			amount, hasAmt := ParseRemarkPurchaseAmount(remarkTextBySource(order, source))
			metas = append(metas, soPurchaseMeta{
				id: soID, order: order, amount: amount, hasAmt: hasAmt,
				expr: primaryExpressNo(order), items: items,
			})
		}

		groups := map[string][]soPurchaseMeta{}
		groupOrder := make([]string, 0)
		for _, m := range metas {
			key := m.expr
			if key == "" {
				key = fmt.Sprintf("so:%d", m.id)
			} else {
				key = "ex:" + key
			}
			if _, ok := groups[key]; !ok {
				groupOrder = append(groupOrder, key)
			}
			groups[key] = append(groups[key], m)
		}

		priceRows := make([]dto.UpdatePOItemPriceInput, 0)
		for _, key := range groupOrder {
			group := groups[key]
			if len(group) == 0 {
				continue
			}
			amount, ok := resolveGroupPurchaseAmount(group)
			if !ok {
				for _, m := range group {
					if !m.hasAmt {
						continue
					}
					priceRows = append(priceRows, allocateOrderPurchaseToItems(m.items, m.amount)...)
				}
				continue
			}
			// 合单：金额只落在第一单（按销售单 ID 升序），其余明细单价置 0
			sort.Slice(group, func(i, j int) bool { return group[i].id < group[j].id })
			primary := group[0]
			if len(group) > 1 {
				log.Printf("[sync-purchase-price] merge-ship fenfa on first so po=%d key=%s primary=%d amount=%.2f others=%d",
					poID, key, primary.id, amount, len(group)-1)
			}
			priceRows = append(priceRows, allocateOrderPurchaseToItems(primary.items, amount)...)
			for _, m := range group[1:] {
				priceRows = append(priceRows, allocateOrderPurchaseToItems(m.items, 0)...)
			}
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

func primaryExpressNo(o *ordercore.OrderBrief) string {
	if o == nil {
		return ""
	}
	for _, sh := range o.Shipments {
		no := strings.TrimSpace(sh.ExpressNo)
		if no != "" {
			return no
		}
	}
	return ""
}

// resolveGroupPurchaseAmount 合单组内：有金额的备注若一致（或仅一侧填写），整包用该金额一次。
func resolveGroupPurchaseAmount(group []soPurchaseMeta) (float64, bool) {
	var amounts []float64
	for _, m := range group {
		if !m.hasAmt {
			continue
		}
		amounts = append(amounts, m.amount)
	}
	if len(amounts) == 0 {
		return 0, false
	}
	first := amounts[0]
	for _, a := range amounts[1:] {
		if math.Abs(a-first) > 0.009 {
			// 同运单但备注金额不同：无法安全合并
			return 0, false
		}
	}
	// 仅当「合单」（多销售单）或单侧有金额时走整包一次；单销售单也适用
	return first, true
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
		line := lineAmt
		out = append(out, dto.UpdatePOItemPriceInput{
			ItemID:     it.ID,
			UnitPrice:  unit,
			LineAmount: &line,
		})
	}
	return out
}
