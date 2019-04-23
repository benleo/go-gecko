package lua

import (
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-value"
	"github.com/yuin/gopher-lua"
)

func messageToLTable(pack *gecko.MessagePacket) *lua.LTable {
	if len(pack.GetFrames()) > 0 {
		log.Error("Lua脚本组件暂不支持MessagePacket.Frames字段")
	}
	return mapToLTable(pack.GetFields())
}

func lTableToMessage(table *lua.LTable) *gecko.MessagePacket {
	if nil == table {
		return gecko.NewMessagePacketFields(make(map[string]interface{}, 0))
	}
	data := make(map[string]interface{}, table.Len())
	table.ForEach(func(key lua.LValue, val lua.LValue) {
		k := key.String()
		switch val.Type() {
		case lua.LTNumber:
			data[k] = value.Of(val).MustFloat64()

		case lua.LTString:
			data[k] = value.Of(val).String()

		case lua.LTBool:
			data[k] = value.Of(val).MustBool()
		}
	})
	return gecko.NewMessagePacketFields(data)
}
