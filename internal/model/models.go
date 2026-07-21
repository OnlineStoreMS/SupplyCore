package model

import "time"

type SupplierCategory struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	TenantID  uint64    `gorm:"index;not null" json:"tenantId"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	ParentID  uint64    `gorm:"index;default:0" json:"parentId"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Status    int8      `gorm:"default:1;not null" json:"status"`
	Remark    string    `gorm:"type:text" json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (SupplierCategory) TableName() string { return "supplier_categories" }

type Supplier struct {
	ID                  uint64    `gorm:"primaryKey" json:"id"`
	TenantID            uint64    `gorm:"index;not null" json:"tenantId"`
	CategoryID          uint64    `gorm:"index" json:"categoryId"`
	CategoryName        string    `gorm:"size:64" json:"categoryName"`
	Code                string    `gorm:"size:64;not null" json:"code"`
	Name                string    `gorm:"size:128;not null" json:"name"`
	ShortName           string    `gorm:"size:64" json:"shortName"`
	Status              int8      `gorm:"default:1;not null" json:"status"`
	BuyerName           string    `gorm:"size:64" json:"buyerName"`
	CutOffTime          string    `gorm:"size:16;default:00:01" json:"cutOffTime"`
	ArrivalDays         int       `json:"arrivalDays"`
	PaymentDays         int       `json:"paymentDays"`
	ContactName         string    `gorm:"size:64" json:"contactName"`
	Address             string    `gorm:"size:255" json:"address"`
	OfficePhone         string    `gorm:"size:32" json:"officePhone"`
	Mobile              string    `gorm:"size:32" json:"mobile"`
	Phone               string    `gorm:"size:32" json:"phone"`
	WangwangID          string    `gorm:"size:64" json:"wangwangId"`
	QQ                  string    `gorm:"size:32" json:"qq"`
	Email               string    `gorm:"size:128" json:"email"`
	Website             string    `gorm:"size:255" json:"website"`
	Remark              string    `gorm:"type:text" json:"remark"`
	DefaultPaymentTerms string    `gorm:"size:255" json:"defaultPaymentTerms"`
	BankName            string    `gorm:"size:128" json:"bankName"`
	BankAccount         string    `gorm:"size:64" json:"bankAccount"`
	AccountName         string    `gorm:"size:128" json:"accountName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

func (Supplier) TableName() string { return "suppliers" }

type SupplierAddress struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	TenantID     uint64    `gorm:"index;not null" json:"tenantId"`
	SupplierID   uint64    `gorm:"index;not null" json:"supplierId"`
	AddressType  string    `gorm:"size:16;not null;default:ship;index" json:"addressType"` // ship / return
	Label        string    `gorm:"size:64;not null" json:"label"`
	ContactName  string    `gorm:"size:64" json:"contactName"`
	Phone        string    `gorm:"size:32" json:"phone"`
	Province     string    `gorm:"size:32" json:"province"`
	City         string    `gorm:"size:32" json:"city"`
	District     string    `gorm:"size:32" json:"district"`
	Address      string    `gorm:"size:255" json:"address"`
	IsDefault    bool      `gorm:"not null;default:false" json:"isDefault"`
	Status       int8      `gorm:"default:1;not null" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (SupplierAddress) TableName() string { return "supplier_addresses" }

// SupplierPaymentAccount 供应商收款账户（银行/支付宝/微信等）
type SupplierPaymentAccount struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	SupplierID  uint64    `gorm:"index;not null" json:"supplierId"`
	Label       string    `gorm:"size:64;not null" json:"label"`
	AccountType string    `gorm:"size:32;not null;default:bank" json:"accountType"` // bank / alipay / wechat / other
	BankName    string    `gorm:"size:128" json:"bankName"`
	BankAccount string    `gorm:"size:64" json:"bankAccount"`
	AccountName string    `gorm:"size:128" json:"accountName"`
	IsDefault   bool      `gorm:"not null;default:false" json:"isDefault"`
	Status      int8      `gorm:"default:1;not null" json:"status"`
	Remark      string    `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (SupplierPaymentAccount) TableName() string { return "supplier_payment_accounts" }

// SupplierPaymentQR 供应商收款码
type SupplierPaymentQR struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	TenantID    uint64    `gorm:"index;not null" json:"tenantId"`
	SupplierID  uint64    `gorm:"index;not null" json:"supplierId"`
	Label       string    `gorm:"size:64;not null" json:"label"`
	PayType     string    `gorm:"size:32;not null;default:wechat" json:"payType"` // wechat / alipay / other
	ImageURL    string    `gorm:"size:512;not null" json:"imageUrl"`
	AccountName string    `gorm:"size:128" json:"accountName"`
	IsDefault   bool      `gorm:"not null;default:false" json:"isDefault"`
	Status      int8      `gorm:"default:1;not null" json:"status"`
	Remark      string    `gorm:"size:255" json:"remark"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (SupplierPaymentQR) TableName() string { return "supplier_payment_qrs" }

type SkuSupplierOffer struct {
	ID                 uint64    `gorm:"primaryKey" json:"id"`
	TenantID           uint64    `gorm:"index;not null" json:"tenantId"`
	SkuID              uint64    `gorm:"index;not null" json:"skuId"`
	SupplierID         uint64    `gorm:"index;not null" json:"supplierId"`
	SupplierSkuCode    string    `gorm:"size:64" json:"supplierSkuCode"`
	SupplyPrice        float64   `gorm:"type:decimal(12,2);not null" json:"supplyPrice"`
	Currency           string    `gorm:"size:8;default:CNY" json:"currency"`
	MinOrderQty        int       `gorm:"default:1" json:"minOrderQty"`
	LeadTimeDays       int       `json:"leadTimeDays"`
	ShipFromAddressID  uint64    `gorm:"index" json:"shipFromAddressId"`
	SupportsDropship   bool      `gorm:"not null;default:false" json:"supportsDropship"`
	SupportsSelfStock  bool      `gorm:"not null;default:true" json:"supportsSelfStock"`
	IsPrimary          bool      `gorm:"not null;default:false" json:"isPrimary"`
	Priority           int       `gorm:"default:0" json:"priority"`
	Status             int8      `gorm:"default:1;not null" json:"status"`
	Remark             string    `gorm:"type:text" json:"remark"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (SkuSupplierOffer) TableName() string { return "sku_supplier_offers" }

type PurchaseOrder struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	TenantID            uint64     `gorm:"index;not null" json:"tenantId"`
	PoNo                string     `gorm:"size:32;not null" json:"poNo"`
	SupplierID          uint64     `gorm:"index;not null" json:"supplierId"`
	Status              string     `gorm:"size:32;not null;default:draft" json:"status"`
	TotalAmount         float64    `gorm:"type:decimal(14,2);not null;default:0" json:"totalAmount"`
	Currency            string     `gorm:"size:8;default:CNY" json:"currency"`
	ExpectedArrivalDate *time.Time `json:"expectedArrivalDate"`
	WarehouseID         uint64     `json:"warehouseId"`
	FulfillmentType     string     `gorm:"size:16;default:stock_in" json:"fulfillmentType"`
	RefTraceID          string     `gorm:"size:64" json:"refTraceId"`
	RefSoID             uint64     `json:"refSoId"`
	BuyerID             uint64     `json:"buyerId"`
	BuyerName           string     `gorm:"size:64" json:"buyerName"`
	PayStatus           string     `gorm:"size:16;default:unpaid" json:"payStatus"`
	Remark              string     `gorm:"type:text" json:"remark"`
	OrderedAt           *time.Time `json:"orderedAt"`
	CompletedAt         *time.Time `json:"completedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	Items               []PurchaseOrderItem `gorm:"foreignKey:POID" json:"items,omitempty"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

type PurchaseOrderItem struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	TenantID            uint64     `gorm:"index;not null" json:"tenantId"`
	POID                uint64     `gorm:"index;not null" json:"poId"`
	SkuID               uint64     `gorm:"index;not null;default:0" json:"skuId"`
	OfferID             uint64     `json:"offerId"`
	ProductName         string     `gorm:"size:512" json:"productName"`
	SupplierSkuCode     string     `gorm:"size:64" json:"supplierSkuCode"`
	Qty                 int        `gorm:"not null" json:"qty"`
	UnitPrice           float64    `gorm:"type:decimal(12,2);not null" json:"unitPrice"`
	LineAmount          float64    `gorm:"type:decimal(14,2);not null" json:"lineAmount"`
	ExpectedArrivalDate *time.Time `json:"expectedArrivalDate"`
	ReceivedQty         int        `gorm:"default:0" json:"receivedQty"`
	Remark              string     `gorm:"type:text" json:"remark"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func (PurchaseOrderItem) TableName() string { return "purchase_order_items" }
