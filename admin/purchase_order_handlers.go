package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/pkg/authcontext"
	"supplycore/internal/pkg/httputil"
	"supplycore/internal/pkg/response"
	"supplycore/internal/repo"
	"supplycore/internal/service"

	"github.com/gin-gonic/gin"
)

type PurchaseOrderHandler struct {
	svc *service.PurchaseOrderService
	oc  *ordercore.Client
}

func NewPurchaseOrderHandler(svc *service.PurchaseOrderService, oc *ordercore.Client) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{svc: svc, oc: oc}
}

func (h *PurchaseOrderHandler) ps(c *gin.Context) *service.PurchaseOrderService {
	return h.svc.ForTenant(authcontext.TenantID(c))
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	page, pageSize := httputil.ParsePage(c)
	supplierID, _ := strconv.ParseUint(c.Query("supplierId"), 10, 64)
	refSoID, _ := strconv.ParseUint(c.Query("refSoId"), 10, 64)
	status := c.Query("status")
	if status == "in_transit" {
		status = "shipped"
	}
	awaitingLogistics := c.Query("awaitingLogistics") == "1" ||
		strings.EqualFold(c.Query("awaitingLogistics"), "true")
	list, total, err := h.ps(c).List(repo.POListFilter{
		Status: status, Statuses: splitCSV(c.Query("statuses")),
		PayStatuses: splitCSV(c.Query("payStatus")), ExcludeStatuses: splitCSV(c.Query("excludeStatuses")),
		AwaitingLogistics: awaitingLogistics,
		FulfillmentType:   c.Query("fulfillmentType"),
		SupplierID:        supplierID, RefSoID: refSoID, RefTraceID: c.Query("refTraceId"),
		Keyword: c.Query("keyword"), SortBy: c.Query("sortBy"), SortOrder: c.Query("sortOrder"),
		CreatedAtStart: parsePOCreatedAtStart(c.Query("createdAtStart")),
		CreatedAtEnd:   parsePOCreatedAtEndExclusive(c.Query("createdAtEnd")),
		OrderedAtStart: parsePOCreatedAtStart(c.Query("orderedAtStart")),
		OrderedAtEnd:   parsePOCreatedAtEndExclusive(c.Query("orderedAtEnd")),
		Page: page, PageSize: pageSize,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, response.PageResult(list, total, page, pageSize))
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// 兼容旧筛选值「运输中」
		if p == "in_transit" {
			p = "shipped"
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// parsePOCreatedAtStart 解析创建日起始（含当日 00:00）。
func parsePOCreatedAtStart(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			if layout == "2006-01-02" {
				day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
				return &day
			}
			return &t
		}
	}
	return nil
}

// parsePOCreatedAtEndExclusive 解析创建日截止：日期按「含当日」转为次日 00:00，查询用 created_at < end。
func parsePOCreatedAtEndExclusive(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			if layout == "2006-01-02" {
				next := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, 1)
				return &next
			}
			// 带时分秒：按开区间上界 < end+1s 近似「含该秒」，这里直接用 < t+1ns 不方便；统一加 1 秒
			end := t.Add(time.Second)
			return &end
		}
	}
	return nil
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

func (h *PurchaseOrderHandler) UpdateItemPrices(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var in dto.UpdatePOItemPricesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ps(c).UpdateItemPrices(id, &in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *PurchaseOrderHandler) SyncPurchasePrices(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if h.oc == nil {
		response.Fail(c, http.StatusBadGateway, "OrderCore 未配置")
		return
	}
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		tok := authcontext.BearerToken(c)
		if tok != "" {
			auth = "Bearer " + tok
		}
	}
	n, err := h.ps(c).SyncPurchasePricesForPOIfConfigured(c.Request.Context(), h.oc, auth, id)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	detail, gerr := h.ps(c).Get(id)
	if gerr != nil {
		httputil.HandleServiceError(c, gerr)
		return
	}
	response.OK(c, gin.H{"updated": n > 0, "synced": n, "purchaseOrder": detail})
}

func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
	id, err := httputil.ParseID(c)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	detail, getErr := h.ps(c).Get(id)
	if getErr != nil {
		httputil.HandleServiceError(c, getErr)
		return
	}
	poNo := ""
	fulfillment := ""
	if detail != nil {
		poNo = strings.TrimSpace(detail.PoNo)
		fulfillment = detail.FulfillmentType
	}
	if err := h.ps(c).Delete(id); err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	// 代发单删除后清空订单中心上的采购单号，避免同步误判「已有单」而不重建
	if h.oc != nil && fulfillment == "dropship" && poNo != "" {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			tok := authcontext.BearerToken(c)
			if tok != "" {
				auth = "Bearer " + tok
			}
		}
		if _, rerr := h.oc.RelinkPurchaseOrder(c.Request.Context(), auth, []string{poNo}, ""); rerr != nil {
			// 单据已删，解绑失败只记日志式返回附加信息
			response.OK(c, gin.H{"deleted": true, "unlinkWarning": rerr.Error()})
			return
		}
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *PurchaseOrderHandler) Merge(c *gin.Context) {
	var in dto.MergePurchaseOrdersInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.ps(c).Merge(&in)
	if err != nil {
		httputil.HandleServiceError(c, err)
		return
	}
	if h.oc != nil && result != nil && len(result.MergedFromPoNos) > 0 && result.PoNo != "" {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			tok := authcontext.BearerToken(c)
			if tok != "" {
				auth = "Bearer " + tok
			}
		}
		n, rerr := h.oc.RelinkPurchaseOrder(c.Request.Context(), auth, result.MergedFromPoNos, result.PoNo)
		if rerr != nil {
			response.Fail(c, http.StatusBadRequest, "代发单已合并，但回写订单中心采购单号失败: "+rerr.Error())
			return
		}
		result.Relinked = n
		if result.ID > 0 {
			if _, serr := h.ps(c).SyncPurchasePricesForPOIfConfigured(c.Request.Context(), h.oc, auth, result.ID); serr != nil {
				// 合并已成功，价格同步失败仅记日志由前端可再改价
				_ = serr
			}
		}
	}
	response.OK(c, result)
}

func (h *PurchaseOrderHandler) DetachSalesOrder(c *gin.Context) {
	var in dto.DetachSalesOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.ps(c).DetachSalesOrder(&in)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			httputil.HandleServiceError(c, err)
			return
		}
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	unlinkWarning := ""
	if h.oc != nil {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			tok := authcontext.BearerToken(c)
			if tok != "" {
				auth = "Bearer " + tok
			}
		}
		var orderIDs []uint64
		var orderNos []string
		if in.SoID > 0 {
			orderIDs = []uint64{in.SoID}
		}
		if no := strings.TrimSpace(in.OrderNo); no != "" {
			orderNos = []string{no}
		}
		remark := strings.TrimSpace(in.Reason)
		if remark == "" {
			remark = "供应链解绑代发销售单"
		}
		if _, rerr := h.oc.UnlinkDropshipPO(c.Request.Context(), auth, orderIDs, orderNos, true, remark); rerr != nil {
			unlinkWarning = rerr.Error()
		}
	}
	response.OK(c, gin.H{"purchaseOrder": item, "unlinkWarning": unlinkWarning})
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
