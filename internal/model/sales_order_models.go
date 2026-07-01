package model

import "time"

type SalesOrder struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	TenantID        uint64    `gorm:"index;not null" json:"tenantId"`
	SoNo            string    `gorm:"size:32;not null" json:"soNo"`
	TraceID         string    `gorm:"size:64;not null" json:"traceId"`
	Status          string    `gorm:"size:32;not null;default:draft" json:"status"`
	SourceChannel   string    `gorm:"size:32;default:manual" json:"sourceChannel"`
	ReceiverName    string    `gorm:"size:64" json:"receiverName"`
	ReceiverPhone   string    `gorm:"size:32" json:"receiverPhone"`
	Province        string    `gorm:"size:32" json:"province"`
	City            string    `gorm:"size:32" json:"city"`
	District        string    `gorm:"size:32" json:"district"`
	ReceiverAddress string    `gorm:"size:255" json:"receiverAddress"`
	Remark          string    `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Items           []SalesOrderItem `gorm:"foreignKey:SOID" json:"items,omitempty"`
}

func (SalesOrder) TableName() string { return "sales_orders" }

type SalesOrderItem struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	TenantID        uint64    `gorm:"index;not null" json:"tenantId"`
	SOID            uint64    `gorm:"index;not null" json:"soId"`
	SkuID           uint64    `gorm:"index;not null" json:"skuId"`
	Qty             int       `gorm:"not null" json:"qty"`
	FulfillmentMode string    `gorm:"size:16;default:dropship" json:"fulfillmentMode"`
	SelectedOfferID uint64    `json:"selectedOfferId"`
	LinkedPOID      uint64    `json:"linkedPoId"`
	Remark          string    `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (SalesOrderItem) TableName() string { return "sales_order_items" }
