package service

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"supplycore/internal/dto"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/model"
	jwtmgr "supplycore/internal/pkg/jwt"
	"supplycore/internal/repo"
)

// SettlementMergeService 按供应商结算周期，在设定时刻自动合并上一完整周期（T+1）的代发采购单。
type SettlementMergeService struct {
	repos *repo.Repos
	po    *PurchaseOrderService
	oc    *ordercore.Client
	jwt   *jwtmgr.Manager
	loc   *time.Location
}

func NewSettlementMergeService(repos *repo.Repos, po *PurchaseOrderService, oc *ordercore.Client, jwt *jwtmgr.Manager) *SettlementMergeService {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.Local
	}
	return &SettlementMergeService{repos: repos, po: po, oc: oc, jwt: jwt, loc: loc}
}

func (s *SettlementMergeService) RunOnce(ctx context.Context) {
	if s == nil || s.repos == nil {
		return
	}
	list, err := s.repos.Supplier.ListWithSettlementCycle()
	if err != nil {
		log.Printf("[settlement-merge] list suppliers: %v", err)
		return
	}
	now := time.Now().In(s.loc)
	for i := range list {
		sup := &list[i]
		if err := s.mergeSupplier(ctx, now, sup); err != nil {
			log.Printf("[settlement-merge] supplier=%d(%s): %v", sup.ID, sup.Name, err)
		}
	}
}

func (s *SettlementMergeService) mergeSupplier(ctx context.Context, now time.Time, sup *model.Supplier) error {
	if !settlementDue(now, sup.SettlementMergeTime) {
		return nil
	}
	// 每个自然日只跑一次（到达合并时刻后）
	if settlementRanToday(now, sup.SettlementLastRunAt) {
		return nil
	}
	from, to := settlementWindow(now, sup.SettlementCycle, sup.SettlementCustomDays)
	if !to.After(from) {
		return nil
	}
	pr := s.repos.PurchaseOrder.ForTenant(sup.TenantID)
	pos, err := pr.ListDropshipMergeable(sup.ID, from, to)
	if err != nil {
		return err
	}
	if len(pos) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(pos))
	for _, po := range pos {
		ids = append(ids, po.ID)
	}
	poSvc := s.po.ForTenant(sup.TenantID)
	source := strings.TrimSpace(sup.SyncPurchasePriceFrom)

	// 仅 1 张可合并单：只做采购价同步，不做合并
	if len(pos) < 2 {
		if source != "" && s.oc != nil && s.jwt != nil {
			if token, terr := s.jwt.IssueServiceToken(sup.TenantID, 15*time.Minute); terr != nil {
				log.Printf("[settlement-merge] issue token tenant=%d: %v", sup.TenantID, terr)
			} else {
				n, serr := poSvc.SyncDropshipPurchasePricesFromOrders(ctx, s.oc, "Bearer "+token, ids, source)
				if serr != nil {
					log.Printf("[settlement-merge] sync purchase price supplier=%d: %v", sup.ID, serr)
				} else if n > 0 {
					log.Printf("[settlement-merge] supplier=%d synced purchase price on %d po(s) from %s", sup.ID, n, source)
				}
			}
		}
		_ = s.repos.Supplier.UpdateSettlementLastRunAt(sup.ID, now)
		return nil
	}

	var token string
	if source != "" && s.oc != nil && s.jwt != nil {
		t, terr := s.jwt.IssueServiceToken(sup.TenantID, 15*time.Minute)
		if terr != nil {
			log.Printf("[settlement-merge] issue token tenant=%d: %v", sup.TenantID, terr)
		} else {
			token = "Bearer " + t
			n, serr := poSvc.SyncDropshipPurchasePricesFromOrders(ctx, s.oc, token, ids, source)
			if serr != nil {
				log.Printf("[settlement-merge] sync purchase price supplier=%d: %v", sup.ID, serr)
			} else if n > 0 {
				log.Printf("[settlement-merge] supplier=%d synced purchase price on %d po(s) from %s", sup.ID, n, source)
			}
		}
	}

	result, err := poSvc.Merge(&dto.MergePurchaseOrdersInput{
		SourcePoIDs: ids,
		TargetPoID:  ids[0],
	})
	if err != nil {
		return err
	}
	if s.oc != nil && s.jwt != nil && result != nil && len(result.MergedFromPoNos) > 0 && result.PoNo != "" {
		if token == "" {
			t, terr := s.jwt.IssueServiceToken(sup.TenantID, 15*time.Minute)
			if terr != nil {
				log.Printf("[settlement-merge] issue token tenant=%d: %v", sup.TenantID, terr)
			} else {
				token = "Bearer " + t
			}
		}
		if token != "" {
			if _, rerr := s.oc.RelinkPurchaseOrder(ctx, token, result.MergedFromPoNos, result.PoNo); rerr != nil {
				log.Printf("[settlement-merge] relink po=%s: %v", result.PoNo, rerr)
			}
			if source != "" && result.ID > 0 {
				if _, serr := poSvc.SyncDropshipPurchasePricesFromOrders(ctx, s.oc, token, []uint64{result.ID}, source); serr != nil {
					log.Printf("[settlement-merge] sync purchase price after merge po=%s: %v", result.PoNo, serr)
				}
			}
		}
	}
	_ = s.repos.Supplier.UpdateSettlementLastRunAt(sup.ID, now)
	log.Printf("[settlement-merge] supplier=%d merged %d -> %s (window %s ~ %s)",
		sup.ID, len(ids), result.PoNo, from.Format(time.RFC3339), to.Format(time.RFC3339))
	return nil
}

func settlementRanToday(now time.Time, last *time.Time) bool {
	if last == nil {
		return false
	}
	a := last.In(now.Location())
	return a.Year() == now.Year() && a.YearDay() == now.YearDay()
}

func settlementDue(now time.Time, mergeTime string) bool {
	h, m, ok := parseHHMM(mergeTime)
	if !ok {
		h, m = 18, 30
	}
	return now.Hour() > h || (now.Hour() == h && now.Minute() >= m)
}

func parseHHMM(v string) (hour, minute int, ok bool) {
	v = strings.TrimSpace(v)
	parts := strings.Split(v, ":")
	if len(parts) < 2 {
		return 0, 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

// settlementWindow 返回上一完整结算周期 [from, to)（T+1：到点合并上一周期，不含今天未完结当天）。
func settlementWindow(now time.Time, cycle string, customDays int) (time.Time, time.Time) {
	y, m, d := now.Date()
	startOfDay := time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	switch cycle {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisMonday := startOfDay.AddDate(0, 0, -(weekday - 1))
		return thisMonday.AddDate(0, 0, -7), thisMonday
	case "month":
		thisMonth := time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
		return thisMonth.AddDate(0, -1, 0), thisMonth
	case "custom":
		days := customDays
		if days < 1 {
			days = 1
		}
		return startOfDay.AddDate(0, 0, -days), startOfDay
	default: // day：昨天 00:00 ~ 今天 00:00
		return startOfDay.AddDate(0, 0, -1), startOfDay
	}
}
