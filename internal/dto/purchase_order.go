package dto

type PurchaseOrderItemInput struct {
	SkuID           uint64  `json:"skuId"`
	OfferID         uint64  `json:"offerId"`
	ProductName     string  `json:"productName"`
	SkuCode         string  `json:"skuCode"`
	SkuSpecs        string  `json:"skuSpecs"`
	PicURL          string  `json:"picUrl"`
	SupplierSkuCode string  `json:"supplierSkuCode"`
	Qty             int     `json:"qty" binding:"required,min=1"`
	SaleUnitPrice   float64 `json:"saleUnitPrice"`
	SaleAmount      float64 `json:"saleAmount"`
	UnitPrice       float64 `json:"unitPrice"`
	Remark          string  `json:"remark"`
}

type PurchaseOrderInput struct {
	SupplierID          uint64                   `json:"supplierId" binding:"required"`
	FulfillmentType     string                   `json:"fulfillmentType"`
	Currency            string                   `json:"currency"`
	ExpectedArrivalDate string                   `json:"expectedArrivalDate"`
	WarehouseID         uint64                   `json:"warehouseId"`
	RefSoID             uint64                   `json:"refSoId"`
	RefTraceID          string                   `json:"refTraceId"`
	OrderedAt           string                   `json:"orderedAt"` // 下单时间，手工新建可改
	SaleAmount          float64                  `json:"saleAmount"` // 销售侧订单总实付
	Remark              string                   `json:"remark"`
	Items               []PurchaseOrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type PurchaseOrderItemDetail struct {
	ID              uint64  `json:"id"`
	SkuID           uint64  `json:"skuId"`
	OfferID         uint64  `json:"offerId"`
	ProductName     string  `json:"productName"`
	SkuCode         string  `json:"skuCode"`
	SkuSpecs        string  `json:"skuSpecs"`
	PicURL          string  `json:"picUrl"`
	SupplierSkuCode string  `json:"supplierSkuCode"`
	Qty             int     `json:"qty"`
	SaleUnitPrice   float64 `json:"saleUnitPrice"`
	SaleAmount      float64 `json:"saleAmount"`
	UnitPrice       float64 `json:"unitPrice"`
	LineAmount      float64 `json:"lineAmount"`
	ReceivedQty     int     `json:"receivedQty"`
	Remark          string  `json:"remark"`
}

type PurchaseOrderDetail struct {
	ID                  uint64                    `json:"id"`
	PoNo                string                    `json:"poNo"`
	SupplierID          uint64                    `json:"supplierId"`
	SupplierName        string                    `json:"supplierName"`
	SupplierCode        string                    `json:"supplierCode"`
	Status              string                    `json:"status"`
	TotalAmount         float64                   `json:"totalAmount"`
	SaleAmount          float64                   `json:"saleAmount"`
	Currency            string                    `json:"currency"`
	ExpectedArrivalDate string                    `json:"expectedArrivalDate,omitempty"`
	WarehouseID         uint64                    `json:"warehouseId"`
	FulfillmentType     string                    `json:"fulfillmentType"`
	RefSoID             uint64                    `json:"refSoId"`
	RefTraceID          string                    `json:"refTraceId"`
	BuyerID             uint64                    `json:"buyerId"`
	BuyerName           string                    `json:"buyerName"`
	PayStatus           string                    `json:"payStatus"`
	Remark              string                    `json:"remark"`
	OrderedAt           string                    `json:"orderedAt,omitempty"`
	CompletedAt         string                    `json:"completedAt,omitempty"`
	CreatedAt           string                    `json:"createdAt"`
	Items               []PurchaseOrderItemDetail `json:"items"`
}

type PurchaseOrderListItem struct {
	ID              uint64  `json:"id"`
	PoNo            string  `json:"poNo"`
	SupplierID      uint64  `json:"supplierId"`
	SupplierName    string  `json:"supplierName"`
	Status          string  `json:"status"`
	PayStatus       string  `json:"payStatus"`
	FulfillmentType string  `json:"fulfillmentType"`
	TotalAmount     float64 `json:"totalAmount"`
	SaleAmount      float64 `json:"saleAmount"`
	Currency        string  `json:"currency"`
	ItemCount       int     `json:"itemCount"`
	RefSoID         uint64  `json:"refSoId,omitempty"`
	RefTraceID      string  `json:"refTraceId,omitempty"`
	OrderedAt       string  `json:"orderedAt,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}
