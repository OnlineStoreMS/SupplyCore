package router

import (
	"path/filepath"

	"supplycore/admin"
	adminmw "supplycore/admin/middleware"
	"supplycore/internal/config"
	jwtmgr "supplycore/internal/pkg/jwt"
	"supplycore/internal/integrations/productcore"
	"supplycore/internal/repo"
	"supplycore/internal/service"
	"supplycore/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
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
	trackSvc := service.NewPOTrackingService(repos)
	soSvc := service.NewSalesOrderService(repos)
	srcSvc := service.NewSourcingService(repos, poSvc)
	extSvc := service.NewPurchaseExtService(repos)
	pcClient := productcore.NewClient(cfg.Integrations.ProductCoreAPIURL)
	supplierH := admin.NewSupplierHandler(supplierSvc)
	offerH := admin.NewOfferHandler(offerSvc)
	poH := admin.NewPurchaseOrderHandler(poSvc)
	trackH := admin.NewPOTrackingHandler(trackSvc, store)
	soH := admin.NewSalesOrderHandler(soSvc, srcSvc)
	skuH := admin.NewProductSkuHandler(pcClient)
	extH := admin.NewPurchaseExtHandler(extSvc)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "supplycore"})
	})

	v1 := r.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	jwtMgr := jwtmgr.NewManager(cfg.Auth.JWTSecret)
	adminGroup.Use(adminmw.AdminAuth(&cfg.Auth, jwtMgr))
	admin.RegisterRoutes(adminGroup, supplierH, offerH, poH, trackH, soH, skuH, extH)

	return r
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
