package core

import (
	"fmt"
	"os"
	"panel_backend/global"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitMysql() *gorm.DB {
	if global.Config.Mysql.Host == "" {
		global.Log.Error("未配置mysql主机，取消数据库初始化")
		os.Exit(1)//退出程序
	}
	dsn := global.Config.Mysql.GetDsn()
	var  mysqlLogger logger.Interface
	var logLevel logger.LogLevel
	switch global.Config.Mysql.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:	
		logLevel = logger.Info
	}
	// 设置数据库日志显式等级
	mysqlLogger = logger.Default.LogMode(logLevel)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		global.Log.Error(fmt.Sprintf("[%s]mysql数据库连接失败[%s]", dsn, err.Error()))
		os.Exit(1)//程序异常退出
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)//设置最大空闲连接数
	sqlDB.SetMaxOpenConns(100)//设置最大连接数
	global.Log.Debugf("mysql数据库连接成功[%s]\n", dsn)
	return db
}