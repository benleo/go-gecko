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
	args       map[string]interface{}
}

func (d *ScriptDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.args = args
	d.scriptFile = value.Of(args["script"]).String()
	if "" == d.scriptFile {
		log.Panic("参数[script]是必须的")
	}
}

func (d *ScriptDriver) OnStart(ctx gecko.Context) {
	d.L = NewLuaEngine()
	if err := d.L.DoFile(d.scriptFile); nil != err {
		log.Panicf("加载LUA脚本出错: %s", d.scriptFile, err)
	}
}

func (d *ScriptDriver) OnStop(ctx gecko.Context) {
	d.L.Close()
}

func (d *ScriptDriver) Drive(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket,
	deliverer gecko.OutputDeliverer, ctx gecko.Context) (out *gecko.MessagePacket, err error) {
	// Lua的函数原型： function driver(inbounds, deliverFn) (response, error)
	nArgs := setupDeliLuaFn(d.L, d.args, "driver", attrs, topic, uuid, in, deliverer)
	// 2 - Lua定义的入口main函数-返回值数量
	if err := d.L.PCall(nArgs, 2, nil); err != nil {
		log.Error("调用Lua.driver脚本发生错误: "+d.scriptFile, err)
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}), err
	}

	// 函数调用后，参数和函数全部出栈，此时栈中为函数返回值。
	retData := d.L.ToTable(1)
	retErr := d.L.ToString(2)
	d.L.Pop(2) // remove received

	if "" != retErr {
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "error",
			"error":  retErr,
		}), errors.New("LuaScript返回错误：" + retErr)
	} else {
		return gecko.NewMessagePacketFields(map[string]interface{}{
			"status": "success",
			"data":   lTableToMessage(retData),
		}), nil
	}
}
