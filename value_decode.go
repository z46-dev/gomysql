package gomysql

import (
	"fmt"
	"strconv"
	"strings"
)

// sqlBool normalizes various SQL bool representations (int, string, []byte) to a Go bool.
func sqlBool(raw any) (bool, error) {
	switch v := raw.(type) {
	case nil:
		return false, nil
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case int:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case []byte:
		return parseBoolString(string(v))
	case string:
		return parseBoolString(v)
	default:
		return false, fmt.Errorf("cannot convert %T to bool", raw)
	}
}

func parseBoolString(s string) (bool, error) {
	cleaned := strings.TrimSpace(strings.ToLower(s))
	switch cleaned {
	case "1", "true":
		return true, nil
	case "0", "false", "":
		return false, nil
	default:
		if b, err := strconv.ParseBool(cleaned); err == nil {
			return b, nil
		}
		return false, fmt.Errorf("cannot parse %q as bool", s)
	}
}
