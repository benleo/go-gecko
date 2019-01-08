package gecko

import "time"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type GeckoScoped interface {
	// 检查操作是否超时
	CheckTimeout(timeout time.Duration, action func())
}

///
