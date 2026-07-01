package service

import (
	"errors"

	"supplycore/internal/dto"
	"supplycore/internal/model"
	"supplycore/internal/repo"

	"gorm.io/gorm"
)

type SupplierService struct {
	repos    *repo.Repos
	tenantID uint64
}

func NewSupplierService(repos *repo.Repos) *SupplierService {
	return &SupplierService{repos: repos}
}

func (s *SupplierService) ForTenant(tenantID uint64) *SupplierService {
	return &SupplierService{repos: s.repos, tenantID: repo.NormalizeTenantID(tenantID)}
}

func (s *SupplierService) List(keyword string, page, pageSize int) ([]model.Supplier, int64, error) {
	return s.repos.Supplier.ForTenant(s.tenantID).List(keyword, page, pageSize)
}

func (s *SupplierService) Get(id uint64) (*model.Supplier, error) {
	item, err := s.repos.Supplier.ForTenant(s.tenantID).GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return item, err
}

func (s *SupplierService) Create(in *dto.SupplierDTO) (*model.Supplier, error) {
	r := s.repos.Supplier.ForTenant(s.tenantID)
	if _, err := r.GetByCode(in.Code); err == nil {
		return nil, ErrDuplicateCode
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	item := &model.Supplier{
		Code: in.Code, Name: in.Name, ShortName: in.ShortName,
		Status: defaultStatus(in.Status), ContactName: in.ContactName,
		Phone: in.Phone, Email: in.Email, Remark: in.Remark,
		DefaultPaymentTerms: in.DefaultPaymentTerms,
		BankName: in.BankName, BankAccount: in.BankAccount, AccountName: in.AccountName,
	}
	if err := r.Create(item); err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateCode
		}
		return nil, err
	}
	return item, nil
}

func (s *SupplierService) Update(id uint64, in *dto.SupplierDTO) (*model.Supplier, error) {
	r := s.repos.Supplier.ForTenant(s.tenantID)
	item, err := r.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if in.Code != item.Code {
		if other, err := r.GetByCode(in.Code); err == nil && other.ID != id {
			return nil, ErrDuplicateCode
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	item.Code = in.Code
	item.Name = in.Name
	item.ShortName = in.ShortName
	item.Status = defaultStatus(in.Status)
	item.ContactName = in.ContactName
	item.Phone = in.Phone
	item.Email = in.Email
	item.Remark = in.Remark
	item.DefaultPaymentTerms = in.DefaultPaymentTerms
	item.BankName = in.BankName
	item.BankAccount = in.BankAccount
	item.AccountName = in.AccountName
	if err := r.Save(item); err != nil {
		if isDuplicateKey(err) {
			return nil, ErrDuplicateCode
		}
		return nil, err
	}
	return item, nil
}

func (s *SupplierService) Delete(id uint64) error {
	r := s.repos.Supplier.ForTenant(s.tenantID)
	if _, err := r.GetByID(id); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return r.Delete(id)
}

func (s *SupplierService) ListAddresses(supplierID uint64) ([]model.SupplierAddress, error) {
	if _, err := s.Get(supplierID); err != nil {
		return nil, err
	}
	return s.repos.Supplier.ForTenant(s.tenantID).ListAddresses(supplierID)
}

func (s *SupplierService) CreateAddress(supplierID uint64, in *dto.SupplierAddressDTO) (*model.SupplierAddress, error) {
	if _, err := s.Get(supplierID); err != nil {
		return nil, err
	}
	r := s.repos.Supplier.ForTenant(s.tenantID)
	item := &model.SupplierAddress{
		SupplierID: supplierID, Label: in.Label, ContactName: in.ContactName,
		Phone: in.Phone, Province: in.Province, City: in.City,
		District: in.District, Address: in.Address, IsDefault: in.IsDefault,
		Status: defaultStatus(in.Status),
	}
	if item.IsDefault {
		_ = r.ClearDefaultAddress(supplierID, 0)
	}
	if err := r.CreateAddress(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *SupplierService) UpdateAddress(supplierID, addressID uint64, in *dto.SupplierAddressDTO) (*model.SupplierAddress, error) {
	r := s.repos.Supplier.ForTenant(s.tenantID)
	item, err := r.GetAddress(supplierID, addressID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	item.Label = in.Label
	item.ContactName = in.ContactName
	item.Phone = in.Phone
	item.Province = in.Province
	item.City = in.City
	item.District = in.District
	item.Address = in.Address
	item.IsDefault = in.IsDefault
	item.Status = defaultStatus(in.Status)
	if item.IsDefault {
		_ = r.ClearDefaultAddress(supplierID, addressID)
	}
	if err := r.SaveAddress(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *SupplierService) DeleteAddress(supplierID, addressID uint64) error {
	r := s.repos.Supplier.ForTenant(s.tenantID)
	if _, err := r.GetAddress(supplierID, addressID); errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}
	return r.DeleteAddress(supplierID, addressID)
}

func defaultStatus(v int8) int8 {
	if v == 0 {
		return 1
	}
	return v
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "duplicate") || contains(msg, "unique")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (stringIndex(s, sub) >= 0)))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
