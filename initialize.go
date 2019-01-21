package gecko

import "github.com/parkingwang/go-conf"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Initialize interface {
	// 指定参数初始化组件
	OnInit(config *cfg.Config, context Context)
}
