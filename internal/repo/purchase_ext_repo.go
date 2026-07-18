package repo

import (
	"fmt"
	"time"

	"supplycore/internal/model"

	"gorm.io/gorm"
)

type PurchaseAccountRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPurchaseAccountRepo(db *gorm.DB) *PurchaseAccountRepo {
	return &PurchaseAccountRepo{db: db}
}

func (r *PurchaseAccountRepo) ForTenant(tenantID uint64) *PurchaseAccountRepo {
	return &PurchaseAccountRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *PurchaseAccountRepo) List(keyword string, page, pageSize int) ([]model.PurchaseAccount, int64, error) {
	q := r.db.Model(&model.PurchaseAccount{}).Where("tenant_id = ?", r.tenantID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("account_alias ILIKE ? OR account_name ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PurchaseAccount
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *PurchaseAccountRepo) GetByID(id uint64) (*model.PurchaseAccount, error) {
	var m model.PurchaseAccount
	err := r.db.Where("tenant_id = ? AND id = ?", r.tenantID, id).First(&m).Error
	return &m, err
}

func (r *PurchaseAccountRepo) Create(m *model.PurchaseAccount) error {
	m.TenantID = r.tenantID
	return r.db.Create(m).Error
}

func (r *PurchaseAccountRepo) Save(m *model.PurchaseAccount) error {
	return r.db.Save(m).Error
}

func (r *PurchaseAccountRepo) Delete(id uint64) error {
	return r.db.Where("tenant_id = ? AND id = ?", r.tenantID, id).Delete(&model.PurchaseAccount{}).Error
}

type PurchaseInboundRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPurchaseInboundRepo(db *gorm.DB) *PurchaseInboundRepo {
	return &PurchaseInboundRepo{db: db}
}

func (r *PurchaseInboundRepo) ForTenant(tenantID uint64) *PurchaseInboundRepo {
	return &PurchaseInboundRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *PurchaseInboundRepo) List(status, keyword string, page, pageSize int) ([]model.PurchaseInbound, int64, error) {
	q := r.db.Model(&model.PurchaseInbound{}).Where("tenant_id = ?", r.tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("inbound_no ILIKE ? OR po_no ILIKE ? OR tracking_no ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PurchaseInbound
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *PurchaseInboundRepo) GetWithItems(id uint64) (*model.PurchaseInbound, error) {
	var m model.PurchaseInbound
	err := r.db.Preload("Items").Where("tenant_id = ? AND id = ?", r.tenantID, id).First(&m).Error
	return &m, err
}

func (r *PurchaseInboundRepo) NextNo() (string, error) {
	var count int64
	day := time.Now().Format("20060102")
	prefix := "IN" + day
	if err := r.db.Model(&model.PurchaseInbound{}).
		Where("tenant_id = ? AND inbound_no LIKE ?", r.tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *PurchaseInboundRepo) Create(m *model.PurchaseInbound, items []model.PurchaseInboundItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		m.TenantID = r.tenantID
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].InboundID = m.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PurchaseInboundRepo) Save(m *model.PurchaseInbound) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: false}).Save(m).Error
}

type PackageReceiveRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPackageReceiveRepo(db *gorm.DB) *PackageReceiveRepo {
	return &PackageReceiveRepo{db: db}
}

func (r *PackageReceiveRepo) ForTenant(tenantID uint64) *PackageReceiveRepo {
	return &PackageReceiveRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *PackageReceiveRepo) List(keyword string, page, pageSize int) ([]model.PackageReceiveRecord, int64, error) {
	q := r.db.Model(&model.PackageReceiveRecord{}).Where("tenant_id = ?", r.tenantID)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("tracking_no ILIKE ? OR po_no ILIKE ? OR carrier ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PackageReceiveRecord
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *PackageReceiveRepo) Create(m *model.PackageReceiveRecord) error {
	m.TenantID = r.tenantID
	return r.db.Create(m).Error
}

func (r *PackageReceiveRepo) GetByID(id uint64) (*model.PackageReceiveRecord, error) {
	var m model.PackageReceiveRecord
	err := r.db.Where("tenant_id = ? AND id = ?", r.tenantID, id).First(&m).Error
	return &m, err
}

func (r *PackageReceiveRepo) Save(m *model.PackageReceiveRecord) error {
	return r.db.Save(m).Error
}

type PurchaseReturnRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPurchaseReturnRepo(db *gorm.DB) *PurchaseReturnRepo {
	return &PurchaseReturnRepo{db: db}
}

func (r *PurchaseReturnRepo) ForTenant(tenantID uint64) *PurchaseReturnRepo {
	return &PurchaseReturnRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *PurchaseReturnRepo) List(status, keyword string, page, pageSize int) ([]model.PurchaseReturn, int64, error) {
	q := r.db.Model(&model.PurchaseReturn{}).Where("tenant_id = ?", r.tenantID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("return_no ILIKE ? OR inbound_no ILIKE ? OR tracking_no ILIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PurchaseReturn
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *PurchaseReturnRepo) GetWithItems(id uint64) (*model.PurchaseReturn, error) {
	var m model.PurchaseReturn
	err := r.db.Preload("Items").Where("tenant_id = ? AND id = ?", r.tenantID, id).First(&m).Error
	return &m, err
}

func (r *PurchaseReturnRepo) NextNo() (string, error) {
	var count int64
	day := time.Now().Format("20060102")
	prefix := "PR" + day
	if err := r.db.Model(&model.PurchaseReturn{}).
		Where("tenant_id = ? AND return_no LIKE ?", r.tenantID, prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *PurchaseReturnRepo) Create(m *model.PurchaseReturn, items []model.PurchaseReturnItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		m.TenantID = r.tenantID
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].ReturnID = m.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PurchaseReturnRepo) Save(m *model.PurchaseReturn) error {
	return r.db.Save(m).Error
}
