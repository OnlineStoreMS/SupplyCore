package database

import (
	"fmt"
	"os"
	"path/filepath"

	"supplycore/internal/config"
	"supplycore/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.PostgresDSN)
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.SQLitePath), 0o755); err != nil {
			return nil, err
		}
		dialector = sqlite.Open(cfg.SQLitePath)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Supplier{},
		&model.SupplierAddress{},
		&model.SkuSupplierOffer{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.PurchaseShipment{},
		&model.PurchaseShipmentItem{},
		&model.PurchasePayment{},
		&model.PurchaseAttachment{},
		&model.PurchaseAccount{},
		&model.PurchaseInbound{},
		&model.PurchaseInboundItem{},
		&model.PackageReceiveRecord{},
		&model.PurchaseReturn{},
		&model.PurchaseReturnItem{},
	); err != nil {
		return err
	}
	return ensureIndexes(db)
}

func ensureIndexes(db *gorm.DB) error {
	switch db.Dialector.Name() {
	case "postgres":
		return db.Exec(`
			CREATE UNIQUE INDEX IF NOT EXISTS idx_suppliers_tenant_code ON suppliers (tenant_id, code);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_offers_tenant_sku_supplier_addr ON sku_supplier_offers (tenant_id, sku_id, supplier_id, ship_from_address_id);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_po_tenant_no ON purchase_orders (tenant_id, po_no);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_shipment_tenant_no ON purchase_shipments (tenant_id, shipment_no);
			CREATE INDEX IF NOT EXISTS idx_po_ref_so ON purchase_orders (tenant_id, ref_so_id);
			CREATE INDEX IF NOT EXISTS idx_po_ref_trace ON purchase_orders (tenant_id, ref_trace_id);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_inbound_tenant_no ON purchase_inbounds (tenant_id, inbound_no);
			CREATE UNIQUE INDEX IF NOT EXISTS idx_return_tenant_no ON purchase_returns (tenant_id, return_no);
			CREATE INDEX IF NOT EXISTS idx_pkg_recv_tracking ON package_receive_records (tenant_id, tracking_no);
		`).Error
	default:
		return nil
	}
}
