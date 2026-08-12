package model

const (
	POStatusDraft           = "draft"
	POStatusOrdered         = "ordered"
	POStatusAwaitingShip    = "awaiting_ship"   // 待发货（付清后、尚未登记物流）
	POStatusPaid            = "paid"            // 历史单据状态，等同 awaiting_ship
	POStatusPartialShipped  = "partial_shipped" // 部分发货
	POStatusShipped         = "shipped"         // 已发货（原 in_transit「运输中」）
	POStatusPartialReceived = "partial_received"
	POStatusCompleted       = "completed"
	POStatusCancelled       = "cancelled"

	POPayStatusUnpaid  = "unpaid"
	POPayStatusPartial = "partial"
	POPayStatusPaid    = "paid"

	POFulfillmentStockIn  = "stock_in"
	POFulfillmentDropship = "dropship"

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

// POStatusesAwaitingShip 待发货状态值（含历史 paid），列表/统计筛「待发货」时用。
func POStatusesAwaitingShip() []string {
	return []string{POStatusAwaitingShip, POStatusPaid}
}
