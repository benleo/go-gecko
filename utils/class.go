package utils

import (
	"reflect"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func GetClassName(obj interface{}) string {
	return GetTypeName(reflect.TypeOf(obj))
}

func GetTypeName(typed reflect.Type) string {
	name := typed.String()
	if dotIdx := strings.LastIndex(name, "."); dotIdx >= 0 {
		name = name[dotIdx+1:]
	}
	return name
}
