package model

import "time"

const (
	InboundStatusDraft          = "draft"
	InboundStatusPendingWH      = "pending_wh"
	InboundStatusPendingFinance = "pending_finance"
	InboundStatusCompleted      = "completed"
	InboundStatusVoid           = "void"

	ReturnStatusDraft          = "draft"
	ReturnStatusPendingReturn  = "pending_return"
	ReturnStatusPendingFinance = "pending_finance"
	ReturnStatusCompleted      = "completed"
	ReturnStatusVoid           = "void"

	PurchaseAccountStatusActive   = "active"
	PurchaseAccountStatusDisabled = "disabled"
	PurchaseAccountStatusRevoked  = "revoked"
)

// PurchaseAccount 采购账号（对应普源：1688/淘供销等采购渠道账号）
type PurchaseAccount struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	TenantID     uint64     `gorm:"index;not null" json:"tenantId"`
	Channel      string     `gorm:"size:32;not null" json:"channel"` // alibaba1688 | taogongxiao | other
	AccountAlias string     `gorm:"size:64;not null" json:"accountAlias"`
	AccountName  string     `gorm:"size:128;not null" json:"accountName"`
	IsPrimary    bool       `gorm:"not null;default:false" json:"isPrimary"`
	Status       string     `gorm:"size:16;not null;default:active" json:"status"`
	AuthStatus   string     `gorm:"size:32;default:unauthorized" json:"authStatus"` // unauthorized|authorized|expired
	LastSyncAt   *time.Time `json:"lastSyncAt"`
	OperatorName string     `gorm:"size:64" json:"operatorName"`
	Remark       string     `gorm:"type:text" json:"remark"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (PurchaseAccount) TableName() string { return "purchase_accounts" }

// PurchaseInbound 采购入库单
type PurchaseInbound struct {
	ID              uint64                `gorm:"primaryKey" json:"id"`
	TenantID        uint64                `gorm:"index;not null" json:"tenantId"`
	InboundNo       string                `gorm:"size:32;not null" json:"inboundNo"`
	Status          string                `gorm:"size:32;not null;default:draft" json:"status"`
	POID            uint64                `gorm:"index" json:"poId"`
	PoNo            string                `gorm:"size:32" json:"poNo"`
	SupplierID      uint64                `gorm:"index" json:"supplierId"`
	WarehouseID     uint64                `json:"warehouseId"`
	WarehouseName   string                `gorm:"size:64" json:"warehouseName"`
	TrackingNo      string                `gorm:"size:64" json:"trackingNo"`
	PlatformOrderNo string                `gorm:"size:64" json:"platformOrderNo"`
	TotalQty        int                   `gorm:"default:0" json:"totalQty"`
	TotalAmount     float64               `gorm:"type:decimal(14,2);default:0" json:"totalAmount"`
	CreatorName     string                `gorm:"size:64" json:"creatorName"`
	WHAuditorName   string                `gorm:"size:64" json:"whAuditorName"`
	WHAuditedAt     *time.Time            `json:"whAuditedAt"`
	FinAuditorName  string                `gorm:"size:64" json:"finAuditorName"`
	FinAuditedAt    *time.Time            `json:"finAuditedAt"`
	BuyerName       string                `gorm:"size:64" json:"buyerName"`
	Remark          string                `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
	Items           []PurchaseInboundItem `gorm:"foreignKey:InboundID" json:"items,omitempty"`
}

func (PurchaseInbound) TableName() string { return "purchase_inbounds" }

type PurchaseInboundItem struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenantId"`
	InboundID     uint64    `gorm:"index;not null" json:"inboundId"`
	POItemID      uint64    `json:"poItemId"`
	SkuID         uint64    `gorm:"index;not null" json:"skuId"`
	SkuCode       string    `gorm:"size:64" json:"skuCode"`
	SkuName       string    `gorm:"size:255" json:"skuName"`
	PurchaseQty   int       `gorm:"default:0" json:"purchaseQty"`
	QCQty         int       `gorm:"default:0" json:"qcQty"`
	RejectQty     int       `gorm:"default:0" json:"rejectQty"`
	InboundQty    int       `gorm:"not null" json:"inboundQty"`
	UnitPrice     float64   `gorm:"type:decimal(12,2);default:0" json:"unitPrice"`
	LineAmount    float64   `gorm:"type:decimal(14,2);default:0" json:"lineAmount"`
	LocationCode  string    `gorm:"size:64" json:"locationCode"`
	Remark        string    `gorm:"type:text" json:"remark"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (PurchaseInboundItem) TableName() string { return "purchase_inbound_items" }

// PackageReceiveRecord 收货记录（包裹扫描）
type PackageReceiveRecord struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TenantID       uint64    `gorm:"index;not null" json:"tenantId"`
	WarehouseID    uint64    `json:"warehouseId"`
	WarehouseName  string    `gorm:"size:64" json:"warehouseName"`
	Carrier        string    `gorm:"size:64" json:"carrier"`
	TrackingNo     string    `gorm:"size:64;index;not null" json:"trackingNo"`
	PackageType    string    `gorm:"size:32;default:normal" json:"packageType"`
	POID           uint64    `gorm:"index" json:"poId"`
	PoNo           string    `gorm:"size:32" json:"poNo"`
	InboundID      uint64    `json:"inboundId"`
	InboundNo      string    `gorm:"size:32" json:"inboundNo"`
	ScannerName    string    `gorm:"size:64" json:"scannerName"`
	Remark         string    `gorm:"type:text" json:"remark"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (PackageReceiveRecord) TableName() string { return "package_receive_records" }

// PurchaseReturn 采购退回单
type PurchaseReturn struct {
	ID              uint64               `gorm:"primaryKey" json:"id"`
	TenantID        uint64               `gorm:"index;not null" json:"tenantId"`
	ReturnNo        string               `gorm:"size:32;not null" json:"returnNo"`
	Status          string               `gorm:"size:32;not null;default:draft" json:"status"`
	InboundID       uint64               `gorm:"index" json:"inboundId"`
	InboundNo       string               `gorm:"size:32" json:"inboundNo"`
	SupplierID      uint64               `gorm:"index" json:"supplierId"`
	WarehouseID     uint64               `json:"warehouseId"`
	WarehouseName   string               `gorm:"size:64" json:"warehouseName"`
	TrackingNo      string               `gorm:"size:64" json:"trackingNo"`
	TotalQty        int                  `gorm:"default:0" json:"totalQty"`
	TotalAmount     float64              `gorm:"type:decimal(14,2);default:0" json:"totalAmount"`
	ActualAmount    float64              `gorm:"type:decimal(14,2);default:0" json:"actualAmount"`
	CreatorName     string               `gorm:"size:64" json:"creatorName"`
	AuditorName     string               `gorm:"size:64" json:"auditorName"`
	AuditedAt       *time.Time           `json:"auditedAt"`
	FinAuditorName  string               `gorm:"size:64" json:"finAuditorName"`
	FinAuditedAt    *time.Time           `json:"finAuditedAt"`
	BuyerName       string               `gorm:"size:64" json:"buyerName"`
	Remark          string               `gorm:"type:text" json:"remark"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	Items           []PurchaseReturnItem `gorm:"foreignKey:ReturnID" json:"items,omitempty"`
}

func (PurchaseReturn) TableName() string { return "purchase_returns" }

type PurchaseReturnItem struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TenantID       uint64    `gorm:"index;not null" json:"tenantId"`
	ReturnID       uint64    `gorm:"index;not null" json:"returnId"`
	InboundItemID  uint64    `json:"inboundItemId"`
	SkuID          uint64    `gorm:"index;not null" json:"skuId"`
	SkuCode        string    `gorm:"size:64" json:"skuCode"`
	SkuName        string    `gorm:"size:255" json:"skuName"`
	OriginalQty    int       `gorm:"default:0" json:"originalQty"`
	ReturnQty      int       `gorm:"not null" json:"returnQty"`
	UnitPrice      float64   `gorm:"type:decimal(12,2);default:0" json:"unitPrice"`
	ReturnAmount   float64   `gorm:"type:decimal(14,2);default:0" json:"returnAmount"`
	ActualAmount   float64   `gorm:"type:decimal(14,2);default:0" json:"actualAmount"`
	Remark         string    `gorm:"type:text" json:"remark"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (PurchaseReturnItem) TableName() string { return "purchase_return_items" }
