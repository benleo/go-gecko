package lua

import (
	"errors"
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-value"
	"github.com/yuin/gopher-lua"
)

//
// Author: 陈哈哈 yoojiachen@gmail.com
//

func ScriptDriverFactory() (string, gecko.Factory) {
	return "ScriptDriver", func() interface{} {
		return NewScriptDriver()
	}
}

func NewScriptDriver() *ScriptDriver {
	return &ScriptDriver{
		AbcDriver: gecko.NewAbcDriver(),
	}
}

// Lua脚本驱动，
type ScriptDriver struct {
	*gecko.AbcDriver
	gecko.LifeCycle
	scriptFile string
	L          *lua.LState
}

func (d *ScriptDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.scriptFile = value.Of(args["script"]).String()
	if "" == d.scriptFile {
		log.Panic("Arg[script] is required")
	}
}

func (d *ScriptDriver) OnStart(ctx gecko.Context) {
	d.L = lua.NewState()
	if err := d.L.DoFile(d.scriptFile); nil != err {
		log.Panicf("Failed to load lua script: %s", d.scriptFile, err)
	}
}

func (d *ScriptDriver) OnStop(ctx gecko.Context) {
	d.L.Close()
}

func (d *ScriptDriver) Drive(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket, deliverer gecko.OutputDeliverer, ctx gecko.Context) (out *gecko.MessagePacket, err error) {
	// 调用Lua脚本
	L := d.L
	// 先函数，后参数，正序入栈:
	// Lua的函数原型： function driver(request, deliverFn) (response, error)
	// 先压入函数
	L.Push(L.GetGlobal("driver"))
	// Param 1
	req := L.CreateTable(0, 4) // 0 arr, 4 Hash
	req.RawSet(lua.LString("attrs"), mapToLTable(attrs.Map()))
	req.RawSet(lua.LString("topic"), lua.LString(topic))
	req.RawSet(lua.LString("uuid"), lua.LString(uuid))
	req.RawSet(lua.LString("inbound"), messageToLTable(in))
	L.Push(req)
	// Param 2
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

	// #1 2 - Lua定义的入口main函数-参数数量
	// #2 2 - Lua定义的入口main函数-返回值数量
	if err := L.PCall(2, 2, nil); err != nil {
		log.Error("调用Lua脚本发生错误", err)
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}), err
	}

	// 函数调用后，参数和函数全部出栈，此时栈中为函数返回值。
	// 负数表示反序读取，-1为栈顶；
	lResp := L.ToTable(1)
	lErr := L.ToString(2)
	L.Pop(2) // remove received

	if "" != lErr {
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "error",
			"error":  "LuaScript返回非法的数据格式",
		}), errors.New("LuaScript返回非法的数据格式")
	} else {
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "success",
			"data":   lTableToMessage(lResp),
		}), nil
	}
}
