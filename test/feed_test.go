package test

import (
	"douyin/models"
	"douyin/services"
	"gopkg.in/ini.v1"
	"log"
	"strings"
	"testing"
)

/**
 * @Author: Ember
 * @Date: 2022/5/28 16:28
 * @Description: 测试
 **/
var(
	feeService = services.GetFeedService()
)
func BenchmarkTestFeed(b *testing.B){
	for i := 0;i < b.N;i++{
		feeService.CreatVideoList(0)
	}
}

func init() {
	file, err := ini.Load("../config/config.ini")
	if err != nil {
		log.Fatal(err)
		return
	}
	path := LoadMysql(file)
	models.InitDB(path)
	LoadRedis(file)
}

func LoadMysql(file *ini.File) string {
	DbHost := file.Section("mysql").Key("DbHost").String()
	DbPort := file.Section("mysql").Key("DbPort").String()
	DbUser := file.Section("mysql").Key("DbUser").String()
	DbPassword := file.Section("mysql").Key("DbPassword").String()
	DbName := file.Section("mysql").Key("DbName").String()
	// 日期解析失败，添加参数
	return strings.Join([]string{DbUser, ":", DbPassword, "@tcp(", DbHost, ":", DbPort, ")/", DbName, "?charset=utf8&parseTime=True"}, "")
}

// LoadRedis
// 读取配置拼接Redis连接路径
func LoadRedis(file *ini.File) {
	Host := file.Section("redis").Key("Host").String()
	Port := file.Section("redis").Key("Port").String()
	RedisPass := file.Section("redis").Key("Pass").String()
	RedisUrl := strings.Join([]string{"redis://", Host, ":", Port}, "")
	models.InitRedis(RedisUrl,RedisPass)
}

