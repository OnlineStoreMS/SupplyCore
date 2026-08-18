package service

import (
	"context"
	"fmt"
	"strings"

	"supplycore/internal/dto"
	"supplycore/internal/integrations/ordercore"
	"supplycore/internal/model"
)

// IsSplitChildPOItem 是否为拆分子行。
func IsSplitChildPOItem(it model.PurchaseOrderItem) bool {
	return strings.TrimSpace(it.SplitKind) != "" || it.ParentPOItemID > 0
}

// IsShippablePOItem 有拆分子行时原行不可发；整单拆分时仅子行可发。
func IsShippablePOItem(items []model.PurchaseOrderItem, it model.PurchaseOrderItem) bool {
	if it.Cancelled {
		return false
	}
	if IsSplitChildPOItem(it) {
		return true
	}
	hasFull := false
	for _, x := range items {
		if x.Cancelled {
			continue
		}
		if strings.TrimSpace(x.SplitKind) == model.SplitKindFull {
			hasFull = true
			break
		}
	}
	if hasFull {
		return false
	}
	for _, x := range items {
		if x.Cancelled {
			continue
		}
		if strings.TrimSpace(x.SplitKind) == model.SplitKindPartial && x.ParentPOItemID == it.ID {
			return false
		}
	}
	return true
}

// SplitPOItem 按商品拆分规格：重建未发子行；代发且已关联销售单时同步订单中心。
func (s *POTrackingService) SplitPOItem(ctx context.Context, poID, parentItemID uint64, bearerToken string, in *dto.SplitPOItemInput) (*dto.SplitPOItemResult, error) {
	if in == nil || len(in.Lines) == 0 {
		return nil, ErrBadRequest
	}
	po, err := s.ensurePOTrackable(poID)
	if err != nil {
		return nil, err
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	full, err := pr.GetWithItems(poID)
	if err != nil {
		return nil, err
	}
	var parent *model.PurchaseOrderItem
	for i := range full.Items {
		if full.Items[i].ID == parentItemID {
			parent = &full.Items[i]
			break
		}
	}
	if parent == nil {
		return nil, ErrNotFound
	}
	if parent.Cancelled {
		return nil, fmt.Errorf("已作废明细不可拆分")
	}
	if IsSplitChildPOItem(*parent) {
		return nil, fmt.Errorf("拆分子行不可再拆分，请编辑父商品拆分")
	}

	for i, line := range in.Lines {
		if strings.TrimSpace(line.SkuName) == "" {
			return nil, fmt.Errorf("第 %d 行规格名称不能为空", i+1)
		}
		if line.Qty <= 0 {
			return nil, fmt.Errorf("第 %d 行数量须大于 0", i+1)
		}
	}

	shippedItemIDs := map[uint64]struct{}{}
	shipments, _ := s.repos.Shipment.ForTenant(s.tenantID).ListByPO(poID)
	for _, sh := range shipments {
		for _, si := range sh.Items {
			if si.POItemID > 0 {
				shippedItemIDs[si.POItemID] = struct{}{}
			}
		}
	}

	type childState struct {
		item    model.PurchaseOrderItem
		shipped bool
	}
	byPlan := map[uint64]*childState{}
	var unshippedChildren []model.PurchaseOrderItem
	for _, it := range full.Items {
		if it.ParentPOItemID != parent.ID {
			continue
		}
		_, shipped := shippedItemIDs[it.ID]
		st := &childState{item: it, shipped: shipped}
		if it.ShipPlanLineID > 0 {
			byPlan[it.ShipPlanLineID] = st
		}
		if !shipped {
			unshippedChildren = append(unshippedChildren, it)
		}
	}

	keepPlan := map[uint64]struct{}{}
	for _, line := range in.Lines {
		if line.ShipPlanLineID > 0 {
			keepPlan[line.ShipPlanLineID] = struct{}{}
		}
	}

	toDelete := make([]uint64, 0)
	for _, ch := range unshippedChildren {
		if ch.ShipPlanLineID > 0 {
			if _, ok := keepPlan[ch.ShipPlanLineID]; ok {
				continue
			}
		}
		toDelete = append(toDelete, ch.ID)
	}
	if err := pr.DeleteItemsByIDs(poID, toDelete); err != nil {
		return nil, err
	}
	for _, id := range toDelete {
		for planID, st := range byPlan {
			if st.item.ID == id {
				delete(byPlan, planID)
			}
		}
	}

	createdOrUpdated := make([]model.PurchaseOrderItem, 0, len(in.Lines))
	for _, line := range in.Lines {
		sku := strings.TrimSpace(line.SkuName)
		planID := line.ShipPlanLineID
		if planID == 0 {
			nid, nerr := pr.NextShipPlanLineID()
			if nerr != nil {
				return nil, nerr
			}
			planID = nid
		}
		if st, ok := byPlan[planID]; ok && !st.shipped {
			st.item.ProductName = sku
			st.item.SkuSpecs = sku
			st.item.Qty = line.Qty
			st.item.SplitKind = model.SplitKindPartial
			st.item.ParentPOItemID = parent.ID
			st.item.ShipPlanLineID = planID
			st.item.SkuID = parent.SkuID
			st.item.SkuCode = parent.SkuCode
			st.item.PicURL = parent.PicURL
			st.item.SupplierSkuCode = parent.SupplierSkuCode
			st.item.RefSoID = parent.RefSoID
			st.item.RefOrderNo = parent.RefOrderNo
			st.item.UnitPrice = 0
			st.item.LineAmount = 0
			if err := pr.SaveItem(&st.item); err != nil {
				return nil, err
			}
			createdOrUpdated = append(createdOrUpdated, st.item)
			continue
		}
		if st, ok := byPlan[planID]; ok && st.shipped {
			st.item.ProductName = sku
			st.item.SkuSpecs = sku
			if line.Qty >= st.item.Qty {
				st.item.Qty = line.Qty
			}
			if err := pr.SaveItem(&st.item); err != nil {
				return nil, err
			}
			createdOrUpdated = append(createdOrUpdated, st.item)
			continue
		}
		child := model.PurchaseOrderItem{
			POID:            poID,
			SkuID:           parent.SkuID,
			OfferID:         parent.OfferID,
			ProductName:     sku,
			SkuCode:         parent.SkuCode,
			SkuSpecs:        sku,
			PicURL:          parent.PicURL,
			SupplierSkuCode:  parent.SupplierSkuCode,
			Qty:             line.Qty,
			UnitPrice:       0,
			LineAmount:      0,
			RefSoID:         parent.RefSoID,
			RefOrderNo:      parent.RefOrderNo,
			ParentPOItemID:  parent.ID,
			SplitKind:       model.SplitKindPartial,
			ShipPlanLineID:  planID,
		}
		if err := pr.CreateItem(&child); err != nil {
			return nil, err
		}
		createdOrUpdated = append(createdOrUpdated, child)
		byPlan[planID] = &childState{item: child, shipped: false}
	}

	result := &dto.SplitPOItemResult{}
	if po.FulfillmentType == model.POFulfillmentDropship && s.oc != nil {
		refSoID := parent.RefSoID
		if refSoID == 0 {
			refSoID = po.RefSoID
		}
		if refSoID > 0 {
			warn, serr := s.pushSplitToOrderCore(ctx, bearerToken, refSoID, parent, createdOrUpdated)
			if serr != nil {
				result.SyncWarning = serr.Error()
			} else {
				result.SyncedToOrderCore = true
				if warn != "" {
					result.SyncWarning = warn
				}
			}
		}
	}

	detail, gerr := pr.GetWithItems(poID)
	if gerr != nil {
		return nil, gerr
	}
	ps := NewPurchaseOrderService(s.repos).ForTenant(s.tenantID)
	result.PurchaseOrderDetail = ps.toDetail(detail)
	return result, nil
}

func (s *POTrackingService) pushSplitToOrderCore(
	ctx context.Context,
	bearerToken string,
	refSoID uint64,
	parent *model.PurchaseOrderItem,
	children []model.PurchaseOrderItem,
) (string, error) {
	order, err := s.oc.GetOrder(ctx, bearerToken, refSoID)
	if err != nil {
		return "", err
	}
	parentOCID := parent.RefOrderItemID
	if parentOCID == 0 {
		parentOCID = matchRootOrderItemID(order, *parent)
		if parentOCID > 0 {
			parent.RefOrderItemID = parentOCID
			_ = s.repos.PurchaseOrder.ForTenant(s.tenantID).SaveItem(parent)
		}
	}
	if parentOCID == 0 {
		return "", fmt.Errorf("无法匹配订单中心父商品行，请确认采购明细已关联销售单商品")
	}
	lines := make([]ordercore.SplitItemLineInput, 0, len(children))
	for _, ch := range children {
		if ch.ShipPlanLineID == 0 {
			continue
		}
		lines = append(lines, ordercore.SplitItemLineInput{
			ParentOrderItemID: parentOCID,
			SkuName:           strings.TrimSpace(ch.SkuSpecs),
			Qty:               ch.Qty,
			ShipPlanLineID:    ch.ShipPlanLineID,
		})
	}
	if len(lines) == 0 {
		return "", fmt.Errorf("没有可同步的拆分行")
	}
	synced, err := s.oc.SyncSplitItems(ctx, bearerToken, refSoID, ordercore.SyncSplitItemsRequest{
		Mode:  model.SplitKindPartial,
		Lines: lines,
	})
	if err != nil {
		return "", err
	}
	byPlan := map[uint64]uint64{}
	for _, l := range synced.Lines {
		if l.ShipPlanLineID > 0 && l.ID > 0 {
			byPlan[l.ShipPlanLineID] = l.ID
		}
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	for i := range children {
		ch := &children[i]
		if id, ok := byPlan[ch.ShipPlanLineID]; ok && ch.RefOrderItemID != id {
			ch.RefOrderItemID = id
			_ = pr.SaveItem(ch)
		}
	}
	return "", nil
}

func matchRootOrderItemID(order *ordercore.OrderBrief, poItem model.PurchaseOrderItem) uint64 {
	if order == nil {
		return 0
	}
	sku := strings.TrimSpace(poItem.SkuCode)
	name := strings.TrimSpace(poItem.ProductName)
	spec := strings.TrimSpace(poItem.SkuSpecs)
	var soft uint64
	for _, it := range order.Items {
		if strings.TrimSpace(it.SplitKind) != "" || it.ParentOrderItemID > 0 {
			continue
		}
		if poItem.RefOrderItemID > 0 && it.ID == poItem.RefOrderItemID {
			return it.ID
		}
		if sku != "" && strings.TrimSpace(it.SkuCode) == sku && it.Quantity == poItem.Qty {
			return it.ID
		}
		if soft == 0 && name != "" && strings.TrimSpace(it.ProductName) == name {
			soft = it.ID
		}
		if soft == 0 && spec != "" && strings.TrimSpace(it.SkuSpecs) == spec {
			soft = it.ID
		}
	}
	return soft
}

// syncSplitChildrenFromOrder 从订单中心拉拆分子行到采购明细（同步物流前调用）。
func (s *POTrackingService) syncSplitChildrenFromOrder(po *model.PurchaseOrder, items *[]model.PurchaseOrderItem, order *ordercore.OrderBrief) error {
	if order == nil || po == nil || items == nil {
		return nil
	}
	pr := s.repos.PurchaseOrder.ForTenant(s.tenantID)
	list := *items
	rootsByOC := map[uint64]*model.PurchaseOrderItem{}
	rootsBySku := map[string]*model.PurchaseOrderItem{}
	for i := range list {
		it := &list[i]
		if IsSplitChildPOItem(*it) || it.Cancelled {
			continue
		}
		if it.RefOrderItemID > 0 {
			rootsByOC[it.RefOrderItemID] = it
		}
		if code := strings.TrimSpace(it.SkuCode); code != "" {
			rootsBySku[code] = it
		}
	}
	childByOC := map[uint64]*model.PurchaseOrderItem{}
	childByPlan := map[uint64]*model.PurchaseOrderItem{}
	for i := range list {
		it := &list[i]
		if !IsSplitChildPOItem(*it) {
			continue
		}
		if it.RefOrderItemID > 0 {
			childByOC[it.RefOrderItemID] = it
		}
		if it.ShipPlanLineID > 0 {
			childByPlan[it.ShipPlanLineID] = it
		}
	}

	for _, ocIt := range order.Items {
		kind := strings.TrimSpace(ocIt.SplitKind)
		if kind == "" {
			continue
		}
		var parent *model.PurchaseOrderItem
		if kind == model.SplitKindPartial && ocIt.ParentOrderItemID > 0 {
			parent = rootsByOC[ocIt.ParentOrderItemID]
			if parent == nil {
				for i := range list {
					it := &list[i]
					if IsSplitChildPOItem(*it) || it.Cancelled {
						continue
					}
					if it.RefOrderItemID == ocIt.ParentOrderItemID {
						parent = it
						break
					}
				}
			}
		}
		if parent == nil {
			continue
		}
		sku := strings.TrimSpace(ocIt.SkuSpecs)
		if sku == "" {
			sku = strings.TrimSpace(ocIt.ProductName)
		}
		if sku == "" {
			continue
		}
		if existing := childByOC[ocIt.ID]; existing != nil {
			existing.ProductName = sku
			existing.SkuSpecs = sku
			existing.Qty = ocIt.Quantity
			if ocIt.ShipPlanLineID > 0 {
				existing.ShipPlanLineID = ocIt.ShipPlanLineID
			}
			_ = pr.SaveItem(existing)
			continue
		}
		if ocIt.ShipPlanLineID > 0 {
			if existing := childByPlan[ocIt.ShipPlanLineID]; existing != nil {
				existing.RefOrderItemID = ocIt.ID
				existing.ProductName = sku
				existing.SkuSpecs = sku
				existing.Qty = ocIt.Quantity
				_ = pr.SaveItem(existing)
				continue
			}
		}
		planID := ocIt.ShipPlanLineID
		if planID == 0 {
			nid, err := pr.NextShipPlanLineID()
			if err != nil {
				return err
			}
			planID = nid
		}
		child := model.PurchaseOrderItem{
			POID:            po.ID,
			SkuID:           parent.SkuID,
			OfferID:         parent.OfferID,
			ProductName:     sku,
			SkuCode:         parent.SkuCode,
			SkuSpecs:        sku,
			PicURL:          firstNonEmpty(ocIt.PicURL, parent.PicURL),
			SupplierSkuCode:  parent.SupplierSkuCode,
			Qty:             ocIt.Quantity,
			RefSoID:         parent.RefSoID,
			RefOrderNo:      parent.RefOrderNo,
			RefOrderItemID:  ocIt.ID,
			ParentPOItemID:  parent.ID,
			SplitKind:       kind,
			ShipPlanLineID:  planID,
		}
		if err := pr.CreateItem(&child); err != nil {
			return err
		}
		list = append(list, child)
		childByOC[ocIt.ID] = &list[len(list)-1]
	}

	for _, ocIt := range order.Items {
		if strings.TrimSpace(ocIt.SplitKind) != "" || ocIt.ParentOrderItemID > 0 {
			continue
		}
		if rootsByOC[ocIt.ID] != nil {
			continue
		}
		if code := strings.TrimSpace(ocIt.SkuCode); code != "" {
			if root := rootsBySku[code]; root != nil && root.RefOrderItemID == 0 {
				root.RefOrderItemID = ocIt.ID
				_ = pr.SaveItem(root)
				rootsByOC[ocIt.ID] = root
			}
		}
	}
	*items = list
	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
