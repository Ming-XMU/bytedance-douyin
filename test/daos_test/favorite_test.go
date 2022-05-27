package daos

import (
	"douyin/daos"
	"douyin/models"
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"strings"
	"testing"
)

/**
 * @Author: Ember
 * @Date: 2022/5/27 21:21
 * @Description: TODO
 **/

var(
	favoriteDao daos.FavoriteDao
)
func TestGetUserFavorites(t *testing.T){
	favorites, err := favoriteDao.UserFavorites(2)
	if err != nil{
		log.Fatal(err)
	}
	for _,favorite := range(favorites){
		fmt.Println(favorite)
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
	favoriteDao = daos.GetFavoriteDao()
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
