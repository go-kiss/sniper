package sqldb

import (
	"context"
	"sniper/pkg/conf"
	"testing"
	"time"
)

var schema = `
CREATE TABLE IF NOT EXISTS users (
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

func (u *user) TableName() string { return "users" }
func (u *user) KeyName() string   { return "id" }

func TestSqlDb(t *testing.T) {
	conf.Set("SQLDB_DSN_foo", ":memory:")
	ctx := context.Background()

	db := Get(ctx, "foo")
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

func TestModel(t *testing.T) {
	conf.Set("SQLDB_DSN_foo", ":memory:")
	ctx := context.Background()

	db := Get(ctx, "foo")
	db.MustExecContext(ctx, schema)

	now := time.Now()
	u1 := &user{Name: "foo", Age: 18, Created: now}
	result, err := db.Insert(u1)
	if err != nil {
		t.Fatal(err)
	}

	id, _ := result.LastInsertId()

	u1.Name = "bar"
	u1.ID = int(id)

	_, err = db.Update(u1)
	if err != nil {
		t.Fatal(err)
	}

	var u2 user
	err = db.Get(&u2, "select * from users where id = ?", id)
	if err != nil {
		t.Fatal(err)
	}

	if u2.Name != "bar" || u2.Age != 18 || !u2.Created.Equal(now) {
		t.Fatal("invalid user", u2)
	}
}

func TestName(t *testing.T) {
	conf.Set("SQLDB_DSN_bar", ":memory:")
	conf.Set("SQLDB_DSN_baz", ":memory:")
	ctx := context.Background()

	db1 := Get(ctx, "bar")
	db1.MustExecContext(ctx, schema)
	db2 := Get(ctx, "baz")
	db2.MustExecContext(ctx, schema)

	now := time.Now()
	u1 := &user{Name: "foo", Age: 18, Created: now}

	result1, err := db1.Insert(u1)
	if err != nil {
		t.Fatal(err)
	}

	id1, _ := result1.LastInsertId()

	result2, err := db2.Insert(u1)
	if err != nil {
		t.Fatal(err)
	}

	id2, _ := result2.LastInsertId()

	if id1 != id2 {
		t.Fatal("invalid id", id1, id2)
	}
}
