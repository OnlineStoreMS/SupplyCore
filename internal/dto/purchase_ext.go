package dto

type PurchaseAccountDTO struct {
	ID           uint64  `json:"id,omitempty"`
	Channel      string  `json:"channel"`
	AccountAlias string  `json:"accountAlias"`
	AccountName  string  `json:"accountName"`
	IsPrimary    bool    `json:"isPrimary"`
	Status       string  `json:"status"`
	AuthStatus   string  `json:"authStatus"`
	LastSyncAt   *string `json:"lastSyncAt,omitempty"`
	OperatorName string  `json:"operatorName"`
	Remark       string  `json:"remark"`
	CreatedAt    string  `json:"createdAt,omitempty"`
	UpdatedAt    string  `json:"updatedAt,omitempty"`
}

type PurchaseInboundItemInput struct {
	POItemID     uint64  `json:"poItemId"`
	SkuID        uint64  `json:"skuId"`
	SkuCode      string  `json:"skuCode"`
	SkuName      string  `json:"skuName"`
	PurchaseQty  int     `json:"purchaseQty"`
	QCQty        int     `json:"qcQty"`
	RejectQty    int     `json:"rejectQty"`
	InboundQty   int     `json:"inboundQty"`
	UnitPrice    float64 `json:"unitPrice"`
	LocationCode string  `json:"locationCode"`
	Remark       string  `json:"remark"`
}

type PurchaseInboundInput struct {
	POID            uint64                     `json:"poId"`
	SupplierID      uint64                     `json:"supplierId"`
	WarehouseID     uint64                     `json:"warehouseId"`
	WarehouseName   string                     `json:"warehouseName"`
	TrackingNo      string                     `json:"trackingNo"`
	PlatformOrderNo string                     `json:"platformOrderNo"`
	BuyerName       string                     `json:"buyerName"`
	Remark          string                     `json:"remark"`
	Items           []PurchaseInboundItemInput `json:"items"`
}

type PurchaseInboundItemDTO struct {
	ID           uint64  `json:"id"`
	POItemID     uint64  `json:"poItemId"`
	SkuID        uint64  `json:"skuId"`
	SkuCode      string  `json:"skuCode"`
	SkuName      string  `json:"skuName"`
	PurchaseQty  int     `json:"purchaseQty"`
	QCQty        int     `json:"qcQty"`
	RejectQty    int     `json:"rejectQty"`
	InboundQty   int     `json:"inboundQty"`
	UnitPrice    float64 `json:"unitPrice"`
	LineAmount   float64 `json:"lineAmount"`
	LocationCode string  `json:"locationCode"`
	Remark       string  `json:"remark"`
}

type PurchaseInboundListItem struct {
	ID            uint64  `json:"id"`
	InboundNo     string  `json:"inboundNo"`
	Status        string  `json:"status"`
	POID          uint64  `json:"poId"`
	PoNo          string  `json:"poNo"`
	SupplierID    uint64  `json:"supplierId"`
	SupplierName  string  `json:"supplierName"`
	WarehouseID   uint64  `json:"warehouseId"`
	WarehouseName string  `json:"warehouseName"`
	TotalQty      int     `json:"totalQty"`
	TotalAmount   float64 `json:"totalAmount"`
	TrackingNo    string  `json:"trackingNo"`
	CreatorName   string  `json:"creatorName"`
	CreatedAt     string  `json:"createdAt"`
}

type PurchaseInboundDetail struct {
	PurchaseInboundListItem
	PlatformOrderNo string                   `json:"platformOrderNo"`
	BuyerName       string                   `json:"buyerName"`
	Remark          string                   `json:"remark"`
	WHAuditorName   string                   `json:"whAuditorName"`
	WHAuditedAt     string                   `json:"whAuditedAt,omitempty"`
	FinAuditorName  string                   `json:"finAuditorName"`
	FinAuditedAt    string                   `json:"finAuditedAt,omitempty"`
	Items           []PurchaseInboundItemDTO `json:"items"`
}

type PackageReceiveInput struct {
	WarehouseID   uint64 `json:"warehouseId"`
	WarehouseName string `json:"warehouseName"`
	Carrier       string `json:"carrier"`
	TrackingNo    string `json:"trackingNo"`
	PackageType   string `json:"packageType"`
	POID          uint64 `json:"poId"`
	PoNo          string `json:"poNo"`
	Remark        string `json:"remark"`
}

type PackageReceiveDTO struct {
	ID            uint64 `json:"id"`
	WarehouseID   uint64 `json:"warehouseId"`
	WarehouseName string `json:"warehouseName"`
	Carrier       string `json:"carrier"`
	TrackingNo    string `json:"trackingNo"`
	PackageType   string `json:"packageType"`
	POID          uint64 `json:"poId"`
	PoNo          string `json:"poNo"`
	InboundID     uint64 `json:"inboundId"`
	InboundNo     string `json:"inboundNo"`
	ScannerName   string `json:"scannerName"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"createdAt"`
}

type PurchaseReturnItemInput struct {
	InboundItemID uint64  `json:"inboundItemId"`
	SkuID         uint64  `json:"skuId"`
	SkuCode       string  `json:"skuCode"`
	SkuName       string  `json:"skuName"`
	OriginalQty   int     `json:"originalQty"`
	ReturnQty     int     `json:"returnQty"`
	UnitPrice     float64 `json:"unitPrice"`
	ActualAmount  float64 `json:"actualAmount"`
	Remark        string  `json:"remark"`
}

type PurchaseReturnInput struct {
	InboundID     uint64                    `json:"inboundId"`
	SupplierID    uint64                    `json:"supplierId"`
	WarehouseID   uint64                    `json:"warehouseId"`
	WarehouseName string                    `json:"warehouseName"`
	TrackingNo    string                    `json:"trackingNo"`
	BuyerName     string                    `json:"buyerName"`
	Remark        string                    `json:"remark"`
	Items         []PurchaseReturnItemInput `json:"items"`
}

type PurchaseReturnItemDTO struct {
	ID            uint64  `json:"id"`
	InboundItemID uint64  `json:"inboundItemId"`
	SkuID         uint64  `json:"skuId"`
	SkuCode       string  `json:"skuCode"`
	SkuName       string  `json:"skuName"`
	OriginalQty   int     `json:"originalQty"`
	ReturnQty     int     `json:"returnQty"`
	UnitPrice     float64 `json:"unitPrice"`
	ReturnAmount  float64 `json:"returnAmount"`
	ActualAmount  float64 `json:"actualAmount"`
	Remark        string  `json:"remark"`
}

type PurchaseReturnListItem struct {
	ID            uint64  `json:"id"`
	ReturnNo      string  `json:"returnNo"`
	Status        string  `json:"status"`
	InboundID     uint64  `json:"inboundId"`
	InboundNo     string  `json:"inboundNo"`
	SupplierID    uint64  `json:"supplierId"`
	SupplierName  string  `json:"supplierName"`
	WarehouseName string  `json:"warehouseName"`
	TotalQty      int     `json:"totalQty"`
	TotalAmount   float64 `json:"totalAmount"`
	ActualAmount  float64 `json:"actualAmount"`
	CreatorName   string  `json:"creatorName"`
	CreatedAt     string  `json:"createdAt"`
}

type PurchaseReturnDetail struct {
	PurchaseReturnListItem
	TrackingNo     string                  `json:"trackingNo"`
	BuyerName      string                  `json:"buyerName"`
	Remark         string                  `json:"remark"`
	AuditorName    string                  `json:"auditorName"`
	AuditedAt      string                  `json:"auditedAt,omitempty"`
	FinAuditorName string                  `json:"finAuditorName"`
	FinAuditedAt   string                  `json:"finAuditedAt,omitempty"`
	Items          []PurchaseReturnItemDTO `json:"items"`
}

type StockoutSuggestionItem struct {
	SkuID           uint64  `json:"skuId"`
	SkuCode         string  `json:"skuCode"`
	SkuName         string  `json:"skuName"`
	SupplierID      uint64  `json:"supplierId"`
	SupplierName    string  `json:"supplierName"`
	StockoutQty     int     `json:"stockoutQty"`
	SalesQty        int     `json:"salesQty"`
	SuggestPurchase int     `json:"suggestPurchase"`
	UnitPrice       float64 `json:"unitPrice"`
	OfferID         uint64  `json:"offerId"`
	Source          string  `json:"source"` // stockout | warning | no_stock
}

type GeneratePOFromSuggestionInput struct {
	SupplierID uint64   `json:"supplierId"`
	SkuIDs     []uint64 `json:"skuIds"`
	Source     string   `json:"source"`
}
