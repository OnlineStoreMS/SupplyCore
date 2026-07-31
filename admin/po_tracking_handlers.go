package admin

import (
	"net/http"
	"strconv"
	"strings"

	"supplycore/internal/dto"
	"supplycore/internal/pkg/authcontext"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/service"
	"supplycore/internal/storage"

	"github.com/gin-gonic/gin"
)

type POTrackingHandler struct {
	svc     *service.POTrackingService
	storage storage.Storage
}

func NewPOTrackingHandler(svc *service.POTrackingService, store storage.Storage) *POTrackingHandler {
	return &POTrackingHandler{svc: svc, storage: store}
}

func (h *POTrackingHandler) ts(c *gin.Context) *service.POTrackingService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func parsePOID(c *gin.Context) (uint64, error) {
	return httputil.ParseID(c)
}

func (h *POTrackingHandler) ListShipments(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	list, err := h.ts(c).ListShipments(poID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *POTrackingHandler) CreateShipment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	var in dto.ShipmentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ts(c).CreateShipment(poID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *POTrackingHandler) SyncShipmentsFromOrders(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	var in dto.SyncShipmentsFromOrdersInput
	_ = c.ShouldBindJSON(&in)
	if v := c.Query("refSoId"); v != "" {
		if id, perr := strconv.ParseUint(v, 10, 64); perr == nil {
			in.RefSoID = id
		}
	}
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		tok := authcontext.BearerToken(c)
		if tok != "" {
			auth = "Bearer " + tok
		}
	}
	result, err := h.ts(c).SyncShipmentsFromOrders(c.Request.Context(), poID, auth, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *POTrackingHandler) UpdateShipmentStatus(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	shipmentID, err := strconv.ParseUint(c.Param("shipmentId"), 10, 64)
	if err != nil || shipmentID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid shipment id")
		return
	}
	var in dto.ShipmentStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ts(c).UpdateShipmentStatus(poID, shipmentID, in.Status)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *POTrackingHandler) DeleteShipment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	shipmentID, err := strconv.ParseUint(c.Param("shipmentId"), 10, 64)
	if err != nil || shipmentID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid shipment id")
		return
	}
	if err := h.ts(c).DeleteShipment(poID, shipmentID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *POTrackingHandler) ListPayments(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	list, err := h.ts(c).ListPayments(poID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *POTrackingHandler) CreatePayment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	var in dto.PaymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ts(c).CreatePayment(poID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *POTrackingHandler) UpdatePayment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	paymentID, err := strconv.ParseUint(c.Param("paymentId"), 10, 64)
	if err != nil || paymentID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid payment id")
		return
	}
	var in dto.PaymentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ts(c).UpdatePayment(poID, paymentID, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *POTrackingHandler) DeletePayment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	paymentID, err := strconv.ParseUint(c.Param("paymentId"), 10, 64)
	if err != nil || paymentID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid payment id")
		return
	}
	if err := h.ts(c).DeletePayment(poID, paymentID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *POTrackingHandler) ListAttachments(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	list, err := h.ts(c).ListAttachments(poID)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *POTrackingHandler) CreateAttachment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	var in dto.AttachmentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ts(c).CreateAttachment(poID, authcontext.UserID(c), &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *POTrackingHandler) DeleteAttachment(c *gin.Context) {
	poID, err := parsePOID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid po id")
		return
	}
	attachmentID, err := strconv.ParseUint(c.Param("attachmentId"), 10, 64)
	if err != nil || attachmentID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid attachment id")
		return
	}
	if err := h.ts(c).DeleteAttachment(poID, attachmentID); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *POTrackingHandler) Upload(c *gin.Context) {
	if h.storage == nil {
		response.Fail(c, http.StatusInternalServerError, "存储未配置")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "请选择文件")
		return
	}
	subdir := strings.Trim(c.PostForm("subdir"), "/")
	if subdir == "" {
		subdir = "po"
	}
	url, err := h.storage.Upload(file, subdir)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, dto.UploadResult{URL: url, FileName: file.Filename})
}
