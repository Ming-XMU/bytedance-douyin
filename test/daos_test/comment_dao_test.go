package daos_test

import (
	"douyin/daos"
	"douyin/models"
	"encoding/json"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"strings"
	"testing"
	"time"
)

var commentDao daos.CommentDao

func TestListCommentById(t *testing.T) {
	comments, err := commentDao.ListCommentById(2)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(comments)
	}
}

func TestInsertComment(t *testing.T) {
	comment := models.Comment{
		VideoId:  3,
		UserId:   4,
		Message:  "hello world",
		CreateAt: time.Now(),
	}
	if err := commentDao.InsertComment(&comment); err != nil {
		fmt.Println(err)
	} else {
		data, e := json.Marshal(comment)
		if e != nil {
			fmt.Println(e)
		}
		fmt.Printf("Insert %+v success\n", string(data))
	}
}

func TestDeleteComment(t *testing.T) {
	if err := commentDao.DeleteComment(6); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Delete success")
	}
}

func init() {
	file, err := ini.Load("../../config/config.ini")
	if err != nil {
		log.Fatal(err)
		return
	}
	path := LoadMysql(file)
	fmt.Println(path)
	models.InitDB(path)

	commentDao = daos.GetCommentDao()
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
