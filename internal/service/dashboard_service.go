package service

import (
	"fmt"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"
)

type DashboardService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewDashboardService(repos *repo.Repos) *DashboardService {
	return &DashboardService{repos: repos}
}

func (s *DashboardService) ForTenant(tenantID uint64) *DashboardService {
	return &DashboardService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *DashboardService) Stats() (*dto.DashboardStats, error) {
	r := s.repos.Dashboard.ForTenant(s.tenantID)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	week := today.AddDate(0, 0, -6)
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	year := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	inProgressStatuses := []string{
		model.POStatusPaid,
		model.POStatusPartialShipped,
		model.POStatusShipped,
		model.POStatusPartialReceived,
	}
	shippedStatuses := []string{
		model.POStatusPartialShipped,
		model.POStatusShipped,
	}

	out := &dto.DashboardStats{}

	var err error
	// 工作场景默认按今日业务日（COALESCE(ordered_at, created_at)）统计
	if out.Workbench.DropshipPO, err = r.CountPOsByFulfillmentSince(model.POFulfillmentDropship, true, &today); err != nil {
		return nil, err
	}
	if out.Workbench.StockInPO, err = r.CountPOsByFulfillmentSince(model.POFulfillmentStockIn, true, &today); err != nil {
		return nil, err
	}
	if out.Workbench.DraftPO, err = r.CountPOsByStatusSince(model.POStatusDraft, &today); err != nil {
		return nil, err
	}
	if out.Workbench.OrderedPO, err = r.CountPOsByStatusSince(model.POStatusOrdered, &today); err != nil {
		return nil, err
	}
	if out.Workbench.UnpaidPO, err = r.CountUnpaidPOsSince(&today); err != nil {
		return nil, err
	}
	if out.Workbench.InTransitPO, err = r.CountPOsByStatusesSince(shippedStatuses, &today); err != nil {
		return nil, err
	}
	if out.Workbench.PartialReceivedPO, err = r.CountPOsByStatusSince(model.POStatusPartialReceived, &today); err != nil {
		return nil, err
	}
	if out.Workbench.ActiveOffers, err = r.CountOffers(true); err != nil {
		return nil, err
	}
	saleAmt, purchaseAmt, err := r.SumDropshipSaleAndPurchaseOnDay(today)
	if err != nil {
		return nil, err
	}
	out.Workbench.TodayDropshipSaleAmount = saleAmt
	out.Workbench.TodayDropshipPurchaseAmount = purchaseAmt
	out.Workbench.TodayDropshipProfit = saleAmt - purchaseAmt

	if out.Supplier.Total, err = r.CountSuppliers(false); err != nil {
		return nil, err
	}
	if out.Supplier.Active, err = r.CountSuppliers(true); err != nil {
		return nil, err
	}
	if out.Supplier.OfferCount, err = r.CountOffers(false); err != nil {
		return nil, err
	}
	if out.Supplier.OrderedThisMonth, err = r.CountDistinctSuppliersSince(month); err != nil {
		return nil, err
	}

	if out.PurchaseOrder.Total, err = r.CountPOs(); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.Draft, err = r.CountPOsByStatus(model.POStatusDraft); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.InProgress, err = r.CountPOsByStatuses(inProgressStatuses); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.Completed, err = r.CountPOsByStatus(model.POStatusCompleted); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.Cancelled, err = r.CountPOsByStatus(model.POStatusCancelled); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.TodayCount, err = r.CountPOsSince(today, true); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.WeekCount, err = r.CountPOsSince(week, true); err != nil {
		return nil, err
	}
	if out.PurchaseOrder.MonthCount, err = r.CountPOsSince(month, true); err != nil {
		return nil, err
	}

	if out.Cost.TodayAmount, err = r.SumPOAmountSince(today); err != nil {
		return nil, err
	}
	if out.Cost.WeekAmount, err = r.SumPOAmountSince(week); err != nil {
		return nil, err
	}
	if out.Cost.MonthAmount, err = r.SumPOAmountSince(month); err != nil {
		return nil, err
	}
	if out.Cost.YearAmount, err = r.SumPOAmountSince(year); err != nil {
		return nil, err
	}
	if out.Cost.UnpaidAmount, err = r.SumUnpaidAmount(); err != nil {
		return nil, err
	}

	statusRows, err := r.GroupPOByStatus()
	if err != nil {
		return nil, err
	}
	out.StatusBreakdown = make([]dto.DashboardStatusCount, 0, len(statusRows))
	for _, row := range statusRows {
		out.StatusBreakdown = append(out.StatusBreakdown, dto.DashboardStatusCount{
			Status: row.Status,
			Count:  row.Count,
		})
	}

	rankRows, err := r.TopSuppliersSince(month, 8)
	if err != nil {
		return nil, err
	}
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	out.TopSuppliers = make([]dto.DashboardSupplierRank, 0, len(rankRows))
	for _, row := range rankRows {
		item := dto.DashboardSupplierRank{
			SupplierID:  row.SupplierID,
			OrderCount:  row.OrderCount,
			TotalAmount: row.TotalAmount,
		}
		if sup, err := sr.GetByID(row.SupplierID); err == nil {
			item.SupplierName = sup.Name
		}
		out.TopSuppliers = append(out.TopSuppliers, item)
	}

	recent, err := r.RecentPOs(8)
	if err != nil {
		return nil, err
	}
	out.RecentOrders = make([]dto.PurchaseOrderListItem, 0, len(recent))
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	for _, po := range recent {
		item := dto.PurchaseOrderListItem{
			ID: po.ID, PoNo: po.PoNo, SupplierID: po.SupplierID,
			Status: po.Status, PayStatus: po.PayStatus,
			FulfillmentType: po.FulfillmentType,
			TotalAmount: po.TotalAmount, Currency: po.Currency,
			RefSoID: po.RefSoID, RefTraceID: po.RefTraceID,
			CreatedAt: formatTime(po.CreatedAt),
		}
		if po.OrderedAt != nil {
			item.OrderedAt = formatTimePtr(po.OrderedAt)
		}
		if sup, err := sr.GetByID(po.SupplierID); err == nil {
			item.SupplierName = sup.Name
		}
		if n, err := pr.CountItems(po.ID); err == nil {
			item.ItemCount = int(n)
		}
		out.RecentOrders = append(out.RecentOrders, item)
	}

	return out, nil
}

func (s *DashboardService) Trend(startDate, endDate string) (*dto.DashboardTrend, error) {
	r := s.repos.Dashboard.ForTenant(s.tenantID)
	var start, end time.Time
	var err error
	if startDate != "" {
		start, err = time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err != nil {
			return nil, fmt.Errorf("%w: startDate 格式应为 YYYY-MM-DD", ErrBadRequest)
		}
	}
	if endDate != "" {
		end, err = time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err != nil {
			return nil, fmt.Errorf("%w: endDate 格式应为 YYYY-MM-DD", ErrBadRequest)
		}
	}
	start, end, err = repo.NormalizeDashboardRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}
	points, err := r.DailyDropshipTrend(start, end)
	if err != nil {
		return nil, err
	}
	out := &dto.DashboardTrend{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
		Points:    points,
	}
	for _, p := range points {
		out.OrderCount += p.OrderCount
		out.SaleAmount += p.SaleAmount
		out.PurchaseAmount += p.PurchaseAmount
		out.Profit += p.Profit
	}
	return out, nil
}
