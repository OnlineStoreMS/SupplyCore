package productcore

import (
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
			Timeout: 15 * time.Second,
		},
	}
}

type SkuSearchItem struct {
	ProductID    uint64            `json:"productId"`
	ProductName  string            `json:"productName"`
	MaterialCode string            `json:"materialCode"`
	ProductSn    string            `json:"productSn"`
	ProductPic   string            `json:"productPic"`
	BrandName    string            `json:"brandName"`
	CategoryName string            `json:"categoryName"`
	PublishStatus int8             `json:"publishStatus"`
	SkuID        uint64            `json:"skuId"`
	SkuCode      string            `json:"skuCode"`
	Specs        map[string]string `json:"specs"`
	SpecLabel    string            `json:"specLabel"`
	Price        float64           `json:"price"`
	Stock        int               `json:"stock"`
	Pic          string            `json:"pic"`
}

type pagePayload struct {
	List     []SkuSearchItem `json:"list"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type apiBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) SearchSkus(ctx context.Context, bearerToken, keyword string, page, pageSize int) ([]SkuSearchItem, int64, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, 0, fmt.Errorf("keyword required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	q := url.Values{}
	q.Set("keyword", keyword)
	q.Set("page", strconv.Itoa(page))
	q.Set("pageSize", strconv.Itoa(pageSize))
	reqURL := c.baseURL + "/api/v1/admin/super-search?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", bearerToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("productcore request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("productcore http %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var wrapped apiBody
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return nil, 0, fmt.Errorf("productcore decode: %w", err)
	}
	if wrapped.Code != 200 {
		msg := wrapped.Message
		if msg == "" {
			msg = "productcore error"
		}
		return nil, 0, fmt.Errorf("%s", msg)
	}

	var pageData pagePayload
	if err := json.Unmarshal(wrapped.Data, &pageData); err != nil {
		return nil, 0, fmt.Errorf("productcore page decode: %w", err)
	}
	if pageData.List == nil {
		pageData.List = []SkuSearchItem{}
	}
	return pageData.List, pageData.Total, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
