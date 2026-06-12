package gosqlite

func (r *RegisteredStruct[T]) runCreation() (err error) {
	if r.driver == nil {
		err = ErrDatabaseNotInitialized
		return
	}

	r.driver.lock.Lock()
	defer r.driver.lock.Unlock()

	if _, err = r.driver.db.Exec(r.createTableSQL); err != nil {
		return
	}

	return
}
