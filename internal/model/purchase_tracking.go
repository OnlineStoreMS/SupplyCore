package model

import "time"

type PurchaseShipment struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	TenantID            uint64     `gorm:"index;not null" json:"tenantId"`
	POID                uint64     `gorm:"index;not null" json:"poId"`
	ShipmentNo          string     `gorm:"size:32;not null" json:"shipmentNo"`
	Status              string     `gorm:"size:32;not null;default:pending" json:"status"`
	CarrierCode         string     `gorm:"size:32" json:"carrierCode"`
	CarrierName         string     `gorm:"size:64" json:"carrierName"`
	TrackingNo          string     `gorm:"size:64" json:"trackingNo"`
	ShippedAt           *time.Time `json:"shippedAt"`
	ExpectedArrivalDate *time.Time `json:"expectedArrivalDate"`
	DeliveredAt         *time.Time `json:"deliveredAt"`
	ShipFromAddressID   uint64     `json:"shipFromAddressId"`
	ReceiverName        string     `gorm:"size:64" json:"receiverName"`
	ReceiverPhone       string     `gorm:"size:32" json:"receiverPhone"`
	ReceiverAddress     string     `gorm:"size:255" json:"receiverAddress"`
	Remark              string     `gorm:"type:text" json:"remark"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	Items               []PurchaseShipmentItem `gorm:"foreignKey:ShipmentID" json:"items,omitempty"`
}

func (PurchaseShipment) TableName() string { return "purchase_shipments" }

type PurchaseShipmentItem struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	TenantID   uint64    `gorm:"index;not null" json:"tenantId"`
	ShipmentID uint64    `gorm:"index;not null" json:"shipmentId"`
	POItemID   uint64    `gorm:"index;not null" json:"poItemId"`
	Qty        int       `gorm:"not null" json:"qty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (PurchaseShipmentItem) TableName() string { return "purchase_shipment_items" }

type PurchasePayment struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	TenantID      uint64     `gorm:"index;not null" json:"tenantId"`
	POID          uint64     `gorm:"index;not null" json:"poId"`
	PayAmount     float64    `gorm:"type:decimal(14,2);not null" json:"payAmount"`
	PayMethod     string     `gorm:"size:32" json:"payMethod"`
	PayAccount    string     `gorm:"size:128" json:"payAccount"`
	PayeeAccount  string     `gorm:"size:128" json:"payeeAccount"`
	PayeeName     string     `gorm:"size:128" json:"payeeName"`
	PayStatus     string     `gorm:"size:16;default:paid" json:"payStatus"`
	PaidAt        *time.Time `json:"paidAt"`
	Remark        string     `gorm:"type:text" json:"remark"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (PurchasePayment) TableName() string { return "purchase_payments" }

type PurchaseAttachment struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	TenantID   uint64    `gorm:"index;not null" json:"tenantId"`
	POID       uint64    `gorm:"index;not null" json:"poId"`
	PaymentID  uint64    `json:"paymentId"`
	FileType   string    `gorm:"size:32;not null" json:"fileType"`
	FileName   string    `gorm:"size:255;not null" json:"fileName"`
	FileURL    string    `gorm:"size:512;not null" json:"fileUrl"`
	UploadedBy uint64    `json:"uploadedBy"`
	Remark     string    `gorm:"type:text" json:"remark"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (PurchaseAttachment) TableName() string { return "purchase_attachments" }
