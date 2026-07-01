package admin

import (
	"net/http"
	"strconv"

	"supplycore/internal/dto"
	"supplycore/internal/pkg/authcontext"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/repo"
	"supplycore/internal/service"

	"github.com/gin-gonic/gin"
)

type PurchaseOrderHandler struct {
	svc *service.PurchaseOrderService
}

func NewPurchaseOrderHandler(svc *service.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{svc: svc}
}

func (h *PurchaseOrderHandler) ps(c *gin.Context) *service.PurchaseOrderService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	supplierID, _ := strconv.ParseUint(c.Query("supplierId"), 10, 64)
	refSoID, _ := strconv.ParseUint(c.Query("refSoId"), 10, 64)
	list, total, err := h.ps(c).List(repo.POListFilter{
		Status: c.Query("status"), SupplierID: supplierID,
		RefSoID: refSoID, RefTraceID: c.Query("refTraceId"),
		Keyword: c.Query("keyword"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *PurchaseOrderHandler) Get(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ps(c).Get(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var in dto.PurchaseOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	claims := authcontext.Claims(c)
	buyerID := authcontext.UserID(c)
	buyerName := ""
	if claims != nil {
		buyerName = claims.DisplayName
	}
	item, err := h.ps(c).Create(&in, buyerID, buyerName)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PurchaseOrderHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.PurchaseOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ps(c).Update(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ps(c).Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *PurchaseOrderHandler) Submit(c *gin.Context) {
	h.doAction(c, func(id uint64) (*dto.PurchaseOrderDetail, error) {
		return h.ps(c).Submit(id)
	})
}

func (h *PurchaseOrderHandler) MarkPaid(c *gin.Context) {
	h.doAction(c, func(id uint64) (*dto.PurchaseOrderDetail, error) {
		return h.ps(c).MarkPaid(id)
	})
}

func (h *PurchaseOrderHandler) Complete(c *gin.Context) {
	h.doAction(c, func(id uint64) (*dto.PurchaseOrderDetail, error) {
		return h.ps(c).Complete(id)
	})
}

func (h *PurchaseOrderHandler) Cancel(c *gin.Context) {
	h.doAction(c, func(id uint64) (*dto.PurchaseOrderDetail, error) {
		return h.ps(c).Cancel(id)
	})
}

func (h *PurchaseOrderHandler) doAction(c *gin.Context, fn func(uint64) (*dto.PurchaseOrderDetail, error)) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := fn(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}
