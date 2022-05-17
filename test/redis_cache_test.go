package test

import (
	"douyin/models"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"strconv"
	"testing"
)

/**
 * @Author: Ember
 * @Date: 2022/5/17 14:56
 * @Description: TODO
 **/

func TestJsonFeeCache(t *testing.T){

	rec, err := redis.Dial("tcp","120.78.238.68:6379")
	result, err := redis.Values(rec.Do("Lrange", "video_cache", 0, -1))
	if err != nil{
		fmt.Println(err)
		return
	}
	for _,val := range(result){
		var video models.Video
		json.Unmarshal(val.([]byte),&video)
		fmt.Println(video)
	}
}
//序列化
func Serialization(value interface{}) ([]byte, error) {
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Map:
	}
	k, err := json.Marshal(value)
	return k, err
}

func Deserialization(byt []byte, ptr interface{}) (err error) {
	if bytes, ok := ptr.(*[]byte); ok {
		*bytes = byt
		return
	}
	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		switch p := v.Elem(); p.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var i int64
			i, err = strconv.ParseInt(string(byt), 10, 64)
			if err != nil {
				fmt.Printf("Deserialization: failed to parse int '%s': %s", string(byt), err)
			} else {
				p.SetInt(i)
			}
			return

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var i uint64
			i, err = strconv.ParseUint(string(byt), 10, 64)
			if err != nil {
				fmt.Printf("Deserialization: failed to parse uint '%s': %s", string(byt), err)
			} else {
				p.SetUint(i)
			}
			return
		}
	}
	err = json.Unmarshal(byt, &ptr)
	return
}