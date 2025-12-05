package client

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go.uber.org/zap"

	"github.com/hoj/hist-oj/internal/config"
	"github.com/hoj/hist-oj/internal/model"
	"github.com/hoj/hist-oj/internal/utils"
)

var DB *gorm.DB

func InitDatabase(cfg *config.DatabaseConfig) error {
	dsn := cfg.DSN()
	
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 使用默认连接池配置
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		utils.GetLogger().Error("数据库连接测试失败", zap.Error(err))
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db

	// 初始化表结构
	if err := model.InitTables(db); err != nil {
		utils.GetLogger().Warn("初始化数据库表失败", zap.Error(err))
		// 不返回错误，允许表已存在的情况
	} else {
		utils.GetLogger().Info("数据库表初始化成功")
	}

	utils.GetLogger().Info("数据库连接成功", 
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName))
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

