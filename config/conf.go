package config

import (
	"douyin/controller"
	"douyin/models"
	"douyin/services"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

var (
	MysqlPath string
	RedisUrl  string
)

// Init
//初始化各个组件
func Init() {
	file, err := ini.Load("config/config.ini")
	if err != nil {
		log.Fatal(err)
		return
	}
	LoadMysql(file)
	LoadRedis(file)
	models.InitDB(MysqlPath)
	models.InitRedis(RedisUrl)
	initService()
}

// LoadMysql
// 读取配置拼接数据库连接路径
func LoadMysql(file *ini.File) {
	DbHost := file.Section("mysql").Key("DbHost").String()
	DbPort := file.Section("mysql").Key("DbPort").String()
	DbUser := file.Section("mysql").Key("DbUser").String()
	DbPassword := file.Section("mysql").Key("DbPassword").String()
	DbName := file.Section("mysql").Key("DbName").String()
	MysqlPath = strings.Join([]string{DbUser, ":", DbPassword, "@tcp(", DbHost, ":", DbPort, ")/", DbName, "?charset=utf8&parseTime=True"}, "")
}

// LoadRedis
// 读取配置拼接Redis连接路径
func LoadRedis(file *ini.File) {
	Host := file.Section("redis").Key("Host").String()
	Port := file.Section("redis").Key("Port").String()
	RedisUrl = strings.Join([]string{"redis://", Host, ":", Port}, "")
}

// 初始化service
func initService() {
	controller.FeeSerivce = services.GetFeedService()
	controller.UserSerivce = services.GetUserService()
}
