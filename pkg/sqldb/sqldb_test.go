package sqldb

import (
	"context"
	"sniper/pkg/conf"
	"testing"
	"time"
)

var schema = `
CREATE TABLE users (
  id integer primary key,
  age integer,
  name varchar(30),
  created datetime default CURRENT_TIMESTAMP
)
`

type user struct {
	ID      int
	Name    string
	Age     int
	Created time.Time
}

func TestSqlDb(t *testing.T) {
	conf.Set("SQLDB_DSN_foo", ":memory:")
	ctx := context.Background()

	ctx, db := Get(ctx, "foo")
	db.MustExecContext(ctx, schema)

	result, err := db.ExecContext(ctx,
		"insert into users(name,age) values (?,?)", "a", 1)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := result.LastInsertId()
	row := db.QueryRowxContext(ctx, "select * from users where id = ?", id)
	var u1 user
	if err := row.StructScan(&u1); err != nil {
		t.Fatal(err)
	}

	if u1.ID != 1 || u1.Name != "a" || u1.Age != 1 || u1.Created.IsZero() {
		t.Fatal("invalid user", u1)
	}

	tx := db.MustBegin()
	tx.Exec("delete from users")
	tx.Rollback()

	stmt, err := db.PreparexContext(ctx, "select * from users where id = ?")
	if err != nil {
		t.Fatal(err)
	}

	row = stmt.QueryRowxContext(ctx, id)
	var u2 user
	if err := row.StructScan(&u2); err != nil {
		t.Fatal(err)
	}

	if u2.ID != 1 || u2.Name != "a" || u2.Age != 1 || u2.Created.IsZero() {
		t.Fatal("invalid user", u2)
	}
}
