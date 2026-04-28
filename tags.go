package gomysql

import (
	"fmt"
	"strings"
)

type ForeignKeyRef struct {
	TableName  string
	ColumnName string
	OnDelete   ForeignKeyAction
}

type ForeignKeyAction string

const (
	ForeignKeyActionCascade ForeignKeyAction = "CASCADE"
)

type SQLTagOpts struct {
	KeyName    string
	PrimaryKey bool
	AutoIncr   bool
	Unique     bool
	NotNull    bool
	ForeignKey *ForeignKeyRef
}

func mustParseTag(tag string) (output SQLTagOpts) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		panic(fmt.Sprintf("invalid tag format: %s", tag))
	}

	var onDelete ForeignKeyAction

	if len(parts) == 1 {
		output.KeyName = strings.TrimSpace(parts[0])
	} else {
		output.KeyName = strings.TrimSpace(parts[0])
		for _, part := range parts[1:] {
			part = strings.TrimSpace(part)
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
				if strings.HasPrefix(part, "fkey:") {
					ref := strings.TrimPrefix(part, "fkey:")
					target := strings.Split(ref, ".")
					if len(target) != 2 || strings.TrimSpace(target[0]) == "" || strings.TrimSpace(target[1]) == "" {
						panic(fmt.Sprintf("invalid foreign key option: %s", part))
					}

					output.ForeignKey = &ForeignKeyRef{
						TableName:  strings.TrimSpace(target[0]),
						ColumnName: strings.TrimSpace(target[1]),
						OnDelete:   onDelete,
					}
					continue
				}

				if strings.HasPrefix(part, "ondelete:") {
					action := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(part, "ondelete:")))
					switch action {
					case "cascade":
						onDelete = ForeignKeyActionCascade
					default:
						panic(fmt.Sprintf("invalid foreign key delete action: %s", part))
					}

					if output.ForeignKey != nil {
						output.ForeignKey.OnDelete = onDelete
					}
					continue
				}

				panic(fmt.Sprintf("unknown tag option: %s", part))
			}
		}
	}

	if output.KeyName == "" {
		panic(fmt.Sprintf("invalid tag format: %s", tag))
	}

	if onDelete != "" && output.ForeignKey == nil {
		panic(fmt.Sprintf("foreign key delete action requires fkey: %s", tag))
	}

	return
}
