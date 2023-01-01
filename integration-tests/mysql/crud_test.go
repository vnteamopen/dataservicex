package mysql

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/vnteamopen/dataservicex"
	"testing"
	"time"
)

type Person struct {
	ID        int64          `db:"id" goqu:"skipinsert,skipupdate"`
	Name      string         `db:"name"`
	Age       int64          `db:"age"`
	Height    float64        `db:"height"`
	CreatedAt mysql.NullTime `db:"created_at"`
}

func (Person) TableName() string {
	return "person"
}

func (Person) IDColumnName() string {
	return "id"
}

func buildMySQLDataSource(host, database, username, password string, port int) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, host, port, database)
}

func setupDB(dataSource string) *sqlx.DB {
	db, err := sqlx.Connect("mysql", dataSource)
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

func TestCRUD(t *testing.T) {
	dataSource := buildMySQLDataSource("localhost", "dataservicex", "dataservicex", "dataservicex", 3306)
	db := setupDB(dataSource)
	dialect := goqu.Dialect("mysql")
	dataService := dataservicex.NewDataServices(db, dataservicex.WithDialect[Person](dialect))

	// Create
	createdAt := time.Now()
	createdPerson, err := dataService.Create(context.Background(), Person{
		Name:      "Duy Nguyen",
		Age:       25,
		Height:    180.5,
		CreatedAt: mysql.NullTime{Valid: true, Time: createdAt},
	})
	assert.NoError(t, err, "create person")
	assert.NotEqual(t, 0, createdPerson, "created person ID shouldn't be 0")
	assert.Equal(t, "Duy Nguyen", createdPerson.Name, "created person Name")
	assert.Equal(t, int64(25), createdPerson.Age, "created person Age")
	assert.Equal(t, 180.5, createdPerson.Height, "created person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), createdPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")

	// Update
	updatingPerson := createdPerson
	updatingPerson.Name = "duyn"
	updatingPerson.Height = 180.2
	updatedPerson, err := dataService.Update(context.Background(), updatingPerson.ID, updatingPerson)
	assert.NoError(t, err, "update person")
	assert.Equal(t, updatingPerson.ID, updatedPerson.ID, "updated person ID")
	assert.Equal(t, "duyn", updatedPerson.Name, "updated person Name")
	assert.Equal(t, int64(25), updatedPerson.Age, "updated person Age")
	assert.Equal(t, 180.2, updatedPerson.Height, "updated person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), updatedPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "updated person CreatedAt")

	// Update columns
	updatingPerson = updatedPerson
	err = dataService.UpdateColumns(context.Background(), updatingPerson.ID, map[string]interface{}{
		"name":   "DuyN",
		"height": 180.1,
	})
	assert.NoError(t, err, "update columns in person table")

	// Get after update
	updatedPerson, err = dataService.GetByID(context.Background(), updatingPerson.ID)
	assert.NoError(t, err, "get person by id")
	assert.Equal(t, "DuyN", updatedPerson.Name, "updated person Name")
	assert.Equal(t, int64(25), updatedPerson.Age, "updated person Age")
	assert.Equal(t, 180.1, updatedPerson.Height, "updated person Height")
	assert.Equal(t, createdAt.UTC().Format(time.RFC3339), updatedPerson.CreatedAt.Time.UTC().Format(time.RFC3339), "updated person CreatedAt")

	// Create another person
	createdAt2 := time.Now()
	createdPerson2, err := dataService.Create(context.Background(), Person{
		Name:      "Duy Ng",
		Age:       26,
		Height:    180.5,
		CreatedAt: mysql.NullTime{Valid: true, Time: createdAt2},
	})
	assert.NoError(t, err, "create person")
	assert.NotEqual(t, 0, createdPerson2, "created person ID shouldn't be 0")
	assert.Equal(t, "Duy Ng", createdPerson2.Name, "created person Name")
	assert.Equal(t, int64(26), createdPerson2.Age, "created person Age")
	assert.Equal(t, 180.5, createdPerson2.Height, "created person Height")
	assert.Equal(t, createdAt2.UTC().Format(time.RFC3339), createdPerson2.CreatedAt.Time.UTC().Format(time.RFC3339), "created person CreatedAt")

	// Get list people
	listPerson, err := dataService.GetList(context.Background())
	assert.NoError(t, err, "get list people")
	assert.Equal(t, 2, len(listPerson), "num of people")
	assert.NotEqual(t, listPerson[0].ID, listPerson[1].ID, "two people shouldn't be same id")

	for _, person := range listPerson {
		switch person.ID {
		case createdPerson.ID:
			assert.NotEqual(t, 0, person, "person ID shouldn't be 0")
			assert.Equal(t, "DuyN", person.Name, "person Name")
			assert.Equal(t, int64(25), person.Age, "person Age")
			assert.Equal(t, 180.1, person.Height, "person Height")
			assert.Equal(t, createdAt.UTC().Format(time.RFC3339), person.CreatedAt.Time.UTC().Format(time.RFC3339), "person CreatedAt")
		case createdPerson2.ID:
			assert.NotEqual(t, 0, person, "person ID shouldn't be 0")
			assert.Equal(t, "Duy Ng", person.Name, "person Name")
			assert.Equal(t, int64(26), person.Age, "person Age")
			assert.Equal(t, 180.5, person.Height, "person Height")
			assert.Equal(t, createdAt2.UTC().Format(time.RFC3339), person.CreatedAt.Time.UTC().Format(time.RFC3339), "person CreatedAt")
		}
	}

	// Delete
	assert.NoError(t, dataService.Delete(context.Background(), createdPerson.ID), "delete person")
	assert.NoError(t, dataService.Delete(context.Background(), createdPerson2.ID), "delete person")
}
