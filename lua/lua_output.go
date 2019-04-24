package lua

import (
	"errors"
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-value"
	"github.com/yuin/gopher-lua"
)

func ScriptOutputFactory() (string, gecko.Factory) {
	return "ScriptOutput", func() interface{} {
		return &ScriptOutput{
			AbcOutputDevice: gecko.NewAbcOutputDevice(),
		}
	}
}

type ScriptOutput struct {
	*gecko.AbcOutputDevice
	scriptFile string
	L          *lua.LState
	args       map[string]interface{}
}

func (d *ScriptOutput) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.args = args
	d.scriptFile = value.Of(args["script"]).String()
	if "" == d.scriptFile {
		log.Panic("参数[script]是必须的")
	}
}

func (d *ScriptOutput) OnStart(ctx gecko.Context) {
	d.L = NewLuaEngine()
	if err := d.L.DoFile(d.scriptFile); nil != err {
		log.Panicf("加载LUA脚本出错: %s", d.scriptFile, err)
	}
}

func (d *ScriptOutput) OnStop(ctx gecko.Context) {
	d.L.Close()
}

func (d *ScriptOutput) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	// Lua的函数原型： function processMain(argsTable, frame) (string, error)
	// 先函数，后参数，正序入栈:
	d.L.Push(d.L.GetGlobal("outputMain"))
	// Arg 1
	d.L.Push(mapToLTable(d.args))
	// Arg 2
	d.L.Push(lua.LString(string(frame)))

	// 2 - Lua定义的入口main函数-返回值数量
	if err := d.L.PCall(2, 2, nil); err != nil {
		log.Error("调用 Lua.process 脚本发生错误: "+d.scriptFile, err)
		return nil, err
	}

	// 函数调用后，参数和函数全部出栈，此时栈中为函数返回值。
	ret := d.L.ToString(1)
	err := d.L.ToString(2)
	d.L.Pop(2) // remove received
	if "" != err {
		return nil, errors.New("LuaScript返回错误：" + err)
	} else {
		ctx.OnIfLogV(func() {
			log.Debug("LuaScript返回结果: " + ret)
		})
		return []byte(ret), nil
	}
}
