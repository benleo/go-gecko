package lua

import (
	"github.com/yoojia/go-gecko"
	"github.com/yuin/gopher-lua"
)

func packToTable(pack *gecko.MessagePacket, l *lua.LState) *lua.LTable {
	t := l.NewTable()
	return t
}

func tableToPack(table *lua.LTable, l *lua.LState) *gecko.MessagePacket {
	return nil
}
