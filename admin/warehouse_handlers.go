package admin

import (
	"net/http"
	"strconv"

	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/integrations/warehousecore"

	"github.com/gin-gonic/gin"
)

type WarehouseHandler struct {
	wc *warehousecore.Client
}

func NewWarehouseHandler(wc *warehousecore.Client) *WarehouseHandler {
	return &WarehouseHandler{wc: wc}
}

func (h *WarehouseHandler) List(c *gin.Context) {
	if h.wc == nil || !h.wc.Enabled() {
		response.Fail(c, http.StatusServiceUnavailable, "WarehouseCore 未配置")
		return
	}
	page, pageSize := httputil.ParsePage(c)
	auth := c.GetHeader("Authorization")
	list, total, err := h.wc.ListWarehouses(c.Request.Context(), auth, c.Query("keyword"), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func (h *WarehouseHandler) ListLocations(c *gin.Context) {
	if h.wc == nil || !h.wc.Enabled() {
		response.Fail(c, http.StatusServiceUnavailable, "WarehouseCore 未配置")
		return
	}
	whID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || whID == 0 {
		response.Fail(c, http.StatusBadRequest, "invalid warehouse id")
		return
	}
	page, pageSize := httputil.ParsePage(c)
	auth := c.GetHeader("Authorization")
	list, total, err := h.wc.ListLocations(c.Request.Context(), auth, whID, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}
