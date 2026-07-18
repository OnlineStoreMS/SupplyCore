package admin

import (
	"net/http"

	"supplycore/internal/dto"
	"supplycore/internal/pkg/authcontext"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/service"

	"github.com/gin-gonic/gin"
)

type PurchaseExtHandler struct {
	svc *service.PurchaseExtService
}

func NewPurchaseExtHandler(svc *service.PurchaseExtService) *PurchaseExtHandler {
	return &PurchaseExtHandler{svc: svc}
}

func (h *PurchaseExtHandler) ss(c *gin.Context) *service.PurchaseExtService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *PurchaseExtHandler) operator(c *gin.Context) string {
	if claims := authcontext.Claims(c); claims != nil && claims.DisplayName != "" {
		return claims.DisplayName
	}
	return "system"
}

// ---- Accounts ----

func (h *PurchaseExtHandler) ListAccounts(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.ss(c).ListAccounts(c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *PurchaseExtHandler) CreateAccount(c *gin.Context) {
	var in dto.PurchaseAccountDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	in.OperatorName = h.operator(c)
	item, err := h.ss(c).CreateAccount(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PurchaseExtHandler) UpdateAccount(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.PurchaseAccountDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	in.OperatorName = h.operator(c)
	item, err := h.ss(c).UpdateAccount(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) DeleteAccount(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ss(c).DeleteAccount(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// ---- Inbounds ----

func (h *PurchaseExtHandler) ListInbounds(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.ss(c).ListInbounds(c.Query("status"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *PurchaseExtHandler) GetInbound(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).GetInbound(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) CreateInbound(c *gin.Context) {
	var in dto.PurchaseInboundInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreateInbound(&in, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PurchaseExtHandler) ApproveInboundWH(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).ApproveInboundWH(c.Request.Context(), id, h.operator(c), c.GetHeader("Authorization"))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) ApproveInboundFinance(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).ApproveInboundFinance(id, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) VoidInbound(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ss(c).VoidInbound(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// ---- Package receive ----

func (h *PurchaseExtHandler) ListPackageReceives(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.ss(c).ListPackageReceives(c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *PurchaseExtHandler) ScanPackage(c *gin.Context) {
	var in dto.PackageReceiveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).ScanPackage(&in, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PurchaseExtHandler) CreateInboundFromPackage(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).CreateInboundFromPackage(id, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

// ---- Returns ----

func (h *PurchaseExtHandler) ListReturns(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	list, total, err := h.ss(c).ListReturns(c.Query("status"), c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *PurchaseExtHandler) GetReturn(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).GetReturn(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) CreateReturn(c *gin.Context) {
	var in dto.PurchaseReturnInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreateReturn(&in, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *PurchaseExtHandler) ApproveReturn(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).ApproveReturn(id, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) ApproveReturnFinance(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.ss(c).ApproveReturnFinance(id, h.operator(c))
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseExtHandler) VoidReturn(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ss(c).VoidReturn(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, nil)
}

// ---- Suggestions ----

func (h *PurchaseExtHandler) ListSuggestions(c *gin.Context) {
	source := c.Query("source")
	if source == "" {
		source = "stockout"
	}
	list, err := h.ss(c).ListSuggestions(source)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}
