package config

import (
	"douyin/controller"
	"douyin/daos"
	"douyin/models"
	"douyin/mq"
	"douyin/services"
	"douyin/tools"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anqiansong/ketty/console"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	MysqlPath       string
	RedisUrl        string
	MQUrl           string
	RedisPass       string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
)

// Init
//初始化各个组件
func Init() {
	initLogRus()
	file, err := ini.Load("config/config.ini")
	if err != nil {
		logrus.Fatalln(err)
	}
	LoadMysql(file)
	LoadRedis(file)
	LoadRabbitMQ(file)
	LoadMinIO(file)
	//开启mq监听
	mq.InitMQ(MQUrl)
	followQueueListen()
	favoriteQueueListen()
	models.InitDB(MysqlPath)
	models.InitRedis(RedisUrl, RedisPass)
	models.InitMinio(Endpoint, AccessKeyID, SecretAccessKey)

	initService()
	//load sensitive
	err = tools.Init("config/sensitive_words.txt")
	if err != nil {
		console.Error(err)
		logrus.Panicln(err.Error())
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

// LoadMinIO
// 读取配置拼接minIO连接路径
func LoadMinIO(file *ini.File) {
	AccessKeyID = file.Section("minio").Key("AccessKeyID").String()
	SecretAccessKey = file.Section("minio").Key("SecretAccessKey").String()
	Host := file.Section("minio").Key("Host").String()
	Port := file.Section("minio").Key("Port").String()
	Endpoint = strings.Join([]string{Host, ":", Port}, "")
}

// 初始化service
func initService() {
	controller.FeeSerivce = services.GetFeedService()
	controller.UserSerivce = services.GetUserService()
	controller.FollowSerivce = services.GetFollowService()
	controller.FavouriteService = services.GetFavoriteService()
}

func followQueueListen() {
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

func favoriteQueueListen() {
	rabbitMQSimple := mq.GetFavoriteMQ()
	rabbitMQSimple.ConsumeSimple(func(msg string) error {
		var fMsg mq.FavoriteActionMsg
		err := json.Unmarshal([]byte(msg), &fMsg)
		if err != nil {
			return errors.New("favoriteQueueListen unmarshal failed")
		}
		if fMsg.Action == 1 {
			err = daos.GetFavoriteDao().InsertFavorite(fMsg.Favorite)
			if err != nil {
				console.Error(err)
				return errors.New("点赞失败")
			}
		}
		if fMsg.Action == 2 {
			err = daos.GetFavoriteDao().DeleteFavorite(fMsg.Favorite.UserId, fMsg.Favorite.VideoId)
			if err != nil {
				console.Error(err)
				return errors.New("取消点赞失败")
			}
		}
		return nil
	})
}

//日志logrus
func initLogRus() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			//处理文件名
			fileName := path.Base(frame.File)
			return frame.Function, fileName
		},
	})
	std := os.Stdout
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("create file log.txt failed: %v", err)
	}
	logrus.SetOutput(io.MultiWriter(std, file))
	logrus.Infoln("application is run")
}
