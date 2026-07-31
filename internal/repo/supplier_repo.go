package repo

import (
	"time"

	"supplycore/internal/model"

	"gorm.io/gorm"
)

type SupplierRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewSupplierRepo(db *gorm.DB) *SupplierRepo {
	return &SupplierRepo{db: db}
}

func (r *SupplierRepo) ForTenant(tenantID uint64) *SupplierRepo {
	return &SupplierRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *SupplierRepo) List(keyword string, categoryID uint64, page, pageSize int) ([]model.Supplier, int64, error) {
	q := r.db.Model(&model.Supplier{}).Scopes(scopeTenant(r.tenantID))
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name ILIKE ? OR code ILIKE ? OR short_name ILIKE ?", like, like, like)
	}
	if categoryID > 0 {
		q = q.Where("category_id = ?", categoryID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.Supplier
	offset := (page - 1) * pageSize
	err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// ListWithSettlementCycle 已配置结算周期的供应商（跨租户，供调度器用）。
func (r *SupplierRepo) ListWithSettlementCycle() ([]model.Supplier, error) {
	var list []model.Supplier
	err := r.db.Where("settlement_cycle IN ?", []string{"day", "week", "month", "custom"}).
		Order("tenant_id ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *SupplierRepo) UpdateSettlementLastRunAt(id uint64, at time.Time) error {
	return r.db.Model(&model.Supplier{}).Where("id = ?", id).Update("settlement_last_run_at", at).Error
}

func (r *SupplierRepo) GetByID(id uint64) (*model.Supplier, error) {
	var item model.Supplier
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierRepo) GetByCode(code string) (*model.Supplier, error) {
	var item model.Supplier
	err := r.db.Scopes(scopeTenant(r.tenantID)).Where("code = ?", code).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierRepo) Create(item *model.Supplier) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *SupplierRepo) Save(item *model.Supplier) error {
	return r.db.Save(item).Error
}

func (r *SupplierRepo) Delete(id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).Delete(&model.Supplier{}, id).Error
}

func (r *SupplierRepo) ListAddresses(supplierID uint64, addressType string) ([]model.SupplierAddress, error) {
	var list []model.SupplierAddress
	q := r.db.Scopes(scopeTenant(r.tenantID)).Where("supplier_id = ?", supplierID)
	if addressType != "" {
		q = q.Where("address_type = ?", addressType)
	}
	err := q.Order("is_default DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *SupplierRepo) GetAddress(supplierID, addressID uint64) (*model.SupplierAddress, error) {
	var item model.SupplierAddress
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, addressID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierRepo) CreateAddress(item *model.SupplierAddress) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *SupplierRepo) SaveAddress(item *model.SupplierAddress) error {
	return r.db.Save(item).Error
}

func (r *SupplierRepo) DeleteAddress(supplierID, addressID uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, addressID).
		Delete(&model.SupplierAddress{}).Error
}

func (r *SupplierRepo) ClearDefaultAddress(supplierID uint64, addressType string, exceptID uint64) error {
	q := r.db.Model(&model.SupplierAddress{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID)
	if addressType != "" {
		q = q.Where("address_type = ?", addressType)
	}
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_default", false).Error
}

func (r *SupplierRepo) ListPaymentAccounts(supplierID uint64) ([]model.SupplierPaymentAccount, error) {
	var list []model.SupplierPaymentAccount
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID).
		Order("is_default DESC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *SupplierRepo) GetPaymentAccount(supplierID, accountID uint64) (*model.SupplierPaymentAccount, error) {
	var item model.SupplierPaymentAccount
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, accountID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierRepo) CreatePaymentAccount(item *model.SupplierPaymentAccount) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *SupplierRepo) SavePaymentAccount(item *model.SupplierPaymentAccount) error {
	return r.db.Save(item).Error
}

func (r *SupplierRepo) DeletePaymentAccount(supplierID, accountID uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, accountID).
		Delete(&model.SupplierPaymentAccount{}).Error
}

func (r *SupplierRepo) ClearDefaultPaymentAccount(supplierID uint64, exceptID uint64) error {
	q := r.db.Model(&model.SupplierPaymentAccount{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_default", false).Error
}

func (r *SupplierRepo) ListPaymentQRs(supplierID uint64) ([]model.SupplierPaymentQR, error) {
	var list []model.SupplierPaymentQR
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID).
		Order("is_default DESC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *SupplierRepo) GetPaymentQR(supplierID, qrID uint64) (*model.SupplierPaymentQR, error) {
	var item model.SupplierPaymentQR
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, qrID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SupplierRepo) CreatePaymentQR(item *model.SupplierPaymentQR) error {
	item.TenantID = r.tenantID
	return r.db.Create(item).Error
}

func (r *SupplierRepo) SavePaymentQR(item *model.SupplierPaymentQR) error {
	return r.db.Save(item).Error
}

func (r *SupplierRepo) DeletePaymentQR(supplierID, qrID uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ? AND id = ?", supplierID, qrID).
		Delete(&model.SupplierPaymentQR{}).Error
}

func (r *SupplierRepo) ClearDefaultPaymentQR(supplierID uint64, exceptID uint64) error {
	q := r.db.Model(&model.SupplierPaymentQR{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	return q.Update("is_default", false).Error
}
