package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(config interface{}) error {
	// 简化版本，直接使用开发配置
	// 生产环境可以使用更复杂的配置
	configZap := zap.NewDevelopmentConfig()
	configZap.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	
	var err error
	Logger, err = configZap.Build()
	if err != nil {
		return err
	}
	
	return nil
}

func GetLogger() *zap.Logger {
	if Logger == nil {
		InitLogger(nil)
	}
	return Logger
}

