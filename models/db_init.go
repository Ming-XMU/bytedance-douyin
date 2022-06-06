package models

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"os"
	"time"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func InitDB(con string) {
	//日志配置
	slowLogger := logger.New(
		//设置Logger
		NewMyWriter(),
		logger.Config{
			//慢SQL阈值
			SlowThreshold: time.Millisecond * 100,
			//设置日志级别
			LogLevel: logger.Warn,
		},
	)
	//创建连接
	open, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       con,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{Logger: slowLogger})
	if err != nil {
		logrus.Fatalln(err)
	}
	sqlDB, err := open.DB()
	if err != nil {
		logrus.Fatalln(err)
	}
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)
	db = open
}

//logrus日志,记录慢sql
type MyWriter struct {
	mysqlLog *logrus.Logger
}

//实现gorm/logger.Writer接口
func (m *MyWriter) Printf(format string, v ...interface{}) {
	logstr := fmt.Sprintf(format, v...)
	//利用loggus记录日志
	m.mysqlLog.Info(logstr)
}

func NewMyWriter() *MyWriter {
	log := logrus.New()
	//配置logrus
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	std := os.Stdout
	file, err := os.OpenFile("./sql.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		logrus.Errorln("create file sql.txt failed: %v", err)
	}
	log.SetOutput(io.MultiWriter(std, file))
	return &MyWriter{mysqlLog: log}
}
