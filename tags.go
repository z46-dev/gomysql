package gomysql

import (
	"fmt"
	"strings"
)

type SQLTagOpts struct {
	KeyName    string
	PrimaryKey bool
	AutoIncr   bool
	Unique     bool
	NotNull    bool
}

func mustParseTag(tag string) (output SQLTagOpts) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		panic(fmt.Sprintf("invalid tag format: %s", tag))
	}

	if len(parts) == 1 {
		output.KeyName = parts[0]
	} else {
		output.KeyName = parts[0]
		for _, part := range parts[1:] {
			switch part {
			case "primary":
				output.PrimaryKey = true
			case "increment":
				output.AutoIncr = true
			case "unique":
				output.Unique = true
			case "notnull":
				output.NotNull = true
			default:
				panic(fmt.Sprintf("unknown tag option: %s", part))
			}
		}
	}

	return
}
