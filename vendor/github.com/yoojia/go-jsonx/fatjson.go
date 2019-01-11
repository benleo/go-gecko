package jsonx

import (
	"bytes"
	"fmt"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 扁平JSON构建器
//

type FatJSON struct {
	writer    *bytes.Buffer
	hasFields bool
	isBuild   bool
}

func NewFatJSON() *FatJSON {
	slf := &FatJSON{
		writer:    new(bytes.Buffer),
		hasFields: false,
		isBuild:   false,
	}
	slf.writer.WriteByte('{')
	return slf
}

func (slf *FatJSON) Bytes() []byte {
	return slf.build().writer.Bytes()
}

func (slf *FatJSON) String() string {
	return slf.build().writer.String()
}

func (slf *FatJSON) FieldNotEscapeValue(key string, value interface{}) *FatJSON {
	return slf.addField(key, value, false)
}

func (slf *FatJSON) Field(key string, value interface{}) *FatJSON {
	return slf.addField(key, value, true)
}

func (slf *FatJSON) addField(key string, value interface{}, escapeValue bool) *FatJSON {
	if slf.hasFields {
		slf.writer.WriteByte(',')
	}
	slf.hasFields = true

	slf.writer.WriteByte('"')
	slf.writer.WriteString(key)
	slf.writer.WriteByte('"')
	slf.writer.WriteByte(':')

	if escapeValue {
		if _, ok := value.(string); ok {
			slf.writer.WriteByte('"')
			slf.writer.WriteString(fmt.Sprintf("%v", value))
			slf.writer.WriteByte('"')
			return slf
		}
	}

	slf.writer.WriteString(fmt.Sprintf("%v", value))

	return slf
}

func (slf *FatJSON) build() *FatJSON {
	if !slf.isBuild {
		slf.isBuild = true
		slf.writer.WriteByte('}')
	}
	return slf
}
