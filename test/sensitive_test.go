package test

import (
	"douyin/tools"
	"fmt"
	"testing"
)

func TestSensitiveFilter(t *testing.T) {

	trie := tools.NewTrie()
	trie.Add("黄色", nil)
	trie.Add("绿色", nil)
	trie.Add("蓝色", nil)

	result, str := trie.Check("黄色阿斯顿后i啊和爱德华尽快哈红色按时到货就哦啊包含绿色", "*")

	fmt.Printf("result:%#v, str:%v\n", result, str)

}
