package lua

import (
	"errors"
	"github.com/yoojia/go-gecko"
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
	script string
	engine *lua.LState
}

func (d *ScriptDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.script = value.Of(args["script"]).String()
	if "" == d.script {
		log.Panic("Arg[script] is required")
	}
}

func (d *ScriptDriver) OnStart(ctx gecko.Context) {
	d.engine = lua.NewState()
	if err := d.engine.DoFile(d.script); nil != err {
		log.Panicf("Failed to load lua script: %s", d.script, err)
	}
}

func (d *ScriptDriver) OnStop(ctx gecko.Context) {
	d.engine.Close()
}

func (d *ScriptDriver) Handle(session gecko.EventSession, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	// 创建Lua调用的函数
	fn := func(l *lua.LState) int {
		// get args
		uuid := l.ToString(1)
		message := l.ToTable(2)
		// process, and set back to lua
		if ret, err := deliverer.Deliver(uuid, tableToPack(message, l)); nil != err {
			log.Error("ScriptDriver.deliver发生错误", err)
			l.Push(lua.LNil)
			l.Push(lua.LString(err.Error()))
		} else {
			l.Push(packToTable(ret, l))
			l.Push(lua.LNil)
		}
		return 2 // Lua函数返回2个结果，下同
	}

	outbound := session.Outbound()

	// 调用Lua脚本
	if err := d.engine.CallByParam(lua.P{
		Fn:      d.engine.GetGlobal("main"),
		NRet:    2, // Lua函数，返回2个结果， 同上
		Protect: true,
	},
		// 第一个参数： session
		d.toTable(session, d.engine),
		// 第二个参数： deliverer 函数
		d.engine.NewFunction(fn),
	); err != nil {
		log.Error("调用Lua脚本发生错误", err)
		outbound.AddField("status", "error")
		outbound.AddField("error", err.Error())
		return err
	}

	// returned values
	lRet := d.engine.Get(-1)
	lErr := d.engine.Get(-2)
	d.engine.Pop(2) // remove received

	if serr, ok := lErr.(lua.LString); ok && "" != serr {
		outbound.AddField("status", "error")
		outbound.AddField("error", serr.String())
		return errors.New(serr.String())
	}

	if table, ok := lRet.(*lua.LTable); !ok {
		return errors.New("LuaScript返回非法的数据格式: NOT_LTABLE")
	} else {
		outbound.AddField("status", "success")
		outbound.AddField("data", tableToPack(table, d.engine))
		return nil
	}
}

func (d *ScriptDriver) toTable(session gecko.EventSession, l *lua.LState) *lua.LTable {
	table := l.NewTable()
	table.RawSet(lua.LString("attributes"), goMapToLuaTable(session.Attrs()))
	table.RawSet(lua.LString("timestamp"), lua.LNumber(session.Timestamp().Unix()))
	table.RawSet(lua.LString("topic"), lua.LString(session.Topic()))
	table.RawSet(lua.LString("uuid"), lua.LString(session.Uuid()))
	table.RawSet(lua.LString("inbound"), packToTable(session.Inbound(), l))
	return table
}
