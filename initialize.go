package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type NeedInitialize interface {
	// 指定参数初始化组件
	OnInit(args map[string]interface{}, scoped GeckoScoped)
}
