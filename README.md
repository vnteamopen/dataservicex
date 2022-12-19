
# dataservicex

`dataservicex` is a go library that wrapup common CRUD functions of data service layer to access to database. As common application, it frequently re-implement following functions with their data model:

 - Create model
 - Update full model
 - Update specific columns
 - Get a model by ID
 - Get list model
 - Delete by ID

The `dataservicex` uses generic, [goqu](https://github.com/doug-martin/goqu), and [sqlx](https://github.com/launchbadge/sqlx) to implement the common functions.

# Quickstart

## Install library

Import library to your project

```bash
go get -u github.com/vnteamopen/godebouncer
```

## Usage

```go
import (
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"


)

type Person struct {
	ID        int64          `db:"person_id" goqu:"skipupdate"`
	Name      string         `db:"name"`
	Age       int64          `db:"age"`
}

func (Person) TableName() string {
	return "person"
}

func (Person) IDColumnName() string {
	return "person_id"
}

func main() {
	db, _ := sqlx.Connect("mysql", connStr)
	dataService := dataservicex.NewDataServices[Person](db)
}
```

You can specific Dialect of database query generating

```go
	dialect := goqu.Dialect("mysql")
	dataService = dataservicex.NewDataServices(db, dataservicex.WithDialect[Person](dialect))
```

# Examples

## Create model

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	createdPerson, err := dataService.Create(context.Background(), Person{
		Name: "Thuc Le",
		Age: 25,
	})
}
```

## Update full model

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	updatingID := 1
	updatedPerson, err := dataService.Update(context.Background(), updatingID, Person{
		Name: "Thuc Le",
		Age: 25,
	})
}
```

## Update specific columns

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	updatingID := 1
	updatedPerson, err := dataService.UpdateColumns(context.Background(),
		updatingID,
		goqu.Record{
			"name": "Thuc Le",
		})
}
```

## Get a model by ID

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	person, err := dataService.GetByID(context.Background(), 1)
}
```

## Get list model

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	personList, err := dataService.GetList(context.Background())
}
```

## Delete by ID

```go
func main() {
	dataService := dataservicex.NewDataServices[Person](...)
	err := dataService.Delete(context.Background(), 1)
}
```

# License

MIT

# Contribution

All your contributions to project and make it better, they are welcome. Feel free to start an [issue](https://github.com/vnteamopen/dataservicex/issues).

# Thanks! 🙌

 - Viet Nam We Build group https://webuild.community for discussion.

[![Stargazers repo roster for @vnteamopen/dataservicex](https://reporoster.com/stars/vnteamopen/dataservicex)](https://github.com/vnteamopen/dataservicex/stargazers)
