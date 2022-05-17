package daos_test

import (
	"douyin/daos"
	"fmt"
	"testing"
)

func TestFindById(t *testing.T) {
	dao := daos.GetUserDao()
	user, err := dao.FindById(1)
	if err != nil {
		fmt.Println(user)
	}
}
