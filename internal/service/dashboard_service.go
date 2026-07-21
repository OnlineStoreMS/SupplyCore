package service

import (
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
		model.POStatusInTransit,
		model.POStatusPartialReceived,
	}
	inTransitStatuses := []string{
		model.POStatusPartialShipped,
		model.POStatusInTransit,
	}

	out := &dto.DashboardStats{}

	var err error
	if out.Workbench.DraftPO, err = r.CountPOsByStatus(model.POStatusDraft); err != nil {
		return nil, err
	}
	if out.Workbench.OrderedPO, err = r.CountPOsByStatus(model.POStatusOrdered); err != nil {
		return nil, err
	}
	if out.Workbench.UnpaidPO, err = r.CountUnpaidPOs(); err != nil {
		return nil, err
	}
	if out.Workbench.InTransitPO, err = r.CountPOsByStatuses(inTransitStatuses); err != nil {
		return nil, err
	}
	if out.Workbench.PartialReceivedPO, err = r.CountPOsByStatus(model.POStatusPartialReceived); err != nil {
		return nil, err
	}
	if out.Workbench.ActiveOffers, err = r.CountOffers(true); err != nil {
		return nil, err
	}

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
