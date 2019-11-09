package query

import (
	"context"
	"database/sql"
	"errors"
)

type Method string

const (
	TX Method = "TX"
	Q  Method = "Q"
)

type Item struct {
	q      string
	args   []interface{}
	method Method
	used   bool
	tx     *sql.Tx
	db     *sql.DB
}

// Initiate a database transaction
func NewTx(db *sql.DB) (Item, error) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	return Item{tx: tx, method: TX}, err
}

// Initiate a database query
func NewQ(db *sql.DB) Item {
	return Item{db: db, method: Q}
}

func (i *Item) Query(str string) *Item {
	i.q = str
	return i
}

func (i *Item) Arguments(args ...interface{}) *Item {
	i.args = args
	return i
}

func (i *Item) Run() error {
	var err error

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		_, err = i.tx.Exec(i.q, i.args...)
	}

	if i.method == Q {
		_, err = i.db.Exec(i.q, i.args...)
	}

	return err
}

func (i *Item) RunWithFunc(op func(*sql.Rows)) error {
	var (
		rows *sql.Rows
		err  error
	)

	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		rows, err = i.tx.Query(i.q, i.args...)
	}

	if i.method == Q {
		rows, err = i.db.Query(i.q, i.args...)
	}

	if err != nil {
		return err
	}

	op(rows)

	return rows.Close()
}

func (i *Item) RunRow(dest ...interface{}) error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		return i.tx.QueryRow(i.q, i.args...).Scan(dest...)
	}

	if i.method == Q {
		return i.db.QueryRow(i.q, i.args...).Scan(dest...)
	}

	return errors.New("unknown method")
}

func (i *Item) Insert() error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		_, err := i.tx.Exec(i.q, i.args...)
		return err
	}

	if i.method == Q {
		_, err := i.db.Exec(i.q, i.args...)
		return err
	}

	return errors.New("unknown method")
}

func (i *Item) InsertReturningId(id interface{}) error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		err := i.tx.QueryRow(i.q, i.args...).Scan(id)
		return err
	}

	if i.method == Q {
		err := i.db.QueryRow(i.q, i.args...).Scan(id)
		return err
	}

	return errors.New("unknown method")
}

func Commit(i *Item) error {
	if i.method != TX {
		return errors.New("could not commit because instance is not a transaction, it is a query")
	}

	if i.used {
		return errors.New("transaction is already committed, you have to start a new transaction")
	}

	if i.tx != nil {
		if err := i.tx.Commit(); err != nil {
			_ = i.tx.Rollback()
			return err
		}
	}

	return nil
}

func Rollback(i *Item) error {
	return i.tx.Rollback()
}
