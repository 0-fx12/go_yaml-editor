package service

import (
	"context"
	"errors"

	"vnf-config/internal/dto"
	"vnf-config/internal/infra/db"
	"vnf-config/internal/model"
)

type DefinitionService struct{}

func NewDefinitionService() *DefinitionService { return &DefinitionService{} }

func (s *DefinitionService) List(ctx context.Context, vnfID uint, page, pageSize int, modifiedOnly bool) ([]model.VNFDefinition, int64, error) {
	if page <= 0 { page = 1 }
	if pageSize <= 0 || pageSize > 200 { pageSize = 10 }
	q := db.MySQLDB.Model(&model.VNFDefinition{}).Where("vnf_id = ?", vnfID)
	if modifiedOnly {
		q = q.Where("modified = ?", true)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	var items []model.VNFDefinition
	if err := q.Order("id asc").Limit(pageSize).Offset((page-1)*pageSize).Find(&items).Error; err != nil { return nil, 0, err }
	return items, total, nil
}

func (s *DefinitionService) Create(ctx context.Context, vnfID uint, req dto.DefinitionCreateRequest) (*model.VNFDefinition, error) {
	item := &model.VNFDefinition{
		VNFID:           vnfID,
		ParameterName:   req.ParameterName,
		DefaultValue:    req.DefaultValue,
		DescriptionText: req.DescriptionText,
		Type:            req.Type,
		CanBeUpdated:    req.CanBeUpdated,
		HiddenCondition: req.HiddenCondition,
		Optional:        req.Optional,
		Constraints:     req.Constraints,
	}
	if req.CurrentValue != nil {
		item.CurrentValue = *req.CurrentValue
		item.Modified = item.CurrentValue != item.DefaultValue
	}
	if err := db.MySQLDB.Create(item).Error; err != nil { return nil, err }
	return item, nil
}

func (s *DefinitionService) Update(ctx context.Context, vnfID, defID uint, req dto.DefinitionUpdateRequest) (*model.VNFDefinition, error) {
	var item model.VNFDefinition
	if err := db.MySQLDB.Where("id = ? AND vnf_id = ?", defID, vnfID).First(&item).Error; err != nil { return nil, err }
	if req.DefaultValue != nil { item.DefaultValue = *req.DefaultValue }
	if req.DescriptionText != nil { item.DescriptionText = *req.DescriptionText }
	if req.Type != nil { item.Type = *req.Type }
	if req.CanBeUpdated != nil { item.CanBeUpdated = *req.CanBeUpdated }
	if req.HiddenCondition != nil { item.HiddenCondition = *req.HiddenCondition }
	if req.Optional != nil { item.Optional = req.Optional }
	if req.Constraints != nil { item.Constraints = *req.Constraints }
	if req.CurrentValue != nil { item.CurrentValue = *req.CurrentValue }
	item.Modified = item.CurrentValue != item.DefaultValue
	if !item.CanBeUpdated && req.CurrentValue != nil {
		return nil, errors.New("参数无法更新")
	}
	if err := db.MySQLDB.Save(&item).Error; err != nil { return nil, err }
	return &item, nil
}

func (s *DefinitionService) Delete(ctx context.Context, vnfID, defID uint) error {
	return db.MySQLDB.Where("vnf_id = ? AND id = ?", vnfID, defID).Delete(&model.VNFDefinition{}).Error
}


