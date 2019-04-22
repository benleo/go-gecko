package lua

import (
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-value"
	"github.com/yuin/gopher-lua"
)

func messageToLTable(pack *gecko.MessagePacket) *lua.LTable {
	return mapToLTable(pack.GetFields())
}

func lTableToMessage(table *lua.LTable) *gecko.MessagePacket {
	fields := make(map[string]interface{})
	table.ForEach(func(key lua.LValue, val lua.LValue) {
		k := key.String()
		switch val.Type() {
		case lua.LTNumber:
			fields[k] = value.Of(val).MustFloat64()

		case lua.LTString:
			fields[k] = value.Of(val).String()

		case lua.LTBool:
			fields[k] = value.Of(val).MustBool()
		}
	})
	return gecko.NewMessagePacketFields(fields)
}
