package admin

import (
	"net/http"
	"strings"

	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	oc *ordercore.Client
}

func NewOrderHandler(oc *ordercore.Client) *OrderHandler {
	return &OrderHandler{oc: oc}
}

func (h *OrderHandler) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.Fail(c, http.StatusBadRequest, "请输入订单号搜索")
		return
	}
	page, pageSize := httputil.ParsePage(c)
	auth := c.GetHeader("Authorization")
	list, total, err := h.oc.SearchOrders(c.Request.Context(), auth, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}
