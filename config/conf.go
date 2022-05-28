package config

import (
	"douyin/controller"
	"douyin/models"
	"douyin/mq"
	"douyin/services"
	"douyin/tools"
	"fmt"
	"github.com/anqiansong/ketty/console"
	"gopkg.in/ini.v1"
	"log"
	"strings"
)

var (
	MysqlPath string
	RedisUrl  string
	MQUrl     string
	RedisPass string
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
	LoadRabbitMQ(file)
	//开启mq监听
	mq.InitMQ(MQUrl)
	FollowQueueListen()
	models.InitDB(MysqlPath)
	models.InitRedis(RedisUrl, RedisPass)
	initService()
	//load sensitive
	err = tools.Init("config/sensitive_words.txt")
	if err != nil{
		console.Error(err)
		panic(err.Error())
	}
}

// LoadMysql
// 读取配置拼接数据库连接路径
func LoadMysql(file *ini.File) {
	DbHost := file.Section("mysql").Key("DbHost").String()
	DbPort := file.Section("mysql").Key("DbPort").String()
	DbUser := file.Section("mysql").Key("DbUser").String()
	DbPassword := file.Section("mysql").Key("DbPassword").String()
	DbName := file.Section("mysql").Key("DbName").String()
	MysqlPath = strings.Join([]string{DbUser, ":", DbPassword, "@tcp(", DbHost, ":", DbPort, ")/", DbName, "?charset=utf8&parseTime=True&loc=Local"}, "")
}

// LoadRedis
// 读取配置拼接Redis连接路径
func LoadRedis(file *ini.File) {
	Host := file.Section("redis").Key("Host").String()
	Port := file.Section("redis").Key("Port").String()
	RedisPass = file.Section("redis").Key("Pass").String()
	RedisUrl = strings.Join([]string{"redis://", Host, ":", Port}, "")
}

// LoadRabbitMQ
// 读取配置拼接mq连接路径
func LoadRabbitMQ(file *ini.File) {
	MqName := file.Section("rabbitmq").Key("MqName").String()
	MqPassword := file.Section("rabbitmq").Key("MqPassword").String()
	Host := file.Section("rabbitmq").Key("Host").String()
	Port := file.Section("rabbitmq").Key("Port").String()
	DefaultHost := file.Section("rabbitmq").Key("Default_Host").String()
	MQUrl = strings.Join([]string{"amqp://", MqName, ":", MqPassword, "@", Host, ":", Port, "/", DefaultHost}, "")
	fmt.Println(MQUrl)
}

// 初始化service
func initService() {
	controller.FeeSerivce = services.GetFeedService()
	controller.UserSerivce = services.GetUserService()
	controller.FollowSerivce = services.GetFollowService()
	controller.FavouriteService = services.GetFavoriteService()
}

func FollowQueueListen() {
	rabbitmq := mq.GetFollowMQ()
	rabbitmq.ConsumeSimple(func(msg string) error {
		//按字符_分割参数
		arr := strings.Split(msg, "_")
		err := controller.FollowSerivce.Action(arr[0], arr[1], arr[2])
		if err != nil {
			return err
		}
		return nil
	})
}
