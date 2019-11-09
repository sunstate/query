# Query

This is a small library that mainstreams the way a database gets queried.

## Usage example
#### Query a row
```go
package main

import (
	"database/sql"
	"github.com/sunstate/query"
	"log"
)

func main() {
	type product struct {
		id   int64
		name string
	}

    // Initiate database properly
	var db *sql.DB
	q := query.NewQ(db)

	var p product
	err := q.Query(`SELECT id, name FROM products WHERE id = $1`).Arguments(1).RunRow(&p.id, &p.name)
	
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("unexpected error: %s", err)
	}
}
```

#### Query several rows
```go
package main

import (
	"database/sql"
	"github.com/sunstate/query"
	"log"
)

func main() {
	type product struct {
		id   int64
		name string
	}

    // Initiate database properly
	var db *sql.DB
	q := query.NewQ(db)

	var result []product
	err := q.Query(`SELECT id, name FROM products`).RunWithFunc(func(rows *sql.Rows) {
		for rows.Next() {
			var p product
			err := rows.Scan(&p.id, &p.name)
			if err != nil {
				log.Fatalf("unexpected error: %s", err)
			}
			
			result = append(result, p)
		}
	})
	
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("unexpected error: %s", err)
	}
}
```

#### Query with insert
```go
package main

import (
	"database/sql"
	"github.com/sunstate/query"
	"log"
)

func main() {
	type product struct {
		id   int64
		name string
	}

    // Initiate database properly
	var db *sql.DB
	q := query.NewQ(db)

	p := product{
		name: "shoes",
	}
	
	err := q.Query(`INSERT INTO products(name) VALUES($1) RETURNING id`).Arguments(&p.name).InsertReturningId(&p.id)

	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("unexpected error: %s", err)
	}
}
```

#### Query as a database transaction
```go
package main

import (
	"database/sql"
	"github.com/sunstate/query"
	"log"
)

func main() {
	type product struct {
		id   int64
		name string
	}

	// Initiate database properly
	var db *sql.DB
	tx, err := query.NewTx(db)
	if err != nil {
		log.Fatalf("could not begin transaction %s", err)
	}
	
	var result []product
	err = tx.Query(`SELECT id, name FROM products`).RunWithFunc(func(rows *sql.Rows) {
		for rows.Next() {
			var p product
			err := rows.Scan(&p.id, &p.name)
			if err != nil {
				log.Fatalf("unexpected error: %s", err)
			}

			result = append(result, p)
		}		
	})
	
	if err != nil {
		_ = query.Rollback(&tx)
		log.Fatalf("unexpected error: %s", err)
	}
	
	p := product{
		name: "boots",
	}
	
	err = tx.Query(`INSERT INTO products(name) VALUES($1) RETURNING id`).Arguments(&p.name).InsertReturningId(&p.id)

	if err != nil {
		_ = query.Rollback(&tx)
		log.Fatalf("unexpected error: %s", err)
	}
	
	err = query.Commit(&tx)
	if err != nil {
		log.Fatalf("could not commit transaction: %s", err)
	}
}
```