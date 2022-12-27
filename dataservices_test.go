package dataservicex_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/vnteamopen/dataservicex"
)

type Person struct {
	ID        int64          `db:"person_id" goqu:"skipupdate"`
	Name      string         `db:"name"`
	Age       int64          `db:"age"`
	Height    float64        `db:"height"`
	CreatedAt mysql.NullTime `db:"created_at"`
}

func (Person) TableName() string {
	return "person"
}

func (Person) IDColumnName() string {
	return "person_id"
}

func TestCRUD(t *testing.T) {
	address, port, dbName := fakeDB()
	db := setup(address, port, dbName)
	dialect := goqu.Dialect("mysql")
	dataService := dataservicex.NewDataServices(db, dataservicex.WithDialect[Person](dialect))

	// CREATE DB
	createdAt := time.Now()
	createdPerson, err := dataService.Create(context.Background(), Person{
		Name:      "Thuc Le",
		Age:       25,
		Height:    165.5,
		CreatedAt: mysql.NullTime{Valid: true, Time: createdAt},
	})
	assert.NoError(t, err, "Create person")
	assert.NotEqual(t, 0, createdPerson.ID, "created person ID shouldn't be 0")
	assert.Equal(t, "Thuc Le", createdPerson.Name, "created person Name")
	assert.Equal(t, int64(25), createdPerson.Age, "created person Age")
	assert.Equal(t, 165.5, createdPerson.Height, "created person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), createdPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")

	// UPDATE
	updatingPerson := createdPerson
	updatingPerson.Name = "ledongthuc"
	updatingPerson.Height = 170.1
	updatedPerson, err := dataService.Update(context.Background(), updatingPerson.ID, updatingPerson)
	assert.NoError(t, err)
	assert.Equal(t, updatingPerson.ID, updatedPerson.ID, "updated person ID")
	assert.Equal(t, "ledongthuc", updatedPerson.Name, "updated person Name")
	assert.Equal(t, int64(25), createdPerson.Age, "updated person Age")
	assert.Equal(t, 170.1, updatedPerson.Height, "updated person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), createdPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")

	// UPDATE columns
	updatingPerson = updatedPerson
	updatingPerson.Name = "Thuc Le"
	updatingPerson.Height = 16
	err = dataService.UpdateColumns(context.Background(), updatingPerson.ID, goqu.Record{
		"name":   "Thuc Le",
		"height": 165.5,
	})
	assert.NoError(t, err)

	// GET after update
	updatedPerson, err = dataService.GetByID(context.Background(), updatingPerson.ID)
	assert.NoError(t, err)
	assert.Equal(t, updatingPerson.ID, updatedPerson.ID, "updated columns person ID")
	assert.Equal(t, "Thuc Le", updatedPerson.Name, "updated columns person Name")
	assert.Equal(t, int64(25), createdPerson.Age, "updated columns person Age")
	assert.Equal(t, 165.5, updatedPerson.Height, "updated columns person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), createdPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")

	// CREATE 2nd records
	createdPerson2, err := dataService.Create(context.Background(), Person{
		Name:      "Thuc Le (2)",
		Age:       252,
		Height:    165.52,
		CreatedAt: mysql.NullTime{Valid: true, Time: createdAt},
	})
	assert.NoError(t, err, "Create person (2)")
	assert.NotEqual(t, 0, createdPerson2.ID, "created person (2) ID shouldn't be 0")
	assert.Equal(t, "Thuc Le (2)", createdPerson2.Name, "created person (2) Name")
	assert.Equal(t, int64(252), createdPerson2.Age, "created person (2) Age")
	assert.Equal(t, 165.52, createdPerson2.Height, "created person (2) Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), createdPerson2.CreatedAt.Time.UTC().Format(time.RFC3339), "created person (2) CreatedAt")

	// GET LIST
	personList, err := dataService.GetList(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(personList))
	assert.NotEqual(t, personList[0].ID, personList[1].ID)

	for index, person := range personList {
		switch person.ID {
		case updatedPerson.ID:
			assert.Equal(t, "Thuc Le", person.Name, "updated columns person Name")
			assert.Equal(t, int64(25), person.Age, "updated columns person Age")
			assert.Equal(t, 165.5, person.Height, "updated columns person Height")
			assert.Equal(t, createdAt.UTC().Format(time.RFC3339), person.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")
		case createdPerson2.ID:
			assert.Equal(t, "Thuc Le (2)", person.Name, "created person (2) Name")
			assert.Equal(t, int64(252), person.Age, "created person (2) Age")
			assert.Equal(t, 165.52, person.Height, "created person (2) Height")
			assert.Equal(t, createdAt.UTC().Format(time.RFC3339), person.CreatedAt.Time.UTC().Format(time.RFC3339), "created person (2) CreatedAt")
		default:
			t.Errorf("Unexpected person[%d] ID: %v. Expected in: %v, %v", index, person.ID, updatedPerson.ID, createdPerson2.ID)
		}
	}

	// DELETE
	assert.NoError(t, dataService.Delete(context.Background(), personList[0].ID))
	assert.NoError(t, dataService.Delete(context.Background(), personList[1].ID))
	personList, err = dataService.GetList(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(personList))
}

func fakeDB() (address string, port int, dbName string) {
	dbName = "testDB"
	engine := sqle.NewDefault(
		sql.NewDatabaseProvider(
			func(_ *sql.Context) *memory.Database {
				db := memory.NewDatabase(dbName)
				return db
			}(sql.NewEmptyContext()),
			information_schema.NewInformationSchemaDatabase(),
		))

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port = listener.Addr().(*net.TCPAddr).Port
	if err != nil {
		panic(err)
	}
	listener.Close()

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", address, port),
	}
	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		panic(err)
	}
	go func() {
		if err = s.Start(); err != nil {
			panic(err)
		}
	}()

	return "localhost", port, dbName
}

func setup(address string, port int, dbName string) *sqlx.DB {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("no_user:@tcp(%s:%d)/%s", address, port, dbName))
	if err != nil {
		panic(err)
	}

	var schema = `
CREATE TABLE person (
  person_id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
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

func TestGetDialect(t *testing.T) {
	address, port, dbName := fakeDB()
	db := setup(address, port, dbName)
	dialect := goqu.Dialect("mysql")
	dataService := dataservicex.NewDataServices(db, dataservicex.WithDialect[Person](dialect))
	assert.Equal(t, dialect, dataService.GetDialect())
}
