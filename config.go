package gecko

import (
	"bytes"
	"fmt"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"strings"
)

// 加载配置文件
func LoadConfig(path string) map[string]interface{} {
	if "" == path {
		path = "conf.d"
	}

	if fi, err := os.Stat(path); nil != err || !fi.IsDir() {
		_configTag(log.Panic).Err(err).Msgf("配置路径必须是目录: %s", path)
		return nil
	}

	mergedTxt := new(bytes.Buffer)
	_configTag(log.Info).Msgf("加载配置目录: %s", path)
	if files, err := ioutil.ReadDir(path); nil != err {
		_configTag(log.Panic).Err(err).Msgf("无法列举目录文件: %s", path)
	} else {
		if 0 == len(files) {
			_configTag(log.Panic).Err(err).Msgf("配置目录中没有任何文件: %s", path)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".toml") {
				continue
			}
			path := fmt.Sprintf("%s%s%s", path, "/", f.Name())
			_configTag(log.Info).Msgf("加载TOML配置文件: %s", path)
			if bs, err := ioutil.ReadFile(path); nil != err {
				_configTag(log.Panic).Err(err).Msgf("加载配置文件出错: %s", path)
			} else {
				mergedTxt.Write(bs)
			}
		}
	}

	if tree, err := toml.LoadBytes(mergedTxt.Bytes()); nil != err {
		_configTag(log.Panic).Err(err).Msg("TOML配置文件无法解析")
		return nil
	} else {
		return tree.ToMap()
	}
}

func _configTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Config")
}
