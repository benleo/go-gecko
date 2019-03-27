package utils

import "github.com/parkingwang/go-conf"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func AnyToInt64(any interface{}) (int64, bool) {
	switch any.(type) {
	case int:
		return int64(any.(int)), true
	case uint8:
		return int64(any.(uint8)), true
	case uint16:
		return int64(any.(uint16)), true
	case uint32:
		return int64(any.(uint32)), true
	case int64:
		return any.(int64), true
	case string:
		v, err := cfg.Value(any.(string)).Int64()
		if nil != err {
			return 0, false
		} else {
			return v, true
		}
	default:
		return 0, false
	}
}
