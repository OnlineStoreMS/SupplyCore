package admin

import (
	"net/http"
	"strings"

	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/integrations/productcore"

	"github.com/gin-gonic/gin"
)

type ProductSkuHandler struct {
	pc *productcore.Client
}

func NewProductSkuHandler(pc *productcore.Client) *ProductSkuHandler {
	return &ProductSkuHandler{pc: pc}
}

func (h *ProductSkuHandler) Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.Fail(c, http.StatusBadRequest, "请输入 SKU 编码、规格值或商品关键字")
		return
	}
	page, pageSize := httputil.ParsePage(c)
	auth := c.GetHeader("Authorization")
	list, total, err := h.pc.SearchSkus(c.Request.Context(), auth, keyword, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}
