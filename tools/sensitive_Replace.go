package tools

// 敏感词替换
func Replace(text string, replace string) (result string, isReplaced bool) {
	result, isReplaced = tril.Check(text, replace)
	return
}
