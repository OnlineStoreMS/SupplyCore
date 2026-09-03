package service

import (
	"testing"

	"supplycore/internal/model"
)

func TestShouldCancelWholePO(t *testing.T) {
	po := &model.PurchaseOrder{Status: model.POStatusCompleted}
	if shouldCancelWholePO(po, 0, "退款完成") {
		t.Fatal("completed PO must not cancel")
	}
	po.Status = model.POStatusShipped
	if !shouldCancelWholePO(po, 0, "退款完成") {
		t.Fatal("refund close with no active lines should cancel shipped PO")
	}
	if shouldCancelWholePO(po, 1, "退款完成") {
		t.Fatal("active lines should block cancel")
	}
	po.Status = model.POStatusOrdered
	if !shouldCancelWholePO(po, 0, "手动解绑") {
		t.Fatal("draft/ordered with no active lines should cancel")
	}
}

func TestIsRefundCloseDetachReason(t *testing.T) {
	if !isRefundCloseDetachReason("退款完成") {
		t.Fatal("expected refund reason")
	}
	if isRefundCloseDetachReason("手动解绑") {
		t.Fatal("manual reason should not match refund")
	}
}
