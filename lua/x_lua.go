package lua

import (
	"github.com/cjoudrey/gluahttp"
	glua "github.com/yuin/gopher-lua"
	"net/http"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

const (
	// BaseLibName is here for consistency; the base functions have no namespace/library.
	BaseLibName = ""
	// LoadLibName is here for consistency; the loading system has no namespace/library.
	LoadLibName = "package"
	// TabLibName is the name of the table Library.
	TabLibName = "table"
	// IoLibName is the name of the io Library.
	IoLibName = "io"
	// OsLibName is the name of the os Library.
	OsLibName = "os"
	// StringLibName is the name of the string Library.
	StringLibName = "string"
	// MathLibName is the name of the math Library.
	MathLibName = "math"
)

type luaLib struct {
	libName string
	libFunc glua.LGFunction
}

var luaLibs = []luaLib{
	{LoadLibName, glua.OpenPackage},
	{BaseLibName, glua.OpenBase},
	{TabLibName, glua.OpenTable},
	{IoLibName, glua.OpenIo},
	{OsLibName, glua.OpenOs},
	{StringLibName, glua.OpenString},
	{MathLibName, glua.OpenMath},
}

func NewLuaEngine() *glua.LState {
	// 简化Lua虚拟机
	ls := glua.NewState(glua.Options{
		CallStackSize: 8,
		SkipOpenLibs:  true,
	})
	ls.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{
		Timeout: time.Second * 10,
	}).Loader)
	for _, lib := range luaLibs {
		ls.Push(ls.NewFunction(lib.libFunc))
		ls.Push(glua.LString(lib.libName))
		ls.Call(1, 0)
	}
	return ls
}
