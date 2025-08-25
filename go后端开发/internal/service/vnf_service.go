package service

import (
	"context"

	"vnf-config/internal/infra/db"
	"vnf-config/internal/model"
)

type VNFService struct{}

func NewVNFService() *VNFService { return &VNFService{} }

func (s *VNFService) List(ctx context.Context, page, pageSize int, keyword string) ([]model.VNFInstance, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 || pageSize > 100 { pageSize = 10 }
	var items []model.VNFInstance
	q := db.MySQLDB.Model(&model.VNFInstance{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	if err := q.Order("id desc").Limit(pageSize).Offset((page-1)*pageSize).Find(&items).Error; err != nil { return nil, 0, err }
	return items, total, nil
}

func (s *VNFService) Get(ctx context.Context, id uint) (*model.VNFInstance, error) {
	var item model.VNFInstance
	if err := db.MySQLDB.First(&item, id).Error; err != nil { return nil, err }
	return &item, nil
}

func (s *VNFService) Delete(ctx context.Context, id uint) error {
	return db.MySQLDB.Delete(&model.VNFInstance{}, id).Error
}


