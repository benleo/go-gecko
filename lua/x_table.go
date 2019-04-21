package lua

import (
	"github.com/yuin/gopher-lua"
	"time"
)

// mapToLTable converts a Go map to a lua table
func mapToLTable(m map[string]interface{}) *lua.LTable {
	// Main table pointer
	out := &lua.LTable{}

	// Loop map
	for key, val := range m {

		switch val.(type) {
		case float64:
			out.RawSetString(key, lua.LNumber(val.(float64)))
		case int64:
			out.RawSetString(key, lua.LNumber(val.(int64)))
		case string:
			out.RawSetString(key, lua.LString(val.(string)))
		case bool:
			out.RawSetString(key, lua.LBool(val.(bool)))
		case []byte:
			out.RawSetString(key, lua.LString(string(val.([]byte))))
		case map[string]interface{}:
			// Get table from map
			t := mapToLTable(val.(map[string]interface{}))
			out.RawSetString(key, t)

		case time.Time:
			out.RawSetString(key, lua.LNumber(val.(time.Time).Unix()))

		case []map[string]interface{}:
			// Create slice table
			array := &lua.LTable{}
			// Loop val
			for _, s := range val.([]map[string]interface{}) {
				// Get table from map
				t := mapToLTable(s)
				array.Append(t)
			}
			// Set slice table
			out.RawSetString(key, array)

		case []interface{}:
			// Create slice table
			array := &lua.LTable{}
			// Loop interface slice
			for _, s := range val.([]interface{}) {
				// Switch interface type
				switch s.(type) {
				case map[string]interface{}:
					// Convert map to table
					t := mapToLTable(s.(map[string]interface{}))
					// Append result
					array.Append(t)

				case float64:
					// Append result as number
					array.Append(lua.LNumber(s.(float64)))

				case string:
					// Append result as string
					array.Append(lua.LString(s.(string)))

				case bool:
					// Append result as bool
					array.Append(lua.LBool(s.(bool)))
				}
			}

			// Append to main table
			out.RawSetString(key, array)
		}
	}

	return out
}
