package repo

import (
	"fmt"
	"strings"
	"time"

	"supplycore/internal/model"

	"gorm.io/gorm"
)

type ShipmentRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewShipmentRepo(db *gorm.DB) *ShipmentRepo {
	return &ShipmentRepo{db: db}
}

func (r *ShipmentRepo) ForTenant(tenantID uint64) *ShipmentRepo {
	return &ShipmentRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *ShipmentRepo) ListByPO(poID uint64) ([]model.PurchaseShipment, error) {
	var list []model.PurchaseShipment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ?", poID).
		Preload("Items").
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *ShipmentRepo) GetByID(poID, id uint64) (*model.PurchaseShipment, error) {
	var s model.PurchaseShipment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND id = ?", poID, id).
		Preload("Items").
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShipmentRepo) FindByTrackingNo(poID uint64, trackingNo string) (*model.PurchaseShipment, error) {
	trackingNo = strings.TrimSpace(trackingNo)
	if trackingNo == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var s model.PurchaseShipment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND tracking_no = ?", poID, trackingNo).
		Preload("Items").
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ShipmentRepo) Create(s *model.PurchaseShipment, items []model.PurchaseShipmentItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		s.TenantID = r.tenantID
		if err := tx.Create(s).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = r.tenantID
			items[i].ShipmentID = s.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		s.Items = items
		return nil
	})
}

// AddItems 向已有物流追加明细（快递助手合单发货：多销售单共用同一运单号）。
func (r *ShipmentRepo) AddItems(shipmentID uint64, items []model.PurchaseShipmentItem) error {
	if shipmentID == 0 || len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].TenantID = r.tenantID
		items[i].ShipmentID = shipmentID
	}
	return r.db.Create(&items).Error
}

func (r *ShipmentRepo) Save(s *model.PurchaseShipment) error {
	return r.db.Save(s).Error
}

func (r *ShipmentRepo) Delete(poID, id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Scopes(scopeTenant(r.tenantID)).
			Where("shipment_id = ?", id).
			Delete(&model.PurchaseShipmentItem{}).Error; err != nil {
			return err
		}
		return tx.Scopes(scopeTenant(r.tenantID)).
			Where("po_id = ? AND id = ?", poID, id).
			Delete(&model.PurchaseShipment{}).Error
	})
}

// RemapItemPOItemIDs 合并重建明细后，把物流明细的 po_item_id 指到新 ID。
func (r *ShipmentRepo) RemapItemPOItemIDs(oldToNew map[uint64]uint64) error {
	if len(oldToNew) == 0 {
		return nil
	}
	now := time.Now()
	for oldID, newID := range oldToNew {
		if oldID == 0 || newID == 0 || oldID == newID {
			continue
		}
		if err := r.db.Model(&model.PurchaseShipmentItem{}).
			Scopes(scopeTenant(r.tenantID)).
			Where("po_item_id = ?", oldID).
			Updates(map[string]any{"po_item_id": newID, "updated_at": now}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *ShipmentRepo) NextShipmentNo() (string, error) {
	// 取当日最大单号 +1；COUNT+1 在删单留洞时会撞唯一索引 idx_shipment_tenant_no
	prefix := "SH" + time.Now().Format("20060102")
	var last string
	if err := r.db.Model(&model.PurchaseShipment{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("shipment_no LIKE ?", prefix+"%").
		Order("shipment_no DESC").
		Limit(1).
		Pluck("shipment_no", &last).Error; err != nil {
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

type PaymentRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewPaymentRepo(db *gorm.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) ForTenant(tenantID uint64) *PaymentRepo {
	return &PaymentRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *PaymentRepo) ListByPO(poID uint64) ([]model.PurchasePayment, error) {
	var list []model.PurchasePayment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ?", poID).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *PaymentRepo) GetByID(poID, id uint64) (*model.PurchasePayment, error) {
	var p model.PurchasePayment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND id = ?", poID, id).
		First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) Create(p *model.PurchasePayment) error {
	p.TenantID = r.tenantID
	return r.db.Create(p).Error
}

func (r *PaymentRepo) Save(p *model.PurchasePayment) error {
	return r.db.Save(p).Error
}

func (r *PaymentRepo) Delete(poID, id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND id = ?", poID, id).
		Delete(&model.PurchasePayment{}).Error
}

func (r *PaymentRepo) SumPaid(poID uint64) (float64, error) {
	var sum float64
	err := r.db.Model(&model.PurchasePayment{}).
		Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND pay_status = ?", poID, model.POPayStatusPaid).
		Select("COALESCE(SUM(pay_amount), 0)").
		Scan(&sum).Error
	return sum, err
}

type AttachmentRepo struct {
	db       *gorm.DB
	tenantID uint64
}

func NewAttachmentRepo(db *gorm.DB) *AttachmentRepo {
	return &AttachmentRepo{db: db}
}

func (r *AttachmentRepo) ForTenant(tenantID uint64) *AttachmentRepo {
	return &AttachmentRepo{db: r.db, tenantID: NormalizeTenantID(tenantID)}
}

func (r *AttachmentRepo) ListByPO(poID uint64) ([]model.PurchaseAttachment, error) {
	var list []model.PurchaseAttachment
	err := r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ?", poID).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

func (r *AttachmentRepo) Create(a *model.PurchaseAttachment) error {
	a.TenantID = r.tenantID
	return r.db.Create(a).Error
}

func (r *AttachmentRepo) Delete(poID, id uint64) error {
	return r.db.Scopes(scopeTenant(r.tenantID)).
		Where("po_id = ? AND id = ?", poID, id).
		Delete(&model.PurchaseAttachment{}).Error
}
