package ordercore

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

type OrderBrief struct {
	ID              uint64  `json:"id"`
	OrderNo         string  `json:"orderNo"`
	PlatformOrderID string  `json:"platformOrderId"`
	ShopName        string  `json:"shopName"`
	BuyerName       string  `json:"buyerName"`
	BuyerNick       string  `json:"buyerNick"`
	Status          string  `json:"status"`
	ShipStatus      string  `json:"shipStatus"`
	TotalAmount     float64 `json:"totalAmount"`
	PayAmount       float64 `json:"payAmount"`
	OrderedAt       *string `json:"orderedAt,omitempty"`
	CreatedAt       string  `json:"createdAt"`
}

type pagePayload[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type apiBody struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) SearchOrders(ctx context.Context, bearerToken, keyword string, page, pageSize int) ([]OrderBrief, int64, error) {
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
	var pageData pagePayload[OrderBrief]
	if err := c.getJSON(ctx, bearerToken, "/api/v1/admin/orders?"+q.Encode(), &pageData); err != nil {
		return nil, 0, err
	}
	if pageData.List == nil {
		pageData.List = []OrderBrief{}
	}
	return pageData.List, pageData.Total, nil
}

func (c *Client) getJSON(ctx context.Context, bearerToken, path string, out any) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", bearerToken)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ordercore request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("ordercore http %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var wrapped apiBody
	if err := json.Unmarshal(body, &wrapped); err != nil {
		return fmt.Errorf("ordercore decode: %w", err)
	}
	if wrapped.Code != 200 {
		msg := wrapped.Message
		if msg == "" {
			msg = "ordercore error"
		}
		return fmt.Errorf("%s", msg)
	}
	if err := json.Unmarshal(wrapped.Data, out); err != nil {
		return fmt.Errorf("ordercore data decode: %w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
