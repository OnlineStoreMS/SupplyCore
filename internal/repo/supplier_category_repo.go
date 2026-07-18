package repo

import (
	"supplycore/internal/model"

	"gorm.io/gorm"
)

type SupplierCategoryRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewSupplierCategoryRepo(db *gorm.DB) *SupplierCategoryRepo {
	return &SupplierCategoryRepo{db: db}
}

func (r *SupplierCategoryRepo) ForTenant(tenantID uint64) *SupplierCategoryRepo {
	return &SupplierCategoryRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *SupplierCategoryRepo) List() ([]model.SupplierCategory, error) {
	var list []model.SupplierCategory
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *SupplierCategoryRepo) GetByID(id uint64) (*model.SupplierCategory, error) {
	var item model.SupplierCategory
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierCategoryRepo) Create(item *model.SupplierCategory) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *SupplierCategoryRepo) Save(item *model.SupplierCategory) error {
	return r.db.Save(item).Error
}

func (r *SupplierCategoryRepo) Delete(id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).Delete(&model.SupplierCategory{}, id).Error
}

func (r *SupplierCategoryRepo) CountSuppliers(categoryID uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Supplier{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("category_id = ?", categoryID).
		Count(&count).Error
	return count, err
}
