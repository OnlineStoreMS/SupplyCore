package model

const (
	POStatusDraft            = "draft"
	POStatusOrdered          = "ordered"
	POStatusPaid             = "paid"
	POStatusPartialShipped   = "partial_shipped"
	POStatusInTransit        = "in_transit"
	POStatusPartialReceived  = "partial_received"
	POStatusCompleted        = "completed"
	POStatusCancelled        = "cancelled"

	POPayStatusUnpaid  = "unpaid"
	POPayStatusPartial = "partial"
	POPayStatusPaid    = "paid"

	POFulfillmentStockIn   = "stock_in"
	POFulfillmentDropship  = "dropship"

	ShipmentStatusPending   = "pending"
	ShipmentStatusShipped   = "shipped"
	ShipmentStatusInTransit = "in_transit"
	ShipmentStatusDelivered = "delivered"
	ShipmentStatusException = "exception"

	AttachmentTypeSupplierSalesOrder = "supplier_sales_order"
	AttachmentTypePaymentScreenshot  = "payment_screenshot"
	AttachmentTypeContract           = "contract"
	AttachmentTypeOther              = "other"
)
