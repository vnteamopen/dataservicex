package tests

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

func BuildDataSource(address, database, username, password string, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, address, port, database)
}

func SetupDB(driverName, dataSource string) *sqlx.DB {
	db, err := sqlx.Connect(driverName, dataSource)
	if err != nil {
		panic(err)
	}

	schema := `
CREATE TABLE IF NOT EXISTS person (
	id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
	name TEXT,
	age INT,
	height REAL,
	created_at TIMESTAMP
);`

	db.MustExec(schema)
	if err = db.Ping(); err != nil {
		panic(err)
	}

	return db
}
