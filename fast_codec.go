package gomysql

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	stringSliceMagic = "GMS1"
	timeMagic        = "GMT1"
	sqlTimeLayout    = "2006-01-02T15:04:05.000000000Z"
)

var timeType = reflect.TypeOf(time.Time{})

func appendUvarint(dst []byte, v uint64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	return append(dst, buf[:n]...)
}

func encodeStringSlice(values []string) []byte {
	if values == nil {
		buf := make([]byte, 0, len(stringSliceMagic)+1)
		buf = append(buf, stringSliceMagic...)
		buf = append(buf, 1)
		return buf
	}

	totalLen := 0
	for _, value := range values {
		totalLen += len(value) + binary.MaxVarintLen64
	}

	buf := make([]byte, 0, len(stringSliceMagic)+1+binary.MaxVarintLen64+totalLen)
	buf = append(buf, stringSliceMagic...)
	buf = append(buf, 0)
	buf = appendUvarint(buf, uint64(len(values)))
	for _, value := range values {
		buf = appendUvarint(buf, uint64(len(value)))
		buf = append(buf, value...)
	}

	return buf
}

func decodeStringSlice(raw []byte) ([]string, bool, error) {
	if len(raw) < len(stringSliceMagic)+1 || string(raw[:len(stringSliceMagic)]) != stringSliceMagic {
		return nil, false, nil
	}

	flags := raw[len(stringSliceMagic)]
	if flags&1 != 0 {
		return nil, true, nil
	}

	idx := len(stringSliceMagic) + 1
	count, n := binary.Uvarint(raw[idx:])
	if n <= 0 {
		return nil, true, fmt.Errorf("invalid string slice length")
	}
	idx += n

	if count == 0 {
		return []string{}, true, nil
	}

	result := make([]string, 0, int(count))
	for i := 0; i < int(count); i++ {
		if idx >= len(raw) {
			return nil, true, fmt.Errorf("invalid string slice data")
		}
		itemLen, n := binary.Uvarint(raw[idx:])
		if n <= 0 {
			return nil, true, fmt.Errorf("invalid string slice item length")
		}
		idx += n
		if idx+int(itemLen) > len(raw) {
			return nil, true, fmt.Errorf("invalid string slice item data")
		}
		result = append(result, string(raw[idx:idx+int(itemLen)]))
		idx += int(itemLen)
	}

	if idx != len(raw) {
		return nil, true, fmt.Errorf("invalid string slice trailing data")
	}

	return result, true, nil
}

func encodeTimeValue(value time.Time) ([]byte, error) {
	raw, err := value.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, len(timeMagic)+len(raw))
	buf = append(buf, timeMagic...)
	buf = append(buf, raw...)
	return buf, nil
}

func decodeTimeValue(raw []byte) (time.Time, bool, error) {
	if len(raw) < len(timeMagic) || string(raw[:len(timeMagic)]) != timeMagic {
		return time.Time{}, false, nil
	}
	var value time.Time
	if err := value.UnmarshalBinary(raw[len(timeMagic):]); err != nil {
		return time.Time{}, true, err
	}
	return value, true, nil
}

func formatSQLTimeValue(value time.Time) string {
	return value.UTC().Format(sqlTimeLayout)
}

func parseSQLTimeValue(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		sqlTimeLayout,
		time.RFC3339Nano,
		"2006-01-02 15:04:05.000000000Z07:00",
		"2006-01-02 15:04:05.000000000",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format %q", raw)
}
