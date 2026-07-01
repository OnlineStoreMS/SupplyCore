package admin

import (
	"net/http"

	"supplycore/internal/dto"
	"supplycore/internal/pkg/authcontext"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/repo"
	"supplycore/internal/service"

	"github.com/gin-gonic/gin"
)

type SalesOrderHandler struct {
	soSvc *service.SalesOrderService
	srcSvc *service.SourcingService
}

func NewSalesOrderHandler(soSvc *service.SalesOrderService, srcSvc *service.SourcingService) *SalesOrderHandler {
	return &SalesOrderHandler{soSvc: soSvc, srcSvc: srcSvc}
}

func (h *SalesOrderHandler) ss(c *gin.Context) *service.SalesOrderService {
	return h.soSvc.ForTenant(authcontext.TenantID(c))
}

func (h *SalesOrderHandler) src(c *gin.Context) *service.SourcingService {
	return h.srcSvc.ForTenant(authcontext.TenantID(c))
}

func (h *SalesOrderHandler) List(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.ss(c).List(repo.SOListFilter{
		Status: c.Query("status"), Keyword: c.Query("keyword"),
		Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *SalesOrderHandler) Get(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).Get(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SalesOrderHandler) Create(c *gin.Context) {
	var in dto.SalesOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).Create(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SalesOrderHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.SalesOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).Update(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SalesOrderHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ss(c).Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *SalesOrderHandler) Confirm(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).Confirm(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SalesOrderHandler) EvaluateSourcing(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	resp, err := h.src(c).Evaluate(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *SalesOrderHandler) CreatePurchaseOrders(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.SourcingCreatePOInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := authcontext.Claims(c)
	buyerName := ""
	if claims != nil {
		buyerName = claims.DisplayName
	}
	resp, err := h.src(c).CreatePOsFromSO(id, &in, authcontext.UserID(c), buyerName)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, resp)
}

func (h *SalesOrderHandler) DropshipPurchaseOrder(c *gin.Context) {
	var in dto.SourcingCreatePOInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := authcontext.Claims(c)
	buyerName := ""
	if claims != nil {
		buyerName = claims.DisplayName
	}
	resp, err := h.src(c).CreateDropshipPO(&in, authcontext.UserID(c), buyerName)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, resp)
}
