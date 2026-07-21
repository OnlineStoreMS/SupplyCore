package repo

import (
	"fmt"
	"time"

	"supplycore/internal/model"

	"gorm.io/gorm"
)

type PurchaseOrderRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPurchaseOrderRepo(db *gorm.DB) *PurchaseOrderRepo {
	return &PurchaseOrderRepo{db: db}
}

func (r *PurchaseOrderRepo) ForTenant(tenantID uint64) *PurchaseOrderRepo {
	return &PurchaseOrderRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

type POListFilter struct {
	Status          string
	FulfillmentType string
	SupplierID      uint64
	RefSoID         uint64
	RefTraceID      string
	Keyword         string
	Page            int
	PageSize        int
}

func (r *PurchaseOrderRepo) List(f POListFilter) ([]model.PurchaseOrder, int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).Scopes(scopeTenant(r.tenantID))
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.FulfillmentType != "" {
		q = q.Where("fulfillment_type = ?", f.FulfillmentType)
	}
	if f.SupplierID > 0 {
		q = q.Where("supplier_id = ?", f.SupplierID)
	}
	if f.RefSoID > 0 {
		q = q.Where("ref_so_id = ?", f.RefSoID)
	}
	if f.RefTraceID != "" {
		q = q.Where("ref_trace_id = ?", f.RefTraceID)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("po_no ILIKE ?", like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PurchaseOrder
	offset := (f.Page - 1) * f.PageSize
	err := q.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *PurchaseOrderRepo) GetByID(id uint64) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepo) GetWithItems(id uint64) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepo) Create(po *model.PurchaseOrder, items []model.PurchaseOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		po.TenantID = r.tenantID
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].POID = po.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		po.Items = items
		return nil
	})
}

func (r *PurchaseOrderRepo) ReplaceItems(poID uint64, items []model.PurchaseOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scopeTenant(r.tenantID)).
			Where("po_id = ?", poID).
			Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].POID = poID
		}
		if len(items) > 0 {
			return tx.Create(&items).Error
		}
		return nil
	})
}

func (r *PurchaseOrderRepo) Save(po *model.PurchaseOrder) error {
	return r.db.Save(po).Error
}

func (r *PurchaseOrderRepo) SaveItem(item *model.PurchaseOrderItem) error {
	return r.db.Save(item).Error
}

func (r *PurchaseOrderRepo) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scopeTenant(r.tenantID)).
			Where("po_id = ?", id).
			Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Scopes(scopeTenant(r.tenantID)).Delete(&model.PurchaseOrder{}, id).Error
	})
}

func (r *PurchaseOrderRepo) CountItems(poID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.PurchaseOrderItem{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ?", poID).
		Count(&n).Error
	return n, err
}

func (r *PurchaseOrderRepo) NextPoNo() (string, error) {
	prefix := "PO" + time.Now().Format("20060102")
	var count int64
	if err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("po_no LIKE ?", prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}
