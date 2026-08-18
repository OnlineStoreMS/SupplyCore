package admin

import "github.com/gin-gonic/gin"

func RegisterRoutes(g *gin.RouterGroup, supplierH *SupplierHandler, offerH *OfferHandler, poH *PurchaseOrderHandler, trackH *POTrackingHandler, skuH *ProductSkuHandler, whH *WarehouseHandler, extH *PurchaseExtHandler, dashH *DashboardHandler, orderH *OrderHandler) {
	g.GET("/dashboard/stats", dashH.Stats)
	g.GET("/dashboard/trend", dashH.Trend)

	g.GET("/orders/search", orderH.Search)
	g.POST("/orders/decrypt", orderH.Decrypt)
	g.GET("/orders/:id", orderH.Get)
	g.POST("/orders/:id/ship", orderH.Ship)

	g.GET("/supplier-categories", supplierH.ListCategories)
	g.POST("/supplier-categories", supplierH.CreateCategory)
	g.PUT("/supplier-categories/:id", supplierH.UpdateCategory)
	g.DELETE("/supplier-categories/:id", supplierH.DeleteCategory)

	g.GET("/suppliers", supplierH.List)
	g.POST("/suppliers", supplierH.Create)
	g.GET("/suppliers/:id", supplierH.Get)
	g.PUT("/suppliers/:id", supplierH.Update)
	g.DELETE("/suppliers/:id", supplierH.Delete)

	g.GET("/suppliers/:id/addresses", supplierH.ListAddresses)
	g.POST("/suppliers/:id/addresses", supplierH.CreateAddress)
	g.PUT("/suppliers/:id/addresses/:addressId", supplierH.UpdateAddress)
	g.DELETE("/suppliers/:id/addresses/:addressId", supplierH.DeleteAddress)

	g.GET("/suppliers/:id/payment-accounts", supplierH.ListPaymentAccounts)
	g.POST("/suppliers/:id/payment-accounts", supplierH.CreatePaymentAccount)
	g.PUT("/suppliers/:id/payment-accounts/:accountId", supplierH.UpdatePaymentAccount)
	g.DELETE("/suppliers/:id/payment-accounts/:accountId", supplierH.DeletePaymentAccount)

	g.GET("/suppliers/:id/payment-qrs", supplierH.ListPaymentQRs)
	g.POST("/suppliers/:id/payment-qrs", supplierH.CreatePaymentQR)
	g.PUT("/suppliers/:id/payment-qrs/:qrId", supplierH.UpdatePaymentQR)
	g.DELETE("/suppliers/:id/payment-qrs/:qrId", supplierH.DeletePaymentQR)

	g.GET("/sku-offers", offerH.List)
	g.POST("/sku-offers", offerH.Create)
	g.GET("/sku-offers/:id", offerH.Get)
	g.PUT("/sku-offers/:id", offerH.Update)
	g.DELETE("/sku-offers/:id", offerH.Delete)

	g.GET("/skus/:id/supply-options", offerH.SupplyOptions)

	g.GET("/product-skus/search", skuH.Search)
	g.GET("/products/search", skuH.SearchProducts)
	g.GET("/products/:id/skus", skuH.GetProductSkus)

	g.GET("/warehouses", whH.List)
	g.GET("/warehouses/:id/locations", whH.ListLocations)

	g.GET("/purchase-orders", poH.List)
	g.POST("/purchase-orders", poH.Create)
	g.POST("/purchase-orders/merge", poH.Merge)
	g.POST("/purchase-orders/detach-sales-order", poH.DetachSalesOrder)
	g.GET("/purchase-orders/:id", poH.Get)
	g.PUT("/purchase-orders/:id", poH.Update)
	g.PUT("/purchase-orders/:id/item-prices", poH.UpdateItemPrices)
	g.POST("/purchase-orders/:id/sync-purchase-prices", poH.SyncPurchasePrices)
	g.DELETE("/purchase-orders/:id", poH.Delete)
	g.POST("/purchase-orders/:id/submit", poH.Submit)
	g.POST("/purchase-orders/:id/mark-paid", poH.MarkPaid)
	g.POST("/purchase-orders/:id/complete", poH.Complete)
	g.POST("/purchase-orders/:id/cancel", poH.Cancel)

	g.POST("/upload", trackH.Upload)

	g.GET("/purchase-orders/:id/shipments", trackH.ListShipments)
	g.POST("/purchase-orders/:id/shipments", trackH.CreateShipment)
	g.POST("/purchase-orders/:id/shipments/sync-from-orders", trackH.SyncShipmentsFromOrders)
	g.POST("/purchase-orders/:id/items/:itemId/split", trackH.SplitItem)
	g.PATCH("/purchase-orders/:id/shipments/:shipmentId/status", trackH.UpdateShipmentStatus)
	g.DELETE("/purchase-orders/:id/shipments/:shipmentId", trackH.DeleteShipment)

	g.GET("/purchase-orders/:id/payments", trackH.ListPayments)
	g.POST("/purchase-orders/:id/payments", trackH.CreatePayment)
	g.PUT("/purchase-orders/:id/payments/:paymentId", trackH.UpdatePayment)
	g.DELETE("/purchase-orders/:id/payments/:paymentId", trackH.DeletePayment)

	g.GET("/purchase-orders/:id/attachments", trackH.ListAttachments)
	g.POST("/purchase-orders/:id/attachments", trackH.CreateAttachment)
	g.DELETE("/purchase-orders/:id/attachments/:attachmentId", trackH.DeleteAttachment)

	// M5 · 采购扩展（对齐普源：账号 / 入库 / 收包 / 退回 / 建议）
	g.GET("/purchase-accounts", extH.ListAccounts)
	g.POST("/purchase-accounts", extH.CreateAccount)
	g.PUT("/purchase-accounts/:id", extH.UpdateAccount)
	g.DELETE("/purchase-accounts/:id", extH.DeleteAccount)

	g.GET("/purchase-inbounds", extH.ListInbounds)
	g.POST("/purchase-inbounds", extH.CreateInbound)
	g.GET("/purchase-inbounds/:id", extH.GetInbound)
	g.POST("/purchase-inbounds/:id/approve-wh", extH.ApproveInboundWH)
	g.POST("/purchase-inbounds/:id/approve-finance", extH.ApproveInboundFinance)
	g.POST("/purchase-inbounds/:id/void", extH.VoidInbound)

	g.GET("/package-receives", extH.ListPackageReceives)
	g.POST("/package-receives/scan", extH.ScanPackage)
	g.POST("/package-receives/:id/create-inbound", extH.CreateInboundFromPackage)

	g.GET("/purchase-returns", extH.ListReturns)
	g.POST("/purchase-returns", extH.CreateReturn)
	g.GET("/purchase-returns/:id", extH.GetReturn)
	g.POST("/purchase-returns/:id/approve", extH.ApproveReturn)
	g.POST("/purchase-returns/:id/approve-finance", extH.ApproveReturnFinance)
	g.POST("/purchase-returns/:id/void", extH.VoidReturn)

	g.GET("/purchase-suggestions", extH.ListSuggestions)
}
