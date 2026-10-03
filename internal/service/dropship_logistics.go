package service

import (
	"fmt"
	"strings"
	"time"

	"supplycore/internal/integrations/ordercore"
)

type ocLogistics struct {
	trackingNo string
	carrier    string
	shippedAt  *time.Time
	remark     string
}

func shipmentTracking(sh ordercore.OrderShipmentBrief) string {
	return strings.TrimSpace(sh.ExpressNo)
}

func shipmentShippedAt(sh ordercore.OrderShipmentBrief) *time.Time {
	if sh.ShippedAt == nil || strings.TrimSpace(*sh.ShippedAt) == "" {
		return nil
	}
	return parseDateTime(*sh.ShippedAt)
}

func shipmentHasPOItem(sh ordercore.OrderShipmentBrief, poRefItemIDs map[uint64]struct{}) bool {
	for _, it := range sh.Items {
		if it.OrderItemID > 0 {
			if _, ok := poRefItemIDs[it.OrderItemID]; ok {
				return true
			}
		}
	}
	return false
}

func shipmentHasAnyItem(sh ordercore.OrderShipmentBrief) bool {
	for _, it := range sh.Items {
		if it.OrderItemID > 0 {
			return true
		}
	}
	return false
}

// collectDropshipLogistics 混单时排除自营运单：明细全是非代发行的票不当成代发物流。
// 无明细的票（快递助手后补的 YT 等）在自营票被排除后作为代发候选。
func collectDropshipLogistics(order *ordercore.OrderBrief, poRefItemIDs map[uint64]struct{}) []ocLogistics {
	if order == nil {
		return nil
	}
	remark := fmt.Sprintf("同步自订单 %s", strings.TrimSpace(order.OrderNo))
	toLog := func(sh ordercore.OrderShipmentBrief) ocLogistics {
		return ocLogistics{
			trackingNo: shipmentTracking(sh),
			carrier:    strings.TrimSpace(sh.ExpressCompany),
			shippedAt:  shipmentShippedAt(sh),
			remark:     remark,
		}
	}

	appendUnique := func(dst []ocLogistics, seen map[string]struct{}, sh ordercore.OrderShipmentBrief) []ocLogistics {
		tn := shipmentTracking(sh)
		if tn == "" {
			return dst
		}
		if _, ok := seen[tn]; ok {
			return dst
		}
		seen[tn] = struct{}{}
		return append(dst, toLog(sh))
	}

	seen := map[string]struct{}{}
	matched := make([]ocLogistics, 0)
	orphans := make([]ocLogistics, 0)
	others := make([]ocLogistics, 0)
	hasRef := len(poRefItemIDs) > 0

	for _, sh := range order.Shipments {
		tn := shipmentTracking(sh)
		if tn == "" {
			continue
		}
		if hasRef && shipmentHasAnyItem(sh) && !shipmentHasPOItem(sh, poRefItemIDs) {
			continue
		}
		if hasRef && shipmentHasPOItem(sh, poRefItemIDs) {
			matched = appendUnique(matched, seen, sh)
			continue
		}
		if !shipmentHasAnyItem(sh) {
			orphans = appendUnique(orphans, seen, sh)
			continue
		}
		others = appendUnique(others, seen, sh)
	}

	logs := matched
	if len(logs) == 0 {
		logs = orphans
	}
	if len(logs) == 0 {
		logs = others
	}
	if len(logs) == 0 && order.ShipStatus == "shipped" {
		logs = append(logs, ocLogistics{
			trackingNo: fmt.Sprintf("SYNC-%s", order.OrderNo),
			carrier:    "订单中心已发货",
			remark:     remark + "（无快递单号）",
		})
	}
	return logs
}
