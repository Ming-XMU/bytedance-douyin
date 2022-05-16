package tools

/**
 * @Author: Ember
 * @Date: 2022/5/16 10:44
 * @Description: TODO
 **/

//校验字符串参数
func VerifyParamsEmpty(strs ...string) bool{
	for _,str := range strs{
		if emptyVerify(str){
			return true
		}
	}
	return false
}
func emptyVerify(str string) bool{
	if str == ""{
		return true
	}
	return false
}