package repo

import (
	"fmt"
	"strings"
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
	Statuses        []string // 多状态，优先于 Status
	PayStatuses      []string // unpaid|partial|paid
	ExcludeStatuses  []string
	FulfillmentType string
	SupplierID      uint64
	RefSoID         uint64
	RefTraceID      string
	Keyword         string
	CreatedAtStart  *time.Time
	CreatedAtEnd    *time.Time // 含当日：传次日 00:00 时用 < End
	OrderedAtStart  *time.Time
	OrderedAtEnd    *time.Time // 含当日：传次日 00:00 时用 < End；按业务日 COALESCE(ordered_at, created_at)
	SortBy          string // orderedAt | createdAt | id | totalAmount
	SortOrder       string // asc | desc
	Page            int
	PageSize        int
}

func (r *PurchaseOrderRepo) List(f POListFilter) ([]model.PurchaseOrder, int64, error) {
	q := r.db.Model(&model.PurchaseOrder{}).Scopes(scopeTenant(r.tenantID))
	if len(f.Statuses) > 0 {
		q = q.Where("status IN ?", f.Statuses)
	} else if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if len(f.ExcludeStatuses) > 0 {
		q = q.Where("status NOT IN ?", f.ExcludeStatuses)
	}
	if len(f.PayStatuses) > 0 {
		q = q.Where("pay_status IN ?", f.PayStatuses)
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
	if f.CreatedAtStart != nil {
		q = q.Where("created_at >= ?", *f.CreatedAtStart)
	}
	if f.CreatedAtEnd != nil {
		q = q.Where("created_at < ?", *f.CreatedAtEnd)
	}
	// 采购时间筛选与工作台业务日一致：COALESCE(ordered_at, created_at)
	if f.OrderedAtStart != nil {
		q = q.Where("COALESCE(ordered_at, created_at) >= ?", *f.OrderedAtStart)
	}
	if f.OrderedAtEnd != nil {
		q = q.Where("COALESCE(ordered_at, created_at) < ?", *f.OrderedAtEnd)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []model.PurchaseOrder
	offset := (f.Page - 1) * f.PageSize
	err := q.Order(poListOrderClause(f.SortBy, f.SortOrder)).Offset(offset).Limit(f.PageSize).Find(&list).Error
	return list, total, err
}

func poListOrderClause(sortBy, sortOrder string) string {
	dir := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		dir = "ASC"
	}
	switch strings.TrimSpace(sortBy) {
	case "orderedAt", "ordered_at":
		// 采购时间为空时回退创建时间，保证排序稳定
		return fmt.Sprintf("COALESCE(ordered_at, created_at) %s, id %s", dir, dir)
	case "createdAt", "created_at":
		return fmt.Sprintf("created_at %s, id %s", dir, dir)
	case "totalAmount", "total_amount":
		return fmt.Sprintf("total_amount %s, id %s", dir, dir)
	case "id":
		return fmt.Sprintf("id %s", dir)
	default:
		return "COALESCE(ordered_at, created_at) DESC, id DESC"
	}
}

// ListDropshipMergeable 可结算合并的代发单：草稿/已下单且未付款，按创建时间窗口。
func (r *PurchaseOrderRepo) ListDropshipMergeable(supplierID uint64, from, to time.Time) ([]model.PurchaseOrder, error) {
	var list []model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("supplier_id = ?", supplierID).
		Where("fulfillment_type = ?", model.POFulfillmentDropship).
		Where("status IN ?", []string{model.POStatusDraft, model.POStatusOrdered}).
		Where("pay_status = ?", model.POPayStatusUnpaid).
		Where("created_at >= ? AND created_at < ?", from, to).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *PurchaseOrderRepo) GetByID(id uint64) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).First(&po, id).Error
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepo) GetByPoNoWithItems(poNo string) (*model.PurchaseOrder, error) {
	poNo = strings.TrimSpace(poNo)
	if poNo == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var po model.PurchaseOrder
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Where("po_no = ?", poNo).
		First(&po).Error
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

// UpdateHeaderTimes 强制更新单头创建时间 / 采购时间（合并后对齐最新单）。
func (r *PurchaseOrderRepo) UpdateHeaderTimes(id uint64, createdAt time.Time, orderedAt *time.Time) error {
	fields := map[string]any{
		"created_at": createdAt,
		"updated_at": time.Now(),
	}
	if orderedAt != nil {
		fields["ordered_at"] = *orderedAt
	} else {
		fields["ordered_at"] = nil
	}
	return r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *PurchaseOrderRepo) SaveItem(item *model.PurchaseOrderItem) error {
	return r.db.Save(item).Error
}

func (r *PurchaseOrderRepo) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var shipmentIDs []uint64
		if err := tx.Model(&model.PurchaseShipment{}).
			Scopes(scopeTenant(r.tenantID)).
			Where("po_id = ?", id).
			Pluck("id", &shipmentIDs).Error; err != nil {
			return err
		}
		if len(shipmentIDs) > 0 {
			if err := tx.Scopes(scopeTenant(r.tenantID)).
				Where("shipment_id IN ?", shipmentIDs).
				Delete(&model.PurchaseShipmentItem{}).Error; err != nil {
				return err
			}
		}
		for _, m := range []any{
			&model.PurchaseShipment{},
			&model.PurchasePayment{},
			&model.PurchaseAttachment{},
			&model.PackageReceiveRecord{},
		} {
			if err := tx.Scopes(scopeTenant(r.tenantID)).
				Where("po_id = ?", id).
				Delete(m).Error; err != nil {
				return err
			}
		}
		var inboundIDs []uint64
		if err := tx.Model(&model.PurchaseInbound{}).
			Scopes(scopeTenant(r.tenantID)).
			Where("po_id = ?", id).
			Pluck("id", &inboundIDs).Error; err != nil {
			return err
		}
		if len(inboundIDs) > 0 {
			var returnIDs []uint64
			if err := tx.Model(&model.PurchaseReturn{}).
				Scopes(scopeTenant(r.tenantID)).
				Where("inbound_id IN ?", inboundIDs).
				Pluck("id", &returnIDs).Error; err != nil {
				return err
			}
			if len(returnIDs) > 0 {
				if err := tx.Scopes(scopeTenant(r.tenantID)).
					Where("return_id IN ?", returnIDs).
					Delete(&model.PurchaseReturnItem{}).Error; err != nil {
					return err
				}
				if err := tx.Scopes(scopeTenant(r.tenantID)).
					Where("id IN ?", returnIDs).
					Delete(&model.PurchaseReturn{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Scopes(scopeTenant(r.tenantID)).
				Where("inbound_id IN ?", inboundIDs).
				Delete(&model.PurchaseInboundItem{}).Error; err != nil {
				return err
			}
			if err := tx.Scopes(scopeTenant(r.tenantID)).
				Where("id IN ?", inboundIDs).
				Delete(&model.PurchaseInbound{}).Error; err != nil {
				return err
			}
		}
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
		Where("po_id = ? AND cancelled = ?", poID, false).
		Count(&n).Error
	return n, err
}

// ItemSpecsByPOIDs 批量汇总采购明细规格（跳过已撤回/空规格），同规格累加数量，输出如「规格 x2」。
func (r *PurchaseOrderRepo) ItemSpecsByPOIDs(poIDs []uint64) (map[uint64][]string, error) {
	out := make(map[uint64][]string, len(poIDs))
	if len(poIDs) == 0 {
		return out, nil
	}
	type row struct {
		POID     uint64 `gorm:"column:po_id"`
		SkuSpecs string `gorm:"column:sku_specs"`
		Qty      int    `gorm:"column:qty"`
	}
	var rows []row
	err := r.db.Model(&model.PurchaseOrderItem{}).
		Scopes(scopeTenant(r.tenantID)).
		Select("po_id, sku_specs, qty").
		Where("po_id IN ? AND cancelled = ? AND sku_specs <> ''", poIDs, false).
		Order("po_id ASC, id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	order := make(map[uint64][]string, len(poIDs)) // first-seen order of specs
	qtyBy := make(map[uint64]map[string]int, len(poIDs))
	for _, row := range rows {
		spec := strings.TrimSpace(row.SkuSpecs)
		if spec == "" {
			continue
		}
		q := row.Qty
		if q <= 0 {
			q = 1
		}
		if _, ok := qtyBy[row.POID]; !ok {
			qtyBy[row.POID] = map[string]int{}
		}
		if _, ok := qtyBy[row.POID][spec]; !ok {
			order[row.POID] = append(order[row.POID], spec)
		}
		qtyBy[row.POID][spec] += q
	}
	for poID, specs := range order {
		lines := make([]string, 0, len(specs))
		for _, spec := range specs {
			q := qtyBy[poID][spec]
			if q > 1 {
				lines = append(lines, fmt.Sprintf("%s x%d", spec, q))
			} else {
				lines = append(lines, spec)
			}
		}
		out[poID] = lines
	}
	return out, nil
}

func (r *PurchaseOrderRepo) NextPoNo() (string, error) {
	prefix := "PO" + time.Now().Format("20060102")
	var last string
	err := r.db.Model(&model.PurchaseOrder{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("po_no LIKE ?", prefix+"%").
		Order("po_no DESC").
		Limit(1).
		Pluck("po_no", &last).Error
	if err != nil {
		return "", err
	}
	seq := 1
	if last != "" && len(last) > len(prefix) {
		var n int
		if _, scanErr := fmt.Sscanf(last[len(prefix):], "%d", &n); scanErr == nil && n >= 0 {
			seq = n + 1
		}
	}
	return fmt.Sprintf("%s%04d", prefix, seq), nil
}
