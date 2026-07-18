package repo

import "gorm.io/gorm"

type Repos struct {
	Supplier         *SupplierRepo
	Offer            *OfferRepo
	PurchaseOrder    *PurchaseOrderRepo
	Shipment         *ShipmentRepo
	Payment          *PaymentRepo
	Attachment       *AttachmentRepo
	SalesOrder       *SalesOrderRepo
	PurchaseAccount  *PurchaseAccountRepo
	PurchaseInbound  *PurchaseInboundRepo
	PackageReceive   *PackageReceiveRepo
	PurchaseReturn   *PurchaseReturnRepo
}

func New(db *gorm.DB) *Repos {
	return &Repos{
		Supplier:        NewSupplierRepo(db),
		Offer:           NewOfferRepo(db),
		PurchaseOrder:   NewPurchaseOrderRepo(db),
		Shipment:        NewShipmentRepo(db),
		Payment:         NewPaymentRepo(db),
		Attachment:      NewAttachmentRepo(db),
		SalesOrder:      NewSalesOrderRepo(db),
		PurchaseAccount: NewPurchaseAccountRepo(db),
		PurchaseInbound: NewPurchaseInboundRepo(db),
		PackageReceive:  NewPackageReceiveRepo(db),
		PurchaseReturn:  NewPurchaseReturnRepo(db),
	}
}
