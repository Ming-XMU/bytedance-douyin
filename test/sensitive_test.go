package test

import (
	"douyin/tools"
	"fmt"
	"testing"
)

func TestSensitiveFilter(t *testing.T) {

	//trie := tools.NewTrie()
	//trie.Add("黄色", nil)
	//trie.Add("绿色", nil)
	//trie.Add("蓝色", nil)
	//
	//result, str := trie.Check("11111", "*")
	//fmt.Printf("result:%#v, str:%v\n", result, str)
	//
	////_ = tools.Init("config/sensitive_words.txt")
	//_, isReplaced := tools.Replace("1", "*")
	//fmt.Printf("result:%#v,", isReplaced)

	err := tools.Init("../config/sensitive_words.txt")
	if err != nil {
		t.Errorf("敏感词库文件加载失败 %#v", err)
	}

	dataStr := `1111111`
	_, isReplaced := tools.Replace(dataStr, "***")
	if isReplaced == false {
		return
	}
	fmt.Printf("result:%#v,", isReplaced)

}
