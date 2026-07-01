package service

import (
	"errors"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type OfferService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewOfferService(repos *repo.Repos) *OfferService {
	return &OfferService{repos: repos}
}

func (s *OfferService) ForTenant(tenantID uint64) *OfferService {
	return &OfferService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *OfferService) List(f repo.OfferListFilter) ([]dto.OfferDetail, int64, error) {
	list, total, err := s.repos.Offer.ForTenant(s.tenantID).List(f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.OfferDetail, 0, len(list))
	for _, item := range list {
		out = append(out, s.toDetail(&item))
	}
	return out, total, nil
}

func (s *OfferService) Get(id uint64) (*dto.OfferDetail, error) {
	item, err := s.repos.Offer.ForTenant(s.tenantID).GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	detail := s.toDetail(item)
	return &detail, nil
}

func (s *OfferService) Create(in *dto.SkuOfferDTO) (*dto.OfferDetail, error) {
	if err := s.validateRefs(in.SupplierID, in.ShipFromAddressID); err != nil {
		return nil, err
	}
	r := s.repos.Offer.ForTenant(s.tenantID)
	item := s.fromDTO(in)
	if item.Currency == "" {
		item.Currency = "CNY"
	}
	if item.MinOrderQty <= 0 {
		item.MinOrderQty = 1
	}
	if item.IsPrimary {
		_ = r.ClearPrimary(item.SkuID, 0)
	}
	if err := r.Create(item); err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateCode
		}
		return nil, err
	}
	detail := s.toDetail(item)
	return &detail, nil
}

func (s *OfferService) Update(id uint64, in *dto.SkuOfferDTO) (*dto.OfferDetail, error) {
	r := s.repos.Offer.ForTenant(s.tenantID)
	item, err := r.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := s.validateRefs(in.SupplierID, in.ShipFromAddressID); err != nil {
		return nil, err
	}
	item.SkuID = in.SkuID
	item.SupplierID = in.SupplierID
	item.SupplierSkuCode = in.SupplierSkuCode
	item.SupplyPrice = in.SupplyPrice
	if in.Currency != "" {
		item.Currency = in.Currency
	}
	item.MinOrderQty = in.MinOrderQty
	if item.MinOrderQty <= 0 {
		item.MinOrderQty = 1
	}
	item.LeadTimeDays = in.LeadTimeDays
	item.ShipFromAddressID = in.ShipFromAddressID
	item.SupportsDropship = in.SupportsDropship
	item.SupportsSelfStock = in.SupportsSelfStock
	item.IsPrimary = in.IsPrimary
	item.Priority = in.Priority
	item.Status = defaultStatus(in.Status)
	item.Remark = in.Remark
	if item.IsPrimary {
		_ = r.ClearPrimary(item.SkuID, id)
	}
	if err := r.Save(item); err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateCode
		}
		return nil, err
	}
	detail := s.toDetail(item)
	return &detail, nil
}

func (s *OfferService) Delete(id uint64) error {
	r := s.repos.Offer.ForTenant(s.tenantID)
	if _, err := r.GetByID(id); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return r.Delete(id)
}

func (s *OfferService) SupplyOptions(skuID uint64, dropshipOnly bool) (*dto.SupplyOptionsResp, error) {
	list, err := s.repos.Offer.ForTenant(s.tenantID).ListBySku(skuID, true)
	if err != nil {
		return nil, err
	}
	resp := &dto.SupplyOptionsResp{SkuID: skuID, Offers: make([]dto.SupplyOptionOffer, 0, len(list))}
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	for _, item := range list {
		if dropshipOnly && !item.SupportsDropship {
			continue
		}
		supplier, _ := sr.GetByID(item.SupplierID)
		opt := dto.SupplyOptionOffer{
			OfferID: item.ID, SupplierID: item.SupplierID,
			SupplierSkuCode: item.SupplierSkuCode, SupplyPrice: item.SupplyPrice,
			Currency: item.Currency, SupportsDropship: item.SupportsDropship,
			SupportsSelfStock: item.SupportsSelfStock, LeadTimeDays: item.LeadTimeDays,
			IsPrimary: item.IsPrimary, Priority: item.Priority,
		}
		if supplier != nil {
			opt.SupplierName = supplier.Name
			opt.SupplierCode = supplier.Code
		}
		if item.ShipFromAddressID > 0 {
			if addr, err := sr.GetAddress(item.SupplierID, item.ShipFromAddressID); err == nil {
				opt.ShipFrom = &dto.ShipFromBrief{
					Label: addr.Label, Province: addr.Province,
					City: addr.City, District: addr.District, Address: addr.Address,
				}
			}
		}
		resp.Offers = append(resp.Offers, opt)
	}
	return resp, nil
}

func (s *OfferService) validateRefs(supplierID, addressID uint64) error {
	if _, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(supplierID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	if addressID == 0 {
		return nil
	}
	if _, err := s.repos.Supplier.ForTenant(s.tenantID).GetAddress(supplierID, addressID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return nil
}

func (s *OfferService) fromDTO(in *dto.SkuOfferDTO) *model.SkuSupplierOffer {
	return &model.SkuSupplierOffer{
		SkuID: in.SkuID, SupplierID: in.SupplierID, SupplierSkuCode: in.SupplierSkuCode,
		SupplyPrice: in.SupplyPrice, Currency: in.Currency, MinOrderQty: in.MinOrderQty,
		LeadTimeDays: in.LeadTimeDays, ShipFromAddressID: in.ShipFromAddressID,
		SupportsDropship: in.SupportsDropship, SupportsSelfStock: in.SupportsSelfStock,
		IsPrimary: in.IsPrimary, Priority: in.Priority, Status: defaultStatus(in.Status),
		Remark: in.Remark,
	}
}

func (s *OfferService) toDetail(item *model.SkuSupplierOffer) dto.OfferDetail {
	detail := dto.OfferDetail{
		ID: item.ID, SkuID: item.SkuID, SupplierID: item.SupplierID,
		SupplierSkuCode: item.SupplierSkuCode, SupplyPrice: item.SupplyPrice,
		Currency: item.Currency, MinOrderQty: item.MinOrderQty,
		LeadTimeDays: item.LeadTimeDays, ShipFromAddressID: item.ShipFromAddressID,
		SupportsDropship: item.SupportsDropship, SupportsSelfStock: item.SupportsSelfStock,
		IsPrimary: item.IsPrimary, Priority: item.Priority, Status: item.Status, Remark: item.Remark,
	}
	sr := s.repos.Supplier.ForTenant(s.tenantID)
	if supplier, err := sr.GetByID(item.SupplierID); err == nil {
		detail.SupplierName = supplier.Name
		detail.SupplierCode = supplier.Code
	}
	if item.ShipFromAddressID > 0 {
		if addr, err := sr.GetAddress(item.SupplierID, item.ShipFromAddressID); err == nil {
			detail.ShipFromLabel = addr.Label
			detail.ShipFromCity = addr.City
		}
	}
	return detail
}
