package repo

import (
	"supplycore/internal/model"

	"gorm.io/gorm"
)

type OfferRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewOfferRepo(db *gorm.DB) *OfferRepo {
	return &OfferRepo{db: db}
}

func (r *OfferRepo) ForTenant(tenantID uint64) *OfferRepo {
	return &OfferRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

type OfferListFilter struct {
	SkuID      uint64
	SupplierID uint64
	Keyword    string
	Page       int
	PageSize   int
}

func (r *OfferRepo) List(f OfferListFilter) ([]model.SkuSupplierOffer, int64, error) {
	q := r.db.Model(&model.SkuSupplierOffer{}).Scopes(scopeTenant(r.tenantID))
	if f.SkuID > 0 {
		q = q.Where("sku_id = ?", f.SkuID)
	}
	if f.SupplierID > 0 {
		q = q.Where("supplier_id = ?", f.SupplierID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SkuSupplierOffer
	offset := (f.Page - 1) * f.PageSize
	err := q.Order("sku_id ASC, is_primary DESC, priority DESC, supply_price ASC, id ASC").
		Offset(offset).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *OfferRepo) ListBySku(skuID uint64, activeOnly bool) ([]model.SkuSupplierOffer, error) {
	q := r.db.Scopes(scopeTenant(r.tenantID)).Where("sku_id = ?", skuID)
	if activeOnly {
		q = q.Where("status = 1")
	}
	var list []model.SkuSupplierOffer
	err := q.Order("is_primary DESC, priority DESC, supply_price ASC").Find(&list).Error
	return list, err
}

func (r *OfferRepo) GetByID(id uint64) (*model.SkuSupplierOffer, error) {
	var item model.SkuSupplierOffer
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OfferRepo) Create(item *model.SkuSupplierOffer) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *OfferRepo) Save(item *model.SkuSupplierOffer) error {
	return r.db.Save(item).Error
}

func (r *OfferRepo) Delete(id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).Delete(&model.SkuSupplierOffer{}, id).Error
}

func (r *OfferRepo) ClearPrimary(skuID uint64, exceptID uint64) error {
	q := r.db.Model(&model.SkuSupplierOffer{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("sku_id = ?", skuID)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_primary", false).Error
}
