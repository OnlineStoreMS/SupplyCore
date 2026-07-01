package dto

type SalesOrderItemInput struct {
	SkuID           uint64 `json:"skuId" binding:"required"`
	Qty             int    `json:"qty" binding:"required,min=1"`
	FulfillmentMode string `json:"fulfillmentMode"`
	Remark          string `json:"remark"`
}

type SalesOrderInput struct {
	SourceChannel   string                `json:"sourceChannel"`
	ReceiverName    string                `json:"receiverName"`
	ReceiverPhone   string                `json:"receiverPhone"`
	Province        string                `json:"province"`
	City            string                `json:"city"`
	District        string                `json:"district"`
	ReceiverAddress string                `json:"receiverAddress"`
	Remark          string                `json:"remark"`
	Items           []SalesOrderItemInput `json:"items" binding:"required,min=1,dive"`
}

type SalesOrderItemDetail struct {
	ID              uint64 `json:"id"`
	SkuID           uint64 `json:"skuId"`
	Qty             int    `json:"qty"`
	FulfillmentMode string `json:"fulfillmentMode"`
	SelectedOfferID uint64 `json:"selectedOfferId"`
	LinkedPOID      uint64 `json:"linkedPoId"`
	LinkedPoNo      string `json:"linkedPoNo,omitempty"`
	Remark          string `json:"remark"`
}

type SalesOrderDetail struct {
	ID              uint64                 `json:"id"`
	SoNo            string                 `json:"soNo"`
	TraceID         string                 `json:"traceId"`
	Status          string                 `json:"status"`
	SourceChannel   string                 `json:"sourceChannel"`
	ReceiverName    string                 `json:"receiverName"`
	ReceiverPhone   string                 `json:"receiverPhone"`
	Province        string                 `json:"province"`
	City            string                 `json:"city"`
	District        string                 `json:"district"`
	ReceiverAddress string                 `json:"receiverAddress"`
	Remark          string                 `json:"remark"`
	CreatedAt       string                 `json:"createdAt"`
	Items           []SalesOrderItemDetail `json:"items"`
}

type SalesOrderListItem struct {
	ID            uint64 `json:"id"`
	SoNo          string `json:"soNo"`
	TraceID       string `json:"traceId"`
	Status        string `json:"status"`
	SourceChannel string `json:"sourceChannel"`
	ItemCount     int    `json:"itemCount"`
	CreatedAt     string `json:"createdAt"`
}

type SourcingLinePlan struct {
	SoItemID        uint64              `json:"soItemId"`
	SkuID           uint64              `json:"skuId"`
	Qty             int                 `json:"qty"`
	FulfillmentMode string              `json:"fulfillmentMode"`
	NeedsPO         bool                `json:"needsPo"`
	Message         string              `json:"message,omitempty"`
	Offers          []SupplyOptionOffer `json:"offers,omitempty"`
	RecommendedID   uint64              `json:"recommendedOfferId,omitempty"`
}

type SourcingEvaluateResp struct {
	SoID    uint64             `json:"soId"`
	TraceID string             `json:"traceId"`
	Lines   []SourcingLinePlan `json:"lines"`
}

type SourcingSelection struct {
	SoItemID uint64 `json:"soItemId" binding:"required"`
	OfferID  uint64 `json:"offerId" binding:"required"`
	Qty      int    `json:"qty"`
}

type SourcingCreatePOInput struct {
	RefSoID     uint64              `json:"refSoId"`
	RefTraceID  string              `json:"refTraceId"`
	AutoSubmit  bool                `json:"autoSubmit"`
	Receiver    *SourcingReceiver   `json:"receiver"`
	Selections  []SourcingSelection `json:"selections" binding:"required,min=1,dive"`
}

type SourcingReceiver struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

type SourcingCreatePOResp struct {
	RefSoID    uint64   `json:"refSoId"`
	RefTraceID string   `json:"refTraceId"`
	PoIDs      []uint64 `json:"poIds"`
	PoNos      []string `json:"poNos"`
}
