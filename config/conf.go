package config

import (
	"douyin/models"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

var (
	MysqlPath string
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
	models.InitDB(MysqlPath)
}

// LoadMysql
// 读取配置拼接数据库连接路径
func LoadMysql(file *ini.File) {
	DbHost := file.Section("mysql").Key("DbHost").String()
	DbPort := file.Section("mysql").Key("DbPort").String()
	DbUser := file.Section("mysql").Key("DbUser").String()
	DbPassword := file.Section("mysql").Key("DbPassword").String()
	DbName := file.Section("mysql").Key("DbName").String()
	MysqlPath = strings.Join([]string{DbUser, ":", DbPassword, "@tcp(", DbHost, ":", DbPort, ")/", DbName}, "")
}
