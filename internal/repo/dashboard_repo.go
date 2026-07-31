package repo

import (
	"time"

	"supplycore/internal/model"

	"gorm.io/gorm"
)

type DashboardRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewDashboardRepo(db *gorm.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (r *DashboardRepo) ForTenant(tenantID uint64) *DashboardRepo {
	return &DashboardRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *DashboardRepo) CountSuppliers(activeOnly bool) (int64, error) {
	q := r.db.Model(&model.Supplier{}).Scopes(scopeTenant(r.tenantID))
	if activeOnly {
		q = q.Where("status = ?", 1)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountOffers(activeOnly bool) (int64, error) {
	q := r.db.Model(&model.SkuSupplierOffer{}).Scopes(scopeTenant(r.tenantID))
	if activeOnly {
		q = q.Where("status = ?", 1)
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountPOsByStatus(status string) (int64, error) {
	return r.CountPOsByStatusSince(status, nil)
}

func (r *DashboardRepo) CountPOsByStatusSince(status string, dayStart *time.Time) (int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("status = ?", status)
	q = scopePOBusinessDay(q, dayStart)
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountPOsByStatuses(statuses []string) (int64, error) {
	return r.CountPOsByStatusesSince(statuses, nil)
}

func (r *DashboardRepo) CountPOsByStatusesSince(statuses []string, dayStart *time.Time) (int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("status IN ?", statuses)
	q = scopePOBusinessDay(q, dayStart)
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountPOsByFulfillment(fulfillmentType string, excludeDraftCancel bool) (int64, error) {
	return r.CountPOsByFulfillmentSince(fulfillmentType, excludeDraftCancel, nil)
}

func (r *DashboardRepo) CountPOsByFulfillmentSince(fulfillmentType string, excludeDraftCancel bool, dayStart *time.Time) (int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("fulfillment_type = ?", fulfillmentType)
	if excludeDraftCancel {
		q = q.Where("status NOT IN ?", []string{"draft", "cancelled"})
	}
	q = scopePOBusinessDay(q, dayStart)
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountUnpaidPOs() (int64, error) {
	return r.CountUnpaidPOsSince(nil)
}

func (r *DashboardRepo) CountUnpaidPOsSince(dayStart *time.Time) (int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("pay_status IN ?", []string{"unpaid", "partial"}).
		Where("status NOT IN ?", []string{"draft", "cancelled"})
	q = scopePOBusinessDay(q, dayStart)
	var n int64
	err := q.Count(&n).Error
	return n, err
}

// scopePOBusinessDay 按业务日筛选：COALESCE(ordered_at, created_at) 落在 [dayStart, dayStart+1)。
func scopePOBusinessDay(q *gorm.DB, dayStart *time.Time) *gorm.DB {
	if dayStart == nil {
		return q
	}
	dayEnd := dayStart.AddDate(0, 0, 1)
	return q.Where("COALESCE(ordered_at, created_at) >= ? AND COALESCE(ordered_at, created_at) < ?", *dayStart, dayEnd)
}

func (r *DashboardRepo) CountPOs() (int64, error) {
	var n int64
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Count(&n).Error
	return n, err
}

func (r *DashboardRepo) CountPOsSince(since time.Time, excludeDraftCancel bool) (int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("COALESCE(ordered_at, created_at) >= ?", since)
	if excludeDraftCancel {
		q = q.Where("status NOT IN ?", []string{"draft", "cancelled"})
	}
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *DashboardRepo) SumPOAmountSince(since time.Time) (float64, error) {
	var sum float64
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("status NOT IN ?", []string{"draft", "cancelled"}).
		Where("COALESCE(ordered_at, created_at) >= ?", since).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *DashboardRepo) SumUnpaidAmount() (float64, error) {
	var sum float64
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("pay_status IN ?", []string{"unpaid", "partial"}).
		Where("status NOT IN ?", []string{"draft", "cancelled"}).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&sum).Error
	return sum, err
}

type StatusCountRow struct {
	Status string
	Count  int64
}

func (r *DashboardRepo) GroupPOByStatus() ([]StatusCountRow, error) {
	var rows []StatusCountRow
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&rows).Error
	return rows, err
}

type SupplierRankRow struct {
	SupplierID  uint64
	OrderCount  int64
	TotalAmount float64
}

func (r *DashboardRepo) TopSuppliersSince(since time.Time, limit int) ([]SupplierRankRow, error) {
	var rows []SupplierRankRow
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("status NOT IN ?", []string{"draft", "cancelled"}).
		Where("COALESCE(ordered_at, created_at) >= ?", since).
		Select("supplier_id, COUNT(*) as order_count, COALESCE(SUM(total_amount), 0) as total_amount").
		Group("supplier_id").
		Order("total_amount DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *DashboardRepo) CountDistinctSuppliersSince(since time.Time) (int64, error) {
	var n int64
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("status NOT IN ?", []string{"draft", "cancelled"}).
		Where("COALESCE(ordered_at, created_at) >= ?", since).
		Select("COUNT(DISTINCT supplier_id)").
		Scan(&n).Error
	return n, err
}

func (r *DashboardRepo) RecentPOs(limit int) ([]model.PurchaseOrder, error) {
	var list []model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
