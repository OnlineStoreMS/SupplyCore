package model

const (
	POStatusDraft           = "draft"
	POStatusOrdered         = "ordered"
	POStatusPaid            = "paid"
	POStatusPartialShipped  = "partial_shipped" // 部分发货
	POStatusShipped         = "shipped"         // 已发货（原 in_transit「运输中」）
	POStatusPartialReceived = "partial_received"
	POStatusCompleted       = "completed"
	POStatusCancelled       = "cancelled"

	POPayStatusUnpaid  = "unpaid"
	POPayStatusPartial = "partial"
	POPayStatusPaid    = "paid"

	POFulfillmentStockIn   = "stock_in"
	POFulfillmentDropship  = "dropship"

	AddressTypeShip   = "ship"
	AddressTypeReturn = "return"

	ShipmentStatusPending   = "pending"
	ShipmentStatusShipped   = "shipped"
	ShipmentStatusInTransit = "in_transit"
	ShipmentStatusDelivered = "delivered"
	ShipmentStatusException = "exception"

	AttachmentTypeSupplierSalesOrder = "supplier_sales_order"
	AttachmentTypePaymentScreenshot  = "payment_screenshot"
	AttachmentTypeShipmentPhoto      = "shipment_photo" // 发货记录 / 物流单号照片等
	AttachmentTypeContract           = "contract"
	AttachmentTypeOther              = "other"
)
