package repo

import (
	"fmt"
	"time"

	"supplycore/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SalesOrderRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewSalesOrderRepo(db *gorm.DB) *SalesOrderRepo {
	return &SalesOrderRepo{db: db}
}

func (r *SalesOrderRepo) ForTenant(tenantID uint64) *SalesOrderRepo {
	return &SalesOrderRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

type SOListFilter struct {
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

func (r *SalesOrderRepo) List(f SOListFilter) ([]model.SalesOrder, int64, error) {
	q := r.db.Model(&model.SalesOrder{}).Scopes(scopeTenant(r.tenantID))
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("so_no ILIKE ? OR trace_id ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.SalesOrder
	offset := (f.Page - 1) * f.PageSize
	err := q.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func (r *SalesOrderRepo) GetWithItems(id uint64) (*model.SalesOrder, error) {
	var so model.SalesOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("id ASC") }).
		First(&so, id).Error
	if err != nil {
		return nil, err
	}
	return &so, nil
}

func (r *SalesOrderRepo) Create(so *model.SalesOrder, items []model.SalesOrderItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		so.TenantID = r.tenantID
		if err := tx.Create(so).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].SOID = so.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		so.Items = items
		return nil
	})
}

func (r *SalesOrderRepo) Save(so *model.SalesOrder) error {
	return r.db.Save(so).Error
}

func (r *SalesOrderRepo) SaveItem(item *model.SalesOrderItem) error {
	return r.db.Save(item).Error
}

func (r *SalesOrderRepo) NextSoNo() (string, error) {
	prefix := "SO" + time.Now().Format("20060102")
	var count int64
	if err := r.db.Model(&model.SalesOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("so_no LIKE ?", prefix+"%").
		Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func (r *SalesOrderRepo) NewTraceID() string {
	return "TR" + time.Now().Format("20060102") + uuid.New().String()[:8]
}

func (r *SalesOrderRepo) ListPOIDsBySO(soID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.db.Model(&model.SalesOrderItem{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("so_id = ? AND linked_po_id > 0", soID).
		Distinct("linked_po_id").
		Pluck("linked_po_id", &ids).Error
	return ids, err
}

func (r *SalesOrderRepo) CountItems(soID uint64) (int64, error) {
	var n int64
	err := r.db.Model(&model.SalesOrderItem{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("so_id = ?", soID).
		Count(&n).Error
	return n, err
}
