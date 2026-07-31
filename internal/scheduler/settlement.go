package scheduler

import (
	"context"
	"log"
	"time"

	"supplycore/internal/service"
)

type SettlementScheduler struct {
	svc    *service.SettlementMergeService
	stopCh chan struct{}
}

func NewSettlementScheduler(svc *service.SettlementMergeService) *SettlementScheduler {
	return &SettlementScheduler{
		svc:    svc,
		stopCh: make(chan struct{}),
	}
}

func (s *SettlementScheduler) Start() {
	go s.loop()
}

func (s *SettlementScheduler) Stop() {
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}

func (s *SettlementScheduler) loop() {
	// 启动稍后跑一次，之后每分钟检查
	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-timer.C:
			s.runOnce()
			timer.Reset(60 * time.Second)
		}
	}
}

func (s *SettlementScheduler) runOnce() {
	if s.svc == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	log.Printf("[settlement-scheduler] checking due merges")
	s.svc.RunOnce(ctx)
}
