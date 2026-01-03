package infrastructure

import (
	"errors"
	configstruct "github.com/zhanshen02154/order/internal/config"
	"go-micro.dev/v4/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"time"
)

// InitDB 初始化数据库
func InitDB(confInfo *configstruct.MySqlConfig, zapLogger gormlogger.Interface) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       confInfo.Dsn,
		SkipInitializeWithVersion: false,
		DefaultStringSize:         50,
	}), &gorm.Config{SkipDefaultTransaction: true, Logger: zapLogger})
	if err != nil {
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if sqlDB == nil {
		return nil, errors.New("获取SQL DB失败")
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(confInfo.MaxOpenConns)
	sqlDB.SetMaxIdleConns(confInfo.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(confInfo.ConnMaxLifeTime) * time.Second)

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.New("数据库连接验证失败: " + err.Error())
	}

	logger.Info("数据库连接成功")
	return db, nil
}
