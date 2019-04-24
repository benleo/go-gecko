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

func ScriptTriggerFactory() (string, gecko.Factory) {
	return "ScriptTrigger", func() interface{} {
		return NewScriptTrigger()
	}
}

func NewScriptTrigger() *ScriptTrigger {
	return &ScriptTrigger{
		AbcTrigger: gecko.NewAbcTrigger(),
	}
}

// Lua脚本触发器，通过外置Lua脚本，执行其 trigger 函数。
type ScriptTrigger struct {
	*gecko.AbcTrigger
	gecko.LifeCycle
	scriptFile string
	L          *lua.LState
	args       map[string]interface{}
}

func (d *ScriptTrigger) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.args = args
	d.scriptFile = value.Of(args["script"]).String()
	if "" == d.scriptFile {
		log.Panic("参数[script]是必须的")
	}
}

func (d *ScriptTrigger) OnStart(ctx gecko.Context) {
	d.L = NewLuaEngine()
	if err := d.L.DoFile(d.scriptFile); nil != err {
		log.Panicf("加载Lua脚本出错: %s", d.scriptFile, err)
	}
}

func (d *ScriptTrigger) OnStop(ctx gecko.Context) {
	d.L.Close()
}

func (d *ScriptTrigger) Touch(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket,
	deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	// Lua的函数原型： function triggerMain(args, inbounds, deliverFn) error
	nArgs := setupDeliLuaFn(d.L, d.args, "triggerMain", attrs, topic, uuid, in, deliverer)
	// 2 - Lua定义的入口main函数-返回值数量
	if err := d.L.PCall(nArgs, 1, nil); err != nil {
		log.Error("调用Lua.trigger脚本发生错误: "+d.scriptFile, err)
		return err
	}
	// 函数调用后，参数和函数全部出栈，此时栈中为函数返回值。
	retErr := d.L.ToString(1)
	d.L.Pop(1) // remove received
	if "" != retErr {
		return errors.New("LuaScript返回错误:" + retErr)
	} else {
		return nil
	}
}
