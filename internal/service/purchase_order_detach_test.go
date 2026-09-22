package service

import (
	"strings"
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

func TestPendingManualUnbindRemark(t *testing.T) {
	if got := pendingManualUnbindRemark("退款完成"); got != "销售单 退款成功。待人工解绑" {
		t.Fatalf("refund remark: %q", got)
	}
	if got := pendingManualUnbindRemark("交易关闭"); got != "销售单 交易关闭。待人工解绑" {
		t.Fatalf("close remark: %q", got)
	}
}

func TestStripPendingManualUnbindRemark(t *testing.T) {
	in := "销售单 退款成功。待人工解绑（OC202609200031） OMS单号：OC202609200031"
	got := stripPendingManualUnbindRemark(in)
	if strings.Contains(got, "待人工解绑") {
		t.Fatalf("still has pending mark: %q", got)
	}
	if !strings.Contains(got, "OMS单号") {
		t.Fatalf("lost original remark: %q", got)
	}
}
