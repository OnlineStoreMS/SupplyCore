package router

import (
	"path/filepath"

	"supplycore/admin"
	adminmw "supplycore/admin/middleware"
	"supplycore/internal/config"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/integrations/productcore"
	"supplycore/internal/integrations/warehousecore"
	jwtmgr "supplycore/internal/pkg/jwt"
	"supplycore/internal/repo"
	"supplycore/internal/scheduler"
	"supplycore/internal/service"
	"supplycore/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) (*gin.Engine, *scheduler.SettlementScheduler) {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg))

	if cfg.Storage.Driver == "local" || cfg.Storage.Driver == "" {
		uploadDir := filepath.Join(cfg.Storage.LocalPath, cfg.Storage.Prefix)
		r.Static("/uploads", uploadDir)
	}

	store, err := storage.New(&cfg.Storage)
	if err != nil {
		panic(err)
	}

	repos := repo.New(db)
	supplierSvc := service.NewSupplierService(repos)
	offerSvc := service.NewOfferService(repos)
	poSvc := service.NewPurchaseOrderService(repos)
	pcClient := productcore.NewClient(cfg.Integrations.ProductCoreAPIURL)
	wcClient := warehousecore.NewClient(cfg.Integrations.WarehouseCoreAPIURL)
	ocClient := ordercore.NewClient(cfg.Integrations.OrderCoreAPIURL)
	trackSvc := service.NewPOTrackingService(repos, ocClient)
	extSvc := service.NewPurchaseExtService(repos, wcClient)
	dashSvc := service.NewDashboardService(repos)
	supplierH := admin.NewSupplierHandler(supplierSvc)
	offerH := admin.NewOfferHandler(offerSvc)
	poH := admin.NewPurchaseOrderHandler(poSvc, ocClient)
	trackH := admin.NewPOTrackingHandler(trackSvc, store)
	skuH := admin.NewProductSkuHandler(pcClient)
	whH := admin.NewWarehouseHandler(wcClient)
	extH := admin.NewPurchaseExtHandler(extSvc)
	dashH := admin.NewDashboardHandler(dashSvc)
	orderH := admin.NewOrderHandler(ocClient)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "supplycore"})
	})

	v1 := r.Group("/api/v1")
	photoH := admin.NewPhotoUploadHandler(store)
	mobile := v1.Group("/mobile")
	{
		mobile.GET("/photo-upload/:token", photoH.MobileGet)
		mobile.POST("/photo-upload/:token", photoH.MobileUpload)
	}

	adminGroup := v1.Group("/admin")
	jwtMgr := jwtmgr.NewManager(cfg.Auth.JWTSecret)
	adminGroup.Use(adminmw.AdminAuth(&cfg.Auth, jwtMgr))
	adminGroup.POST("/photo-upload-sessions", photoH.CreateSession)
	adminGroup.GET("/photo-upload-sessions/:token", photoH.GetSession)
	admin.RegisterRoutes(adminGroup, supplierH, offerH, poH, trackH, skuH, whH, extH, dashH, orderH)

	settlementSvc := service.NewSettlementMergeService(repos, poSvc, ocClient, jwtMgr)
	sched := scheduler.NewSettlementScheduler(settlementSvc)

	return r, sched
}

func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORS.AllowOrigins
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := origin == ""
		for _, o := range origins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}
		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
