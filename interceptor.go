package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Interceptor interface {
	InitialAware

	GetPriority() int
	SetPriority(p int)

	// 拦截处理过程。抛出 {@link DropException} 来中断拦截。
	Handle(ctx GeckoContext, scoped GeckoScoped) error
}
