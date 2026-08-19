package service

import (
	"testing"

	"supplycore/internal/model"
)

func TestDropshipPOHasSalesOrderRef(t *testing.T) {
	tests := []struct {
		name string
		po   *model.PurchaseOrder
		want bool
	}{
		{name: "nil", po: nil, want: false},
		{name: "empty", po: &model.PurchaseOrder{}, want: false},
		{name: "ref_so_id", po: &model.PurchaseOrder{RefSoID: 3184}, want: true},
		{name: "ref_trace_id", po: &model.PurchaseOrder{RefTraceID: "OC202608180019"}, want: true},
		{name: "trace_whitespace", po: &model.PurchaseOrder{RefTraceID: "  "}, want: false},
		{name: "manual_dropship", po: &model.PurchaseOrder{RefSoID: 0, RefTraceID: ""}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dropshipPOHasSalesOrderRef(tt.po); got != tt.want {
				t.Fatalf("dropshipPOHasSalesOrderRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
