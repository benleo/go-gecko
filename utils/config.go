package utils

import (
	"github.com/yoojia/go-value"
	"strings"
)

func ToMap(any interface{}) map[string]interface{} {
	switch any.(type) {
	case map[string]interface{}:
		return any.(map[string]interface{})

	case map[interface{}]interface{}:
		mm := any.(map[interface{}]interface{})
		out := make(map[string]interface{}, len(mm))
		for k, v := range mm {
			out[value.ToString(k)] = v
		}
		return out

	default:
		return make(map[string]interface{}, 0)
	}
}

func ToStringArray(any interface{}) []string {
	switch any.(type) {

	case []interface{}:
		array := any.([]interface{})
		out := make([]string, len(array))
		for i, v := range array {
			out[i] = value.ToString(v)
		}
		return out

	case []string:
		return any.([]string)

	case string:
		return strings.Split(any.(string), ",")

	default:
		return make([]string, 0)
	}
}
