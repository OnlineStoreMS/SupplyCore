package dto

// DashboardStats 工作台聚合统计
type DashboardStats struct {
	Workbench      DashboardWorkbench       `json:"workbench"`
	Supplier       DashboardSupplierStats   `json:"supplier"`
	PurchaseOrder  DashboardPOStats         `json:"purchaseOrder"`
	Cost           DashboardCostStats       `json:"cost"`
	TopSuppliers   []DashboardSupplierRank  `json:"topSuppliers"`
	RecentOrders   []PurchaseOrderListItem  `json:"recentOrders"`
	StatusBreakdown []DashboardStatusCount  `json:"statusBreakdown"`
}

type DashboardWorkbench struct {
	// DropshipPO：工作台「代发订单」卡数量（今日全部类型、排除已取消；字段名历史兼容）
	DropshipPO        int64 `json:"dropshipPO"`
	StockInPO         int64 `json:"stockInPO"`
	DraftPO           int64 `json:"draftPO"`
	OrderedPO         int64 `json:"orderedPO"`
	UnpaidPO          int64 `json:"unpaidPO"`
	InTransitPO       int64 `json:"inTransitPO"`
	PartialReceivedPO int64 `json:"partialReceivedPO"`
	ActiveOffers      int64 `json:"activeOffers"`
	// 今日代发毛利：仅 total_amount > 0；毛利润 = sale_amount - total_amount
	TodayDropshipSaleAmount     float64 `json:"todayDropshipSaleAmount"`
	TodayDropshipPurchaseAmount float64 `json:"todayDropshipPurchaseAmount"`
	TodayDropshipProfit         float64 `json:"todayDropshipProfit"`
}

type DashboardSupplierStats struct {
	Total              int64 `json:"total"`
	Active             int64 `json:"active"`
	OfferCount         int64 `json:"offerCount"`
	OrderedThisMonth   int64 `json:"orderedThisMonth"`
}

type DashboardPOStats struct {
	Total       int64 `json:"total"`
	Draft       int64 `json:"draft"`
	InProgress  int64 `json:"inProgress"`
	Completed   int64 `json:"completed"`
	Cancelled   int64 `json:"cancelled"`
	TodayCount  int64 `json:"todayCount"`
	WeekCount   int64 `json:"weekCount"`
	MonthCount  int64 `json:"monthCount"`
}

type DashboardCostStats struct {
	TodayAmount  float64 `json:"todayAmount"`
	WeekAmount   float64 `json:"weekAmount"`
	MonthAmount  float64 `json:"monthAmount"`
	UnpaidAmount float64 `json:"unpaidAmount"`
	YearAmount   float64 `json:"yearAmount"`
}

type DashboardSupplierRank struct {
	SupplierID   uint64  `json:"supplierId"`
	SupplierName string  `json:"supplierName"`
	OrderCount   int64   `json:"orderCount"`
	TotalAmount  float64 `json:"totalAmount"`
}

type DashboardStatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// DashboardTrend 代发趋势（按采购业务日）
type DashboardTrend struct {
	StartDate      string                `json:"startDate"`
	EndDate        string                `json:"endDate"`
	OrderCount     int64                 `json:"orderCount"`
	SaleAmount     float64               `json:"saleAmount"`
	PurchaseAmount float64               `json:"purchaseAmount"`
	Profit         float64               `json:"profit"`
	Points         []DashboardTrendPoint `json:"points"`
}

type DashboardTrendPoint struct {
	Date           string  `json:"date"`
	OrderCount     int64   `json:"orderCount"`
	SaleAmount     float64 `json:"saleAmount"`
	PurchaseAmount float64 `json:"purchaseAmount"`
	Profit         float64 `json:"profit"`
}
