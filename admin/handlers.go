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

type SupplierHandler struct {
	svc *service.SupplierService
}

func NewSupplierHandler(svc *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

func (h *SupplierHandler) ss(c *gin.Context) *service.SupplierService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *SupplierHandler) List(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	categoryID, _ := strconv.ParseUint(c.Query("categoryId"), 10, 64)
	list, total, err := h.ss(c).List(c.Query("keyword"), categoryID, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *SupplierHandler) Get(c *gin.Context) {
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

func (h *SupplierHandler) Create(c *gin.Context) {
	var in dto.SupplierDTO
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

func (h *SupplierHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.SupplierDTO
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

func (h *SupplierHandler) Delete(c *gin.Context) {
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

func (h *SupplierHandler) ListCategories(c *gin.Context) {
	list, err := h.ss(c).ListCategories()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *SupplierHandler) CreateCategory(c *gin.Context) {
	var in dto.SupplierCategoryDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreateCategory(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SupplierHandler) UpdateCategory(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.SupplierCategoryDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).UpdateCategory(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SupplierHandler) DeleteCategory(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.ss(c).DeleteCategory(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *SupplierHandler) ListAddresses(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	list, err := h.ss(c).ListAddresses(supplierID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *SupplierHandler) CreateAddress(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	var in dto.SupplierAddressDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreateAddress(supplierID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SupplierHandler) UpdateAddress(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	addressID, err := strconv.ParseUint(c.Param("addressId"), 10, 64)
	if err != nil || addressID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid address id")
		return
	}
	var in dto.SupplierAddressDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).UpdateAddress(supplierID, addressID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SupplierHandler) DeleteAddress(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	addressID, err := strconv.ParseUint(c.Param("addressId"), 10, 64)
	if err != nil || addressID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid address id")
		return
	}
	if err := h.ss(c).DeleteAddress(supplierID, addressID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *SupplierHandler) ListPaymentAccounts(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	list, err := h.ss(c).ListPaymentAccounts(supplierID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *SupplierHandler) CreatePaymentAccount(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	var in dto.SupplierPaymentAccountDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreatePaymentAccount(supplierID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SupplierHandler) UpdatePaymentAccount(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	accountID, err := strconv.ParseUint(c.Param("accountId"), 10, 64)
	if err != nil || accountID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid account id")
		return
	}
	var in dto.SupplierPaymentAccountDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).UpdatePaymentAccount(supplierID, accountID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SupplierHandler) DeletePaymentAccount(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	accountID, err := strconv.ParseUint(c.Param("accountId"), 10, 64)
	if err != nil || accountID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid account id")
		return
	}
	if err := h.ss(c).DeletePaymentAccount(supplierID, accountID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *SupplierHandler) ListPaymentQRs(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	list, err := h.ss(c).ListPaymentQRs(supplierID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *SupplierHandler) CreatePaymentQR(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	var in dto.SupplierPaymentQRDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).CreatePaymentQR(supplierID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *SupplierHandler) UpdatePaymentQR(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	qrID, err := strconv.ParseUint(c.Param("qrId"), 10, 64)
	if err != nil || qrID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid qr id")
		return
	}
	var in dto.SupplierPaymentQRDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ss(c).UpdatePaymentQR(supplierID, qrID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SupplierHandler) DeletePaymentQR(c *gin.Context) {
	supplierID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid supplier id")
		return
	}
	qrID, err := strconv.ParseUint(c.Param("qrId"), 10, 64)
	if err != nil || qrID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid qr id")
		return
	}
	if err := h.ss(c).DeletePaymentQR(supplierID, qrID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

type OfferHandler struct {
	svc *service.OfferService
}

func NewOfferHandler(svc *service.OfferService) *OfferHandler {
	return &OfferHandler{svc: svc}
}

func (h *OfferHandler) os(c *gin.Context) *service.OfferService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *OfferHandler) List(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	skuID, _ := strconv.ParseUint(c.Query("skuId"), 10, 64)
	supplierID, _ := strconv.ParseUint(c.Query("supplierId"), 10, 64)
	list, total, err := h.os(c).List(repo.OfferListFilter{
		SkuID: skuID, SupplierID: supplierID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *OfferHandler) Get(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.os(c).Get(id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *OfferHandler) Create(c *gin.Context) {
	var in dto.SkuOfferDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.os(c).Create(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *OfferHandler) Update(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.SkuOfferDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.os(c).Update(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *OfferHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.os(c).Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *OfferHandler) SupplyOptions(c *gin.Context) {
	skuID, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid sku id")
		return
	}
	dropshipOnly := c.Query("dropshipOnly") == "1" || c.Query("dropshipOnly") == "true"
	resp, err := h.os(c).SupplyOptions(skuID, dropshipOnly)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, resp)
}
