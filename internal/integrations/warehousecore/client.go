package warehousecore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != ""
}

type Warehouse struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    int8   `json:"status"`
	IsDefault int8   `json:"isDefault"`
}

type Location struct {
	ID          uint64 `json:"id"`
	WarehouseID uint64 `json:"warehouseId"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Status      int8   `json:"status"`
}

type PimMapping struct {
	ID         uint64 `json:"id"`
	InvSkuID   uint64 `json:"invSkuId"`
	PimSkuID   uint64 `json:"pimSkuId"`
	PimSkuCode string `json:"pimSkuCode"`
}

type SkuByCode struct {
	ID      uint64 `json:"id"`
	SkuCode string `json:"skuCode"`
}

type PurchaseInboundItem struct {
	InvSkuID   uint64  `json:"invSkuId"`
	Qty        float64 `json:"qty"`
	Cost       float64 `json:"cost"`
	LocationID uint64  `json:"locationId,omitempty"`
	Remark     string  `json:"remark,omitempty"`
}

type PurchaseInboundRequest struct {
	WarehouseID uint64                `json:"warehouseId"`
	LocationID  uint64                `json:"locationId"`
	RefDocType  string                `json:"refDocType"`
	RefDocID    uint64                `json:"refDocId"`
	RefDocNo    string                `json:"refDocNo"`
	Remark      string                `json:"remark"`
	Items       []PurchaseInboundItem `json:"items"`
}

type pagePayload struct {
	List     json.RawMessage `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type apiBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) ListWarehouses(ctx context.Context, bearerToken, keyword string, page, pageSize int) ([]Warehouse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	q := url.Values{}
	if keyword != "" {
		q.Set("keyword", keyword)
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	reqURL := c.baseURL + "/api/v1/admin/warehouses?" + q.Encode()
	body, err := c.doGET(ctx, bearerToken, reqURL)
	if err != nil {
		return nil, 0, err
	}
	var pageData pagePayload
	if err := json.Unmarshal(body, &pageData); err != nil {
		return nil, 0, fmt.Errorf("warehousecore page decode: %w", err)
	}
	var list []Warehouse
	if len(pageData.List) > 0 && string(pageData.List) != "null" {
		if err := json.Unmarshal(pageData.List, &list); err != nil {
			return nil, 0, fmt.Errorf("warehousecore list decode: %w", err)
		}
	}
	if list == nil {
		list = []Warehouse{}
	}
	return list, pageData.Total, nil
}

func (c *Client) ListLocations(ctx context.Context, bearerToken string, warehouseID uint64, page, pageSize int) ([]Location, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	q := url.Values{}
	q.Set("warehouseId", strconv.FormatUint(warehouseID, 10))
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	reqURL := c.baseURL + "/api/v1/admin/locations?" + q.Encode()
	body, err := c.doGET(ctx, bearerToken, reqURL)
	if err != nil {
		return nil, 0, err
	}
	var pageData pagePayload
	if err := json.Unmarshal(body, &pageData); err != nil {
		return nil, 0, fmt.Errorf("warehousecore page decode: %w", err)
	}
	var list []Location
	if len(pageData.List) > 0 && string(pageData.List) != "null" {
		if err := json.Unmarshal(pageData.List, &list); err != nil {
			return nil, 0, fmt.Errorf("warehousecore list decode: %w", err)
		}
	}
	if list == nil {
		list = []Location{}
	}
	return list, pageData.Total, nil
}

func (c *Client) ResolveInvSkuID(ctx context.Context, bearerToken string, pimSkuID uint64, skuCode string) (uint64, error) {
	if pimSkuID > 0 {
		invID, err := c.lookupPimMapping(ctx, bearerToken, pimSkuID)
		if err == nil && invID > 0 {
			return invID, nil
		}
	}
	code := strings.TrimSpace(skuCode)
	if code != "" {
		sku, err := c.GetSkuByCode(ctx, bearerToken, code)
		if err != nil {
			return 0, err
		}
		if sku != nil && sku.ID > 0 {
			return sku.ID, nil
		}
	}
	if pimSkuID > 0 && code != "" {
		return 0, fmt.Errorf("未找到 SKU 映射: pimSkuId=%d, skuCode=%s；请在 WarehouseCore 配置 PIM 映射或维护 SKU 编码", pimSkuID, code)
	}
	if pimSkuID > 0 {
		return 0, fmt.Errorf("未找到 SKU 映射: pimSkuId=%d；请在 WarehouseCore 配置 PIM 映射", pimSkuID)
	}
	return 0, fmt.Errorf("未找到 SKU 映射: skuCode=%s", code)
}

func (c *Client) lookupPimMapping(ctx context.Context, bearerToken string, pimSkuID uint64) (uint64, error) {
	q := url.Values{}
	q.Set("pimSkuId", strconv.FormatUint(pimSkuID, 10))
	q.Set("page", "1")
	q.Set("pageSize", "20")
	reqURL := c.baseURL + "/api/v1/admin/pim-mappings?" + q.Encode()
	body, err := c.doGET(ctx, bearerToken, reqURL)
	if err != nil {
		return 0, err
	}
	var pageData pagePayload
	if err := json.Unmarshal(body, &pageData); err != nil {
		return 0, fmt.Errorf("warehousecore pim-mappings decode: %w", err)
	}
	var list []PimMapping
	if len(pageData.List) > 0 && string(pageData.List) != "null" {
		if err := json.Unmarshal(pageData.List, &list); err != nil {
			return 0, fmt.Errorf("warehousecore pim list decode: %w", err)
		}
	}
	for _, m := range list {
		if m.PimSkuID == pimSkuID && m.InvSkuID > 0 {
			return m.InvSkuID, nil
		}
	}
	// Server may ignore pimSkuId filter; scan first page for a match.
	for page := 1; page <= 5; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("pageSize", "100")
		reqURL := c.baseURL + "/api/v1/admin/pim-mappings?" + q.Encode()
		body, err := c.doGET(ctx, bearerToken, reqURL)
		if err != nil {
			return 0, err
		}
		var pd pagePayload
		if err := json.Unmarshal(body, &pd); err != nil {
			return 0, err
		}
		var batch []PimMapping
		if len(pd.List) > 0 && string(pd.List) != "null" {
			if err := json.Unmarshal(pd.List, &batch); err != nil {
				return 0, err
			}
		}
		if len(batch) == 0 {
			break
		}
		for _, m := range batch {
			if m.PimSkuID == pimSkuID && m.InvSkuID > 0 {
				return m.InvSkuID, nil
			}
		}
		if int64(page*100) >= pd.Total {
			break
		}
	}
	return 0, nil
}

func (c *Client) GetSkuByCode(ctx context.Context, bearerToken, skuCode string) (*SkuByCode, error) {
	code := strings.TrimSpace(skuCode)
	if code == "" {
		return nil, fmt.Errorf("skuCode required")
	}
	q := url.Values{}
	q.Set("skuCode", code)
	reqURL := c.baseURL + "/api/v1/admin/skus/by-code?" + q.Encode()
	body, err := c.doGET(ctx, bearerToken, reqURL)
	if err != nil {
		return nil, err
	}
	var sku SkuByCode
	if err := json.Unmarshal(body, &sku); err != nil {
		return nil, fmt.Errorf("warehousecore sku decode: %w", err)
	}
	return &sku, nil
}

func (c *Client) PostPurchaseInbound(ctx context.Context, bearerToken string, in *PurchaseInboundRequest) error {
	if in == nil {
		return fmt.Errorf("request required")
	}
	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}
	reqURL := c.baseURL + "/api/v1/admin/integrations/purchase-inbound"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", bearerToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("warehousecore request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("warehousecore http %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var wrapped apiBody
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return fmt.Errorf("warehousecore decode: %w", err)
	}
	if wrapped.Code != 200 && wrapped.Code != 201 {
		msg := wrapped.Message
		if msg == "" {
			msg = "warehousecore error"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func (c *Client) doGET(ctx context.Context, bearerToken, reqURL string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", bearerToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("warehousecore request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("warehousecore http %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var wrapped apiBody
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("warehousecore decode: %w", err)
	}
	if wrapped.Code != 200 {
		msg := wrapped.Message
		if msg == "" {
			msg = "warehousecore error"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return wrapped.Data, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
