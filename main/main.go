package main

import (
	"bytes"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Main
func main() {
	conf := loadConfig("conf.d")
	if len(conf) <= 0 {
		withTag(log.Panic).Msgf("Config is empty")
	}

	engine := gecko.SharedEngine()
	engine.PrepareEnv()
	engine.Init(conf)
	engine.Start()
	defer engine.Stop()
	// 等待终止信号
	sysSignal := make(chan os.Signal, 1)
	signal.Notify(sysSignal, syscall.SIGINT, syscall.SIGTERM)
	<-sysSignal
	withTag(log.Warn).Msgf("接收到系统停止信号")
}

// 加载配置文件
func loadConfig(path string) map[string]interface{} {
	if "" == path {
		path = "conf.d"
	}

	if fi, err := os.Stat(path); nil != err || !fi.IsDir() {
		withTag(log.Panic).Err(err).Msgf("配置路径必须是目录，当前是: %s", path)
		return nil
	}

	mergedTxt := new(bytes.Buffer)
	withTag(log.Info).Msgf("加载配置目录: %s", path)
	if files, err := ioutil.ReadDir(path); nil != err {
		withTag(log.Panic).Err(err).Msgf("无法列举目录文件: %s", path)
	} else {
		if 0 == len(files) {
			withTag(log.Panic).Err(err).Msgf("配置目录中没有任何文件: %s", path)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".toml") {
				continue
			}
			path := fmt.Sprintf("%s%s%s", path, "/", f.Name())
			withTag(log.Info).Msgf("加载TOML配置文件: %s", path)
			if bs, err := ioutil.ReadFile(path); nil != err {
				withTag(log.Panic).Err(err).Msgf("加载配置文件出错: %s", path)
			} else {
				mergedTxt.Write(bs)
			}
		}
	}

	if tree, err := toml.LoadBytes(mergedTxt.Bytes()); nil != err {
		withTag(log.Panic).Err(err).Msg("TOML配置文件无法解析")
		return nil
	} else {
		return tree.ToMap()
	}
}

func withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Main")
}
