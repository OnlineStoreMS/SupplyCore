package service

import (
	"testing"

	"supplycore/internal/integrations/ordercore"
)

func TestCollectDropshipLogistics_mixedSkipsSelfShipTracking(t *testing.T) {
	zto := "76985339423296"
	yt := "YT0078075632991"
	order := &ordercore.OrderBrief{
		OrderNo:    "OC202609300022",
		ShipStatus: "shipped",
		Shipments: []ordercore.OrderShipmentBrief{
			{
				ExpressCompany: "中通快递",
				ExpressNo:      zto,
				Items:          []ordercore.OrderShipmentItemBrief{{OrderItemID: 1419616, Qty: 1}},
			},
			{
				ExpressCompany: "圆通快递",
				ExpressNo:      yt,
			},
		},
	}
	poItems := map[uint64]struct{}{1419626: {}, 1419627: {}, 1419628: {}}
	logs := collectDropshipLogistics(order, poItems)
	if len(logs) != 1 {
		t.Fatalf("len=%d want 1 %#v", len(logs), logs)
	}
	if logs[0].trackingNo != yt {
		t.Fatalf("tracking=%s want %s", logs[0].trackingNo, yt)
	}
	if logs[0].carrier != "圆通快递" {
		t.Fatalf("carrier=%s", logs[0].carrier)
	}
}

func TestCollectDropshipLogistics_matchedDropshipItemsWin(t *testing.T) {
	order := &ordercore.OrderBrief{
		OrderNo: "OC1",
		Shipments: []ordercore.OrderShipmentBrief{
			{ExpressNo: "SELF", Items: []ordercore.OrderShipmentItemBrief{{OrderItemID: 1}}},
			{ExpressNo: "DS", Items: []ordercore.OrderShipmentItemBrief{{OrderItemID: 2}}},
		},
	}
	logs := collectDropshipLogistics(order, map[uint64]struct{}{2: {}})
	if len(logs) != 1 || logs[0].trackingNo != "DS" {
		t.Fatalf("%#v", logs)
	}
}

func TestCollectDropshipLogistics_pureDropshipKeepsAll(t *testing.T) {
	order := &ordercore.OrderBrief{
		OrderNo: "OC2",
		Shipments: []ordercore.OrderShipmentBrief{
			{ExpressNo: "A", Items: []ordercore.OrderShipmentItemBrief{{OrderItemID: 10}}},
		},
	}
	logs := collectDropshipLogistics(order, map[uint64]struct{}{10: {}})
	if len(logs) != 1 || logs[0].trackingNo != "A" {
		t.Fatalf("%#v", logs)
	}
}
