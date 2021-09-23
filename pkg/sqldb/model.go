package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
)

// Modeler 接口提供查询模型的表结构信息
// 所有模型都需要实现本接口
type Modeler interface {
	// TableName 返回表名
	TableName() string
	// TableName 返回主键字段名
	KeyName() string
}

// 统一 DB 和 Tx 对象
type mapExecer interface {
	DriverName() string
	GetMapper() *reflectx.Mapper
	Rebind(string) string
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// MustBegin 封装 sqlx.DB.MustBegin，返回自定义的 *Tx
func (db *DB) MustBegin() *Tx {
	tx := db.DB.MustBegin()
	return &Tx{tx}
}

// Beginx 封装 sqlx.DB.Beginx，返回自定义的 *Tx
func (db *DB) Beginx() (*Tx, error) {
	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

// BeginTxx 封装 sqlx.DB.BeginTxx，返回自定义的 *Tx
func (db *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

// InsertContext 生成并执行 insert 语句
func (db *DB) InsertContext(ctx context.Context, m Modeler) (sql.Result, error) {
	return insert(ctx, db, m)
}

func (db *DB) Insert(m Modeler) (sql.Result, error) {
	return db.InsertContext(context.Background(), m)
}

// UpdateContext 生成并执行 update 语句
func (db *DB) UpdateContext(ctx context.Context, m Modeler) (sql.Result, error) {
	return update(ctx, db, m)
}

func (db *DB) Update(m Modeler) (sql.Result, error) {
	return db.UpdateContext(context.Background(), m)
}

// InsertContext 生成并执行 insert 语句
func (tx *Tx) InsertContext(ctx context.Context, m Modeler) (sql.Result, error) {
	return insert(ctx, tx, m)
}

func (tx *Tx) Insert(m Modeler) (sql.Result, error) {
	return tx.InsertContext(context.Background(), m)
}

// UpdateContext 生成并执行 update 语句
func (tx *Tx) UpdateContext(ctx context.Context, m Modeler) (sql.Result, error) {
	return update(ctx, tx, m)
}

func (tx *Tx) Update(m Modeler) (sql.Result, error) {
	return tx.UpdateContext(context.Background(), m)
}

// 添加 GetMapper 方法，方便与 Tx 统一
func (db *DB) GetMapper() *reflectx.Mapper {
	return db.Mapper
}

// 添加 GetMapper 方法，方便与 DB 统一
func (tx *Tx) GetMapper() *reflectx.Mapper {
	return tx.Mapper
}

func insert(ctx context.Context, db mapExecer, m Modeler) (sql.Result, error) {
	names, args, err := bindModeler(m, db.GetMapper())
	if err != nil {
		return nil, err
	}

	marks := ""
	var k int
	for i := 0; i < len(names); i++ {
		if names[i] == m.KeyName() {
			args = append(args[:i], args[i+1:]...)
			k = i
			continue
		}
		marks += "?,"
	}
	names = append(names[:k], names[k+1:]...)
	marks = marks[:len(marks)-1]
	query := "INSERT INTO " + m.TableName() + "(" + strings.Join(names, ",") + ") VALUES (" + marks + ")"
	query = db.Rebind(query)
	return db.ExecContext(ctx, query, args...)
}

func update(ctx context.Context, db mapExecer, m Modeler) (sql.Result, error) {
	names, args, err := bindModeler(m, db.GetMapper())
	if err != nil {
		return nil, err
	}

	query := "UPDATE " + m.TableName() + " set "
	var id interface{}
	for i := 0; i < len(names); i++ {
		name := names[i]
		if name == m.KeyName() {
			id = args[i]
			args = append(args[:i], args[i+1:]...)
			continue
		}
		query += name + "=?,"
	}
	query = query[:len(query)-1] + " WHERE " + m.KeyName() + " = ?"
	query = db.Rebind(query)
	args = append(args, id)
	return db.ExecContext(ctx, query, args...)
}

func bindModeler(arg interface{}, m *reflectx.Mapper) ([]string, []interface{}, error) {
	t := reflect.TypeOf(arg)
	names := []string{}
	for k := range m.TypeMap(t).Names {
		names = append(names, k)
	}
	args, err := bindArgs(names, arg, m)
	if err != nil {
		return nil, nil, err
	}

	return names, args, nil
}

func bindArgs(names []string, arg interface{}, m *reflectx.Mapper) ([]interface{}, error) {
	arglist := make([]interface{}, 0, len(names))

	// grab the indirected value of arg
	v := reflect.ValueOf(arg)
	for v = reflect.ValueOf(arg); v.Kind() == reflect.Ptr; {
		v = v.Elem()
	}

	err := m.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("could not find name %s in %#v", names[i], arg)
		}

		val := reflectx.FieldByIndexesReadOnly(v, t)
		arglist = append(arglist, val.Interface())

		return nil
	})

	return arglist, err
}
