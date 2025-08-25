package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"vnf-config/internal/infra/db"
	"vnf-config/internal/model"
)

// DualStorageService 双数据库存储服务
type DualStorageService struct {
	mysqlDB *gorm.DB
	mongoDB *mongo.Client
}

func NewDualStorageService() *DualStorageService {
	return &DualStorageService{
		mysqlDB: db.MySQLDB,
		mongoDB: db.MongoDB,
	}
}

// StorageResult 存储结果
type StorageResult struct {
	MySQLSuccess bool
	MongoSuccess bool
	MySQLError   error
	MongoError   error
	Data         interface{}
}

// VNFInstanceMongo VNF实例MongoDB模型
type VNFInstanceMongo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VNFID       uint               `bson:"vnf_id" json:"vnfId"`
	Name        string             `bson:"name" json:"name"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
	YAMLConfig  interface{}        `bson:"yaml_config" json:"yamlConfig"`
	FormFields  interface{}        `bson:"form_fields" json:"formFields"`
	Metadata    map[string]interface{} `bson:"metadata" json:"metadata"`
}

// VNFDefinitionMongo VNF定义MongoDB模型
type VNFDefinitionMongo struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VNFID           uint               `bson:"vnf_id" json:"vnfId"`
	ParameterName   string             `bson:"parameter_name" json:"parameterName"`
	DefaultValue    interface{}        `bson:"default_value" json:"defaultValue"`
	DescriptionText string             `bson:"description_text" json:"descriptionTxt"`
	Type            string             `bson:"type" json:"type"`
	CanBeUpdated    bool               `bson:"can_be_updated" json:"canBeUpdated"`
	HiddenCondition string             `bson:"hidden_condition" json:"hiddenCondition"`
	Optional        *bool              `bson:"optional" json:"optional"`
	Constraints     interface{}        `bson:"constraints" json:"constraints"`
	CurrentValue    interface{}        `bson:"current_value" json:"currentValue"`
	Modified       bool               `bson:"modified" json:"modified"`
	CreatedAt      time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updatedAt"`
	Metadata       map[string]interface{} `bson:"metadata" json:"metadata"`
}

// StoreVNFInstance 存储VNF实例到双数据库
func (s *DualStorageService) StoreVNFInstance(instance *model.VNFInstance, yamlConfig *YAMLConfig) *StorageResult {
	result := &StorageResult{}

	// 存储到MySQL
	if err := s.mysqlDB.Create(instance).Error; err != nil {
		result.MySQLError = err
	} else {
		result.MySQLSuccess = true
	}

	// 存储到MongoDB
	mongoInstance := &VNFInstanceMongo{
		VNFID:      instance.ID,
		Name:       instance.Name,
		CreatedAt:  instance.CreatedAt,
		UpdatedAt:  instance.UpdatedAt,
		YAMLConfig: yamlConfig,
		FormFields: yamlConfig.Fields,
		Metadata:   yamlConfig.Metadata,
	}

	collection := db.GetMongoCollection("vnf_instances")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, mongoInstance); err != nil {
		result.MongoError = err
	} else {
		result.MongoSuccess = true
	}

	result.Data = instance
	return result
}

// StoreVNFDefinitions 存储VNF定义到双数据库
func (s *DualStorageService) StoreVNFDefinitions(definitions []model.VNFDefinition) *StorageResult {
	result := &StorageResult{}

	// 存储到MySQL
	if err := s.mysqlDB.Create(&definitions).Error; err != nil {
		result.MySQLError = err
	} else {
		result.MySQLSuccess = true
	}

	// 存储到MongoDB
	var mongoDefs []interface{}
	for _, def := range definitions {
		mongoDef := &VNFDefinitionMongo{
			VNFID:           def.VNFID,
			ParameterName:   def.ParameterName,
			DefaultValue:    def.DefaultValue,
			DescriptionText: def.DescriptionText,
			Type:            def.Type,
			CanBeUpdated:    def.CanBeUpdated,
			HiddenCondition: def.HiddenCondition,
			Optional:        def.Optional,
			Constraints:     def.Constraints,
			CurrentValue:    def.CurrentValue,
			Modified:        def.Modified,
			CreatedAt:       def.CreatedAt,
			UpdatedAt:       def.UpdatedAt,
			Metadata:        make(map[string]interface{}),
		}
		mongoDefs = append(mongoDefs, mongoDef)
	}

	if len(mongoDefs) > 0 {
		collection := db.GetMongoCollection("vnf_definitions")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := collection.InsertMany(ctx, mongoDefs); err != nil {
			result.MongoError = err
		} else {
			result.MongoSuccess = true
		}
	}

	result.Data = definitions
	return result
}

// GetVNFInstanceFromMongo 从MongoDB获取VNF实例
func (s *DualStorageService) GetVNFInstanceFromMongo(vnfID uint) (*VNFInstanceMongo, error) {
	collection := db.GetMongoCollection("vnf_instances")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var instance VNFInstanceMongo
	err := collection.FindOne(ctx, bson.M{"vnf_id": vnfID}).Decode(&instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// GetVNFDefinitionsFromMongo 从MongoDB获取VNF定义
func (s *DualStorageService) GetVNFDefinitionsFromMongo(vnfID uint) ([]VNFDefinitionMongo, error) {
	collection := db.GetMongoCollection("vnf_definitions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"vnf_id": vnfID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var definitions []VNFDefinitionMongo
	if err := cursor.All(ctx, &definitions); err != nil {
		return nil, err
	}

	return definitions, nil
}

// UpdateVNFDefinitionInMongo 在MongoDB中更新VNF定义
func (s *DualStorageService) UpdateVNFDefinitionInMongo(vnfID uint, defID uint, updates map[string]interface{}) error {
	collection := db.GetMongoCollection("vnf_definitions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates["updated_at"] = time.Now()
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"vnf_id": vnfID, "_id": defID},
		bson.M{"$set": updates},
	)
	return err
}

// DeleteVNFDefinitionFromMongo 从MongoDB删除VNF定义
func (s *DualStorageService) DeleteVNFDefinitionFromMongo(vnfID uint, defID uint) error {
	collection := db.GetMongoCollection("vnf_definitions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"vnf_id": vnfID, "_id": defID})
	return err
}

// SearchVNFInstancesInMongo 在MongoDB中搜索VNF实例
func (s *DualStorageService) SearchVNFInstancesInMongo(query map[string]interface{}) ([]VNFInstanceMongo, error) {
	collection := db.GetMongoCollection("vnf_instances")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	for key, value := range query {
		if key == "name" {
			filter[key] = bson.M{"$regex": value, "$options": "i"}
		} else {
			filter[key] = value
		}
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var instances []VNFInstanceMongo
	if err := cursor.All(ctx, &instances); err != nil {
		return nil, err
	}

	return instances, nil
}

// GetFormFieldsFromMongo 从MongoDB获取表单项
func (s *DualStorageService) GetFormFieldsFromMongo(vnfID uint) (map[string]interface{}, error) {
	instance, err := s.GetVNFInstanceFromMongo(vnfID)
	if err != nil {
		return nil, err
	}

	return instance.FormFields, nil
}

// SyncDataBetweenDatabases 在数据库之间同步数据
func (s *DualStorageService) SyncDataBetweenDatabases() error {
	log.Println("开始同步MySQL和MongoDB数据...")

	// 从MySQL获取所有VNF实例
	var instances []model.VNFInstance
	if err := s.mysqlDB.Find(&instances).Error; err != nil {
		return fmt.Errorf("从MySQL获取VNF实例失败: %v", err)
	}

	// 同步到MongoDB
	for _, instance := range instances {
		// 检查MongoDB中是否已存在
		existing, err := s.GetVNFInstanceFromMongo(instance.ID)
		if err != nil && err != mongo.ErrNoDocuments {
			log.Printf("检查VNF实例 %d 失败: %v", instance.ID, err)
			continue
		}

		if existing == nil {
			// 创建新的MongoDB记录
			mongoInstance := &VNFInstanceMongo{
				VNFID:     instance.ID,
				Name:      instance.Name,
				CreatedAt: instance.CreatedAt,
				UpdatedAt: instance.UpdatedAt,
			}

			collection := db.GetMongoCollection("vnf_instances")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if _, err := collection.InsertOne(ctx, mongoInstance); err != nil {
				log.Printf("同步VNF实例 %d 到MongoDB失败: %v", instance.ID, err)
			} else {
				log.Printf("成功同步VNF实例 %d 到MongoDB", instance.ID)
			}
		}
	}

	log.Println("数据同步完成")
	return nil
}

// GetStorageStatus 获取存储状态
func (s *DualStorageService) GetStorageStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// MySQL状态
	if s.mysqlDB != nil {
		sqlDB, err := s.mysqlDB.DB()
		if err == nil {
			status["mysql"] = map[string]interface{}{
				"connected": true,
				"stats":     sqlDB.Stats(),
			}
		} else {
			status["mysql"] = map[string]interface{}{
				"connected": false,
				"error":     err.Error(),
			}
		}
	}

	// MongoDB状态
	if s.mongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.mongoDB.Ping(ctx, nil); err != nil {
			status["mongodb"] = map[string]interface{}{
				"connected": false,
				"error":     err.Error(),
			}
		} else {
			status["mongodb"] = map[string]interface{}{
				"connected": true,
			}
		}
	}

	return status
}
