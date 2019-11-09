package query_test

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sunstate/query"
	"testing"
)

func TestNewQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	mock.ExpectExec("UPDATE products").WithArgs("shoes", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM products").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
	mock.ExpectQuery("SELECT name FROM products").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("shoes"))

	q := query.NewQ(db)
	err = q.Query(`
		UPDATE
			products
		SET
			name = $1
		WHERE
			id = $2
	`).Arguments("shoes", 1).Run()

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	var result []string
	q = query.NewQ(db)
	err = q.Query(`
		SELECT
			id
		FROM
			products
	`).RunWithFunc(func(rows *sql.Rows) {
		for rows.Next() {
			var id string
			err := rows.Scan(&id)
			if err != nil {
				t.Errorf("scaning rows should not return error: %s", err)
			}

			result = append(result, id)
		}
	})

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 result values and not %d values", len(result))
	}

	var name string
	q = query.NewQ(db)
	err = q.Query(`
		SELECT
			name
		FROM
			products
		WHERE
			id = $1
	`).Arguments(1).RunRow(&name)

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	if name != "shoes" {
		t.Errorf("expected result string to be shoes and not %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestNewTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE products").WithArgs("shoes", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT id FROM products").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2).AddRow(3))
	mock.ExpectQuery("SELECT name FROM products").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("shoes"))
	mock.ExpectExec("INSERT INTO products").WithArgs(1, "boots").WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectCommit()
	mock.ExpectClose()

	tx, err := query.NewTx(db)
	if err != nil {
		t.Fatalf("error was not expected when opening a new database transaction: %s", err)
	}

	err = tx.Query(`
		UPDATE
			products
		SET
			name = $1
		WHERE
			id = $2
	`).Arguments("shoes", 1).Run()

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	var result []string
	err = tx.Query(`
		SELECT
			id
		FROM
			products
	`).RunWithFunc(func(rows *sql.Rows) {
		for rows.Next() {
			var id string
			err := rows.Scan(&id)
			if err != nil {
				t.Errorf("scaning rows should not return error: %s", err)
			}

			result = append(result, id)
		}
	})

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 result values and not %d values", len(result))
	}

	var name string
	err = tx.Query(`
		SELECT
			name
		FROM
			products
		WHERE
			id = $1
	`).Arguments(1).RunRow(&name)

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	if name != "shoes" {
		t.Errorf("expected result string to be shoes and not %s", err)
	}

	err = tx.Query(`
		INSERT INTO
			products(id, name)
		VALUES($1, $2)
	`).Arguments(1, "boots").Insert()

	if err != nil {
		t.Errorf("error was not expected while querying database: %s", err)
	}

	err = query.Commit(&tx)
	if err != nil {
		t.Fatalf("error was not expected when doing a database commit: %s", err)
	}

	err = db.Close()


	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}