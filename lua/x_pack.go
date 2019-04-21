package lua

import (
	"github.com/yoojia/go-gecko"
	"github.com/yuin/gopher-lua"
)

func messageToLTable(pack *gecko.MessagePacket, l *lua.LState) *lua.LTable {
	t := l.NewTable()
	return t
}

func lTableToMessage(table *lua.LTable, l *lua.LState) *gecko.MessagePacket {
	return nil
}
