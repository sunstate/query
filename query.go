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

// Initiate a database transaction with external context
func NewTxWithContext(db *sql.DB, ctx context.Context) (Item, error) {
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

func (i *Item) run(ctx context.Context) error {
	var err error

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		_, err = i.tx.ExecContext(ctx, i.q, i.args...)
	}

	if i.method == Q {
		_, err = i.db.ExecContext(ctx, i.q, i.args...)
	}

	return err
}

func (i *Item) Run() error {
	return i.run(context.Background())
}

func (i *Item) RunContext(ctx context.Context) error {
	return i.run(ctx)
}

func (i *Item) runWithFunc(ctx context.Context, op func(*sql.Rows)) error {
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

		rows, err = i.tx.QueryContext(ctx, i.q, i.args...)
	}

	if i.method == Q {
		rows, err = i.db.QueryContext(ctx, i.q, i.args...)
	}

	if err != nil {
		return err
	}

	op(rows)

	return rows.Close()
}


func (i *Item) RunWithFunc(op func(*sql.Rows)) error {
	return i.runWithFunc(context.Background(), op)
}

func (i *Item) RunWithFuncContext(ctx context.Context, op func(*sql.Rows)) error {
	return i.runWithFunc(ctx, op)
}

func (i *Item) runRow(ctx context.Context, dest ...interface{}) error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		return i.tx.QueryRowContext(ctx, i.q, i.args...).Scan(dest...)
	}

	if i.method == Q {
		return i.db.QueryRowContext(ctx, i.q, i.args...).Scan(dest...)
	}

	return errors.New("unknown method")
}

func (i *Item) RunRow(dest ...interface{}) error {
	return i.runRow(context.Background(), dest...)
}

func (i *Item) RunRowContext(ctx context.Context, dest ...interface{}) error {
	return i.runRow(ctx, dest...)
}

func (i *Item) insert(ctx context.Context) error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		_, err := i.tx.ExecContext(ctx, i.q, i.args...)
		return err
	}

	if i.method == Q {
		_, err := i.db.ExecContext(ctx, i.q, i.args...)
		return err
	}

	return errors.New("unknown method")
}

func (i *Item) Insert() error {
	return i.insert(context.Background())
}

func (i *Item) InsertContext(ctx context.Context) error {
	return i.insert(ctx)
}

func (i *Item) insertReturning(ctx context.Context, args ...interface{}) error {
	if i.q == "" {
		return errors.New("you need to define a query")
	}

	if i.method == TX {
		if i.used {
			return errors.New("transaction is already committed, you have to start a new transaction")
		}

		err := i.tx.QueryRowContext(ctx, i.q, i.args...).Scan(args...)
		return err
	}

	if i.method == Q {
		err := i.db.QueryRowContext(ctx, i.q, i.args...).Scan(args...)
		return err
	}

	return errors.New("unknown method")
}

func (i *Item) InsertReturning(args ...interface{}) error {
	return i.insertReturning(context.Background(), args...)
}

func (i *Item) InsertReturningContext(ctx context.Context, args ...interface{}) error {
	return i.insertReturning(ctx, args...)
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
