package utils

import "container/list"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func ForEach(list *list.List, consumer func(it interface{})) {
	for el := list.Front(); el != nil; el = el.Next() {
		consumer(el.Value)
	}
}
