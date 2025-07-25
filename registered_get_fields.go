package gomysql

func (r *RegisteredStruct[T]) FieldBySQLName(sqlName string) *RegisteredStructField {
	for _, p := range r.Fields {
		if p.Opts.KeyName == sqlName {
			return &p
		}
	}

	return nil
}

func (r *RegisteredStruct[T]) FieldByGoName(goName string) *RegisteredStructField {
	for _, p := range r.Fields {
		if p.RealName == goName {
			return &p
		}
	}

	return nil
}
