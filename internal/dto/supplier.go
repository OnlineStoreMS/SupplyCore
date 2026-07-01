package dto

type SupplierDTO struct {
	Code                string `json:"code" binding:"required"`
	Name                string `json:"name" binding:"required"`
	ShortName           string `json:"shortName"`
	Status              int8   `json:"status"`
	ContactName         string `json:"contactName"`
	Phone               string `json:"phone"`
	Email               string `json:"email"`
	Remark              string `json:"remark"`
	DefaultPaymentTerms string `json:"defaultPaymentTerms"`
	BankName            string `json:"bankName"`
	BankAccount         string `json:"bankAccount"`
	AccountName         string `json:"accountName"`
}

type SupplierAddressDTO struct {
	Label       string `json:"label" binding:"required"`
	ContactName string `json:"contactName"`
	Phone       string `json:"phone"`
	Province    string `json:"province"`
	City        string `json:"city"`
	District    string `json:"district"`
	Address     string `json:"address"`
	IsDefault   bool   `json:"isDefault"`
	Status      int8   `json:"status"`
}

type SkuOfferDTO struct {
	SkuID             uint64  `json:"skuId" binding:"required"`
	SupplierID        uint64  `json:"supplierId" binding:"required"`
	SupplierSkuCode   string  `json:"supplierSkuCode"`
	SupplyPrice       float64 `json:"supplyPrice" binding:"required"`
	Currency          string  `json:"currency"`
	MinOrderQty       int     `json:"minOrderQty"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	ShipFromAddressID uint64  `json:"shipFromAddressId"`
	SupportsDropship  bool    `json:"supportsDropship"`
	SupportsSelfStock bool    `json:"supportsSelfStock"`
	IsPrimary         bool    `json:"isPrimary"`
	Priority          int     `json:"priority"`
	Status            int8    `json:"status"`
	Remark            string  `json:"remark"`
}

type ShipFromBrief struct {
	Label    string `json:"label"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

type SupplyOptionOffer struct {
	OfferID          uint64         `json:"offerId"`
	SupplierID       uint64         `json:"supplierId"`
	SupplierName     string         `json:"supplierName"`
	SupplierCode     string         `json:"supplierCode"`
	SupplierSkuCode  string         `json:"supplierSkuCode"`
	SupplyPrice      float64        `json:"supplyPrice"`
	Currency         string         `json:"currency"`
	SupportsDropship bool           `json:"supportsDropship"`
	SupportsSelfStock bool          `json:"supportsSelfStock"`
	LeadTimeDays     int            `json:"leadTimeDays"`
	IsPrimary        bool           `json:"isPrimary"`
	Priority         int            `json:"priority"`
	ShipFrom         *ShipFromBrief `json:"shipFrom,omitempty"`
}

type SupplyOptionsResp struct {
	SkuID  uint64              `json:"skuId"`
	Offers []SupplyOptionOffer `json:"offers"`
}

type OfferDetail struct {
	ID                uint64  `json:"id"`
	SkuID             uint64  `json:"skuId"`
	SupplierID        uint64  `json:"supplierId"`
	SupplierName      string  `json:"supplierName"`
	SupplierCode      string  `json:"supplierCode"`
	SupplierSkuCode   string  `json:"supplierSkuCode"`
	SupplyPrice       float64 `json:"supplyPrice"`
	Currency          string  `json:"currency"`
	MinOrderQty       int     `json:"minOrderQty"`
	LeadTimeDays      int     `json:"leadTimeDays"`
	ShipFromAddressID uint64  `json:"shipFromAddressId"`
	ShipFromLabel     string  `json:"shipFromLabel"`
	ShipFromCity      string  `json:"shipFromCity"`
	SupportsDropship  bool    `json:"supportsDropship"`
	SupportsSelfStock bool    `json:"supportsSelfStock"`
	IsPrimary         bool    `json:"isPrimary"`
	Priority          int     `json:"priority"`
	Status            int8    `json:"status"`
	Remark            string  `json:"remark"`
}
