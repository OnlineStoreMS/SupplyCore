package dto

type PurchaseOrderItemInput struct {
	SkuID           uint64  `json:"skuId" binding:"required"`
	OfferID         uint64  `json:"offerId"`
	SupplierSkuCode string  `json:"supplierSkuCode"`
	Qty             int     `json:"qty" binding:"required,min=1"`
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
	Remark              string                   `json:"remark"`
	Items               []PurchaseOrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type PurchaseOrderItemDetail struct {
	ID              uint64  `json:"id"`
	SkuID           uint64  `json:"skuId"`
	OfferID         uint64  `json:"offerId"`
	SupplierSkuCode string  `json:"supplierSkuCode"`
	Qty             int     `json:"qty"`
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
	Currency        string  `json:"currency"`
	ItemCount       int     `json:"itemCount"`
	RefSoID         uint64  `json:"refSoId,omitempty"`
	RefTraceID      string  `json:"refTraceId,omitempty"`
	OrderedAt       string  `json:"orderedAt,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}
