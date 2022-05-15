package test

import (
	"douyin/services"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"testing"
)

func TestTopicPublishInfo(t *testing.T) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/douyin"
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		log.Panicln(err)
	}
	username := "root"
	password := "123456"
	_, err = services.GetUserService(db).UserLogin(username, password)
	if err != nil {
		log.Panicln(err)
	}

}
