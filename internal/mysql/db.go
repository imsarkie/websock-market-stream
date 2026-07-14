package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql" //Run this package's init() function.
)

type Store struct{
	db *sql.DB
}

func NewStore(dsn string) (*Store, error){
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}