package tools

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var (
	tril *Trie
)

// Init 过滤器初始化
func Init(filename string) (err error) {
	//
	tril = NewTrie()
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("敏感词字典文件读取失败")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logrus.Errorln(err.Error())
		}
	}(file)
	// 从敏感词文件里面读取数据生成敏感词数据库
	reader := bufio.NewReader(file)
	for {
		word, errRec := reader.ReadString('\n')
		// 读取到最后一行 读取成功了
		if errRec == io.EOF {
			return
		}
		if errRec != nil {
			err = errRec
			return
		}

		// 把读出的单词加入到敏感词库
		err = tril.Add(word, nil)

		if err != nil {
			return
		}
	}
}
