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

	engine := new(gecko.GeckoEngine)
	engine.PrepareEnv()
	engine.Init(conf)
	engine.Start()
	defer engine.Stop()
	// 等待终止信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//
	<-sig
	withTag(log.Warn).Msgf("接收到系统停止信号")
}

// 加载配置文件
func loadConfig(dirpath string) map[string]interface{} {
	if "" == dirpath {
		dirpath = "conf.d"
	}

	if fi, err := os.Stat(dirpath); nil != err || !fi.IsDir() {
		withTag(log.Panic).Err(err).Msgf("Config path muse be a dir: %s", dirpath)
		return nil
	}

	mergedTxt := new(bytes.Buffer)
	withTag(log.Info).Msgf("Load config dir: %s", dirpath)
	if files, err := ioutil.ReadDir(dirpath); nil != err {
		withTag(log.Panic).Err(err).Msgf("Failed to list file in dir: %s", dirpath)
	} else {
		if 0 == len(files) {
			withTag(log.Panic).Err(err).Msgf("Config file NOT FOUND in dir: %s", dirpath)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".toml") {
				continue
			}
			path := fmt.Sprintf("%s%s%s", dirpath, "/", f.Name())
			withTag(log.Info).Msgf("Load config file: %s", path)
			if bs, err := ioutil.ReadFile(path); nil != err {
				withTag(log.Panic).Err(err).Msgf("Failed to load file: %s", path)
			} else {
				mergedTxt.Write(bs)
			}
		}
	}

	if tree, err := toml.LoadBytes(mergedTxt.Bytes()); nil != err {
		withTag(log.Panic).Err(err).Msg("Failed to decode toml config file")
		return nil
	} else {
		return tree.ToMap()
	}
}

func withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Main")
}
