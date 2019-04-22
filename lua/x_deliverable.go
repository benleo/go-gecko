package lua

import (
	"github.com/yoojia/go-gecko/v2"
	"github.com/yuin/gopher-lua"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func setupDeliLuaFn(
	L *lua.LState,
	args map[string]interface{},
	fn string,
	attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket,
	deliverer gecko.OutputDeliverer) (nArgs int) {
	// 先函数，后参数，正序入栈:
	// 先压入函数
	L.Push(L.GetGlobal(fn))
	// Param 1
	L.Push(mapToLTable(args))
	// Param 2
	req := L.CreateTable(0, 4) // 0 arr, 4 Hash
	req.RawSet(lua.LString("attrs"), mapToLTable(attrs.Map()))
	req.RawSet(lua.LString("topic"), lua.LString(topic))
	req.RawSet(lua.LString("uuid"), lua.LString(uuid))
	req.RawSet(lua.LString("inbound"), messageToLTable(in))
	L.Push(req)
	// Param 3
	L.Push(L.NewFunction(func(l *lua.LState) int {
		// 为Lua注入的deliver函数，
		// 原型为： function deliver(uuid, message) (message, error)
		// Lua调用传递的参数：
		uuid := l.ToString(1)
		message := l.ToTable(2)
		// Go处理，并返回结果到Lua中：
		if ret, err := deliverer.Deliver(uuid, lTableToMessage(message)); nil != err {
			log.Error("ScriptDriver.deliver发生错误", err)
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(messageToLTable(ret))
			l.Push(lua.LNil)
		}
		// Lua函数返回2个结果。此为 deliver 函数的结果数量。
		return 2
	}))

	return 3 // 封装的函数参数个数
}
