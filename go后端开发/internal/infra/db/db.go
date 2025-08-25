package db

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"vnf-config/internal/model"
)

var (
	MySQLDB *gorm.DB
	MongoDB *mongo.Client
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MySQLDSN      string
	MongoURI      string
	MongoDatabase string
	MaxOpen       int
	MaxIdle       int
}

// Init 初始化双数据库连接
func Init() error {
	config := &DatabaseConfig{
		MySQLDSN:      os.Getenv("MYSQL_DSN"),
		MongoURI:      os.Getenv("MONGO_URI"),
		MongoDatabase: os.Getenv("MONGO_DATABASE"),
		MaxOpen:       20,
		MaxIdle:       10,
	}

	// 设置默认值
	if config.MySQLDSN == "" {
		config.MySQLDSN = "root:password@tcp(127.00.1:3306)/vnf_config?charset=utf8mb4&parseTime=True&loc=Local"
	}
	if config.MongoURI == "" {
		config.MongoURI = "mongodb://localhost:27017"
	}
	if config.MongoDatabase == "" {
		config.MongoDatabase = "vnf_config"
	}

	// 初始化MySQL
	if err := initMySQL(config); err != nil {
		return err
	}

	// 初始化MongoDB
	if err := initMongoDB(config); err != nil {
		return err
	}

	log.Println("双数据库初始化完成")
	return nil
}

// initMySQL 初始化MySQL连接
func initMySQL(config *DatabaseConfig) error {
	logMode := logger.Silent
	if os.Getenv("APP_ENV") == "development" {
		logMode = logger.Info
	}

	database, err := gorm.Open(mysql.Open(config.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpen)
	sqlDB.SetMaxIdleConns(config.MaxIdle)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)

	// 自动迁移MySQL表结构
	if err := database.AutoMigrate(&model.VNFInstance{}, &model.VNFDefinition{}); err != nil {
		return err
	}

	MySQLDB = database
	log.Println("MySQL数据库初始化完成")
	return nil
}

// initMongoDB 初始化MongoDB连接
func initMongoDB(config *DatabaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		return err
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	MongoDB = client
	log.Println("MongoDB数据库初始化完成")
	return nil
}

// GetMongoCollection 获取MongoDB集合
func GetMongoCollection(collectionName string) *mongo.Collection {
	database := os.Getenv("MONGO_DATABASE")
	if database == "" {
		database = "vnf_config"
	}
	return MongoDB.Database(database).Collection(collectionName)
}

// Close 关闭数据库连接
func Close() {
	if MySQLDB != nil {
		if sqlDB, err := MySQLDB.DB(); err == nil {
			sqlDB.Close()
		}
	}
	if MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		MongoDB.Disconnect(ctx)
	}
}

func defaultString(v string, d string) string {
	if v == "" {
		return d
	}
	return v
}


