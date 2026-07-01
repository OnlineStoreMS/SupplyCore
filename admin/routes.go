package admin

import "github.com/gin-gonic/gin"

func RegisterRoutes(g *gin.RouterGroup, supplierH *SupplierHandler, offerH *OfferHandler, poH *PurchaseOrderHandler, trackH *POTrackingHandler, soH *SalesOrderHandler, skuH *ProductSkuHandler) {
	g.GET("/suppliers", supplierH.List)
	g.POST("/suppliers", supplierH.Create)
	g.GET("/suppliers/:id", supplierH.Get)
	g.PUT("/suppliers/:id", supplierH.Update)
	g.DELETE("/suppliers/:id", supplierH.Delete)

	g.GET("/suppliers/:id/addresses", supplierH.ListAddresses)
	g.POST("/suppliers/:id/addresses", supplierH.CreateAddress)
	g.PUT("/suppliers/:id/addresses/:addressId", supplierH.UpdateAddress)
	g.DELETE("/suppliers/:id/addresses/:addressId", supplierH.DeleteAddress)

	g.GET("/sku-offers", offerH.List)
	g.POST("/sku-offers", offerH.Create)
	g.GET("/sku-offers/:id", offerH.Get)
	g.PUT("/sku-offers/:id", offerH.Update)
	g.DELETE("/sku-offers/:id", offerH.Delete)

	g.GET("/skus/:id/supply-options", offerH.SupplyOptions)

	g.GET("/product-skus/search", skuH.Search)

	g.GET("/purchase-orders", poH.List)
	g.POST("/purchase-orders", poH.Create)
	g.GET("/purchase-orders/:id", poH.Get)
	g.PUT("/purchase-orders/:id", poH.Update)
	g.DELETE("/purchase-orders/:id", poH.Delete)
	g.POST("/purchase-orders/:id/submit", poH.Submit)
	g.POST("/purchase-orders/:id/mark-paid", poH.MarkPaid)
	g.POST("/purchase-orders/:id/complete", poH.Complete)
	g.POST("/purchase-orders/:id/cancel", poH.Cancel)

	g.POST("/upload", trackH.Upload)

	g.GET("/purchase-orders/:id/shipments", trackH.ListShipments)
	g.POST("/purchase-orders/:id/shipments", trackH.CreateShipment)
	g.PATCH("/purchase-orders/:id/shipments/:shipmentId/status", trackH.UpdateShipmentStatus)
	g.DELETE("/purchase-orders/:id/shipments/:shipmentId", trackH.DeleteShipment)

	g.GET("/purchase-orders/:id/payments", trackH.ListPayments)
	g.POST("/purchase-orders/:id/payments", trackH.CreatePayment)
	g.PUT("/purchase-orders/:id/payments/:paymentId", trackH.UpdatePayment)
	g.DELETE("/purchase-orders/:id/payments/:paymentId", trackH.DeletePayment)

	g.GET("/purchase-orders/:id/attachments", trackH.ListAttachments)
	g.POST("/purchase-orders/:id/attachments", trackH.CreateAttachment)
	g.DELETE("/purchase-orders/:id/attachments/:attachmentId", trackH.DeleteAttachment)

	g.GET("/sales-orders", soH.List)
	g.POST("/sales-orders", soH.Create)
	g.GET("/sales-orders/:id", soH.Get)
	g.PUT("/sales-orders/:id", soH.Update)
	g.DELETE("/sales-orders/:id", soH.Delete)
	g.POST("/sales-orders/:id/confirm", soH.Confirm)
	g.POST("/sales-orders/:id/evaluate-sourcing", soH.EvaluateSourcing)
	g.POST("/sales-orders/:id/create-purchase-orders", soH.CreatePurchaseOrders)

	g.POST("/sourcing/dropship-purchase-order", soH.DropshipPurchaseOrder)
}
