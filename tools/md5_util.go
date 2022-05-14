package tools

import (
	"bytes"
	"crypto/md5"
	"fmt"
)

//给字符串生成md5
//@params password 需要加密的字符串
//@params salt  加密的盐
//@return md5str 返回md5码
func Md5Util(password string, salt int) (CryptStr string) {

	//将salt转为字符串
	s := fmt.Sprintf("%v", salt)
	//拼接字符串
	//定义Buffer类型
	var bt bytes.Buffer
	//向bt中写入字符串
	bt.WriteString(password)
	bt.WriteString(s)
	//获得拼接后的字符串
	str := bt.String()

	//md5加密
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制

	return md5str
}
