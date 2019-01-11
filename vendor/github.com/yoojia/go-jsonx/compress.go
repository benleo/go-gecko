package jsonx

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var ErrNotJSONData = errors.New("json.compress: not json data")

// 将多行JSON文本压缩成一行
func CompressJSONText(data string) []byte {
	out := bytes.NewBuffer(make([]byte, 0))
	CompressJSON(strings.NewReader(data), out)
	return out.Bytes()
}

// 将多行JSON文本压缩成一行
func CompressJSONBytes(data []byte) []byte {
	out := bytes.NewBuffer(make([]byte, 0))
	CompressJSON(bytes.NewReader(data), out)
	return out.Bytes()
}

// 压缩JSON数据。如果数据并非JSON起始标记，则原样输出
func CompressJSON(in io.Reader, out io.Writer) error {

	buff := make([]byte, 1)
	scopeJSON := false
	scopeValue := false
	size := 0

	for {
		if read, err := in.Read(buff); nil != err {
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		} else if read != 1 {
			return errors.New("json.compress: read size != 1")
		}

		// Skip space
		switch buff[0] {
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
			if !scopeValue {
				continue
			}

		case '"':
			scopeValue = scopeJSON && !scopeValue

		case '[', '{':
			scopeJSON = scopeJSON || size == 0
		}

		if w, err := out.Write(buff); nil != err {
			return err
		} else {
			size += w
		}
	}

	if size == 0 {
		return ErrNotJSONData
	} else {
		return nil
	}
}

func HasJSONMark(bytes []byte) bool {
	size := len(bytes)
	if size < len(`[0]`) {
		return false
	}
	idx := size - 1
	start := bytes[0]
	end := bytes[idx]
	if start == '{' && end == '}' {
		return true
	}
	if start == '[' && end == ']' {
		return true
	}
	return false
}
