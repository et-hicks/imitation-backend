package sqlitedriver

/*
#cgo LDFLAGS: -lsqlite3
#include <sqlite3.h>
#include <stdlib.h>

static inline sqlite3_destructor_type goSqliteTransient() {
        return SQLITE_TRANSIENT;
}
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"unsafe"
)

type Driver struct{}

var registerOnce sync.Once

// Register registers the driver with database/sql.
func Register() {
	registerOnce.Do(func() {
		sql.Register("custom_sqlite", &Driver{})
	})
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	cpath := C.CString(name)
	defer C.free(unsafe.Pointer(cpath))
	var db *C.sqlite3
	flags := C.int(C.SQLITE_OPEN_CREATE | C.SQLITE_OPEN_READWRITE | C.SQLITE_OPEN_URI)
	if rc := C.sqlite3_open_v2(cpath, &db, flags, nil); rc != C.SQLITE_OK {
		err := errorFromCode(rc, db)
		if db != nil {
			C.sqlite3_close(db)
		}
		return nil, err
	}
	C.sqlite3_busy_timeout(db, 5000)
	if err := execSimple(db, "PRAGMA foreign_keys = ON;"); err != nil {
		C.sqlite3_close(db)
		return nil, err
	}
	return &conn{db: db}, nil
}

type conn struct {
	db *C.sqlite3
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return nil, driver.ErrSkip
}

func (c *conn) Close() error {
	if c.db == nil {
		return nil
	}
	if rc := C.sqlite3_close(c.db); rc != C.SQLITE_OK {
		return errorFromCode(rc, c.db)
	}
	c.db = nil
	return nil
}

func (c *conn) Begin() (driver.Tx, error) {
	if err := execSimple(c.db, "BEGIN"); err != nil {
		return nil, err
	}
	return &tx{c: c}, nil
}

type tx struct {
	c *conn
}

func (t *tx) Commit() error {
	return execSimple(t.c.db, "COMMIT")
}

func (t *tx) Rollback() error {
	return execSimple(t.c.db, "ROLLBACK")
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	defer C.sqlite3_finalize(stmt)
	if err := bindArgs(stmt, args); err != nil {
		return nil, err
	}
	for {
		rc := C.sqlite3_step(stmt)
		switch rc {
		case C.SQLITE_DONE:
			lastID := int64(C.sqlite3_last_insert_rowid(c.db))
			changes := int64(C.sqlite3_changes(c.db))
			return result{lastID: lastID, rowsAffected: changes}, nil
		case C.SQLITE_ROW:
			continue
		case C.SQLITE_BUSY:
			continue
		default:
			return nil, errorFromCode(rc, c.db)
		}
	}
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	stmt, err := c.prepare(query)
	if err != nil {
		return nil, err
	}
	if err := bindArgs(stmt, args); err != nil {
		C.sqlite3_finalize(stmt)
		return nil, err
	}
	cols := int(C.sqlite3_column_count(stmt))
	colNames := make([]string, cols)
	for i := 0; i < cols; i++ {
		name := C.sqlite3_column_name(stmt, C.int(i))
		colNames[i] = C.GoString(name)
	}
	return &rows{stmt: stmt, conn: c, cols: colNames}, nil
}

func (c *conn) prepare(query string) (*C.sqlite3_stmt, error) {
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))
	var stmt *C.sqlite3_stmt
	if rc := C.sqlite3_prepare_v2(c.db, cquery, -1, &stmt, nil); rc != C.SQLITE_OK {
		return nil, errorFromCode(rc, c.db)
	}
	return stmt, nil
}

type rows struct {
	stmt   *C.sqlite3_stmt
	conn   *conn
	cols   []string
	closed bool
}

func (r *rows) Columns() []string {
	return r.cols
}

func (r *rows) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	if rc := C.sqlite3_finalize(r.stmt); rc != C.SQLITE_OK {
		return errorFromCode(rc, r.conn.db)
	}
	return nil
}

func (r *rows) Next(dest []driver.Value) error {
	rc := C.sqlite3_step(r.stmt)
	switch rc {
	case C.SQLITE_ROW:
		for i := range dest {
			dest[i] = columnValue(r.stmt, C.int(i))
		}
		return nil
	case C.SQLITE_DONE:
		r.Close()
		return io.EOF
	case C.SQLITE_BUSY:
		return r.Next(dest)
	default:
		return errorFromCode(rc, r.conn.db)
	}
}

type result struct {
	lastID       int64
	rowsAffected int64
}

func (r result) LastInsertId() (int64, error) {
	return r.lastID, nil
}

func (r result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

func bindArgs(stmt *C.sqlite3_stmt, args []driver.NamedValue) error {
	for i, arg := range args {
		idx := C.int(arg.Ordinal)
		if idx <= 0 {
			idx = C.int(i + 1)
		}
		if err := bindValue(stmt, idx, arg.Value); err != nil {
			return err
		}
	}
	return nil
}

func bindValue(stmt *C.sqlite3_stmt, idx C.int, val any) error {
	switch v := val.(type) {
	case nil:
		C.sqlite3_bind_null(stmt, idx)
	case int:
		C.sqlite3_bind_int64(stmt, idx, C.sqlite3_int64(v))
	case int64:
		C.sqlite3_bind_int64(stmt, idx, C.sqlite3_int64(v))
	case int32:
		C.sqlite3_bind_int64(stmt, idx, C.sqlite3_int64(v))
	case float64:
		C.sqlite3_bind_double(stmt, idx, C.double(v))
	case float32:
		C.sqlite3_bind_double(stmt, idx, C.double(v))
	case bool:
		if v {
			C.sqlite3_bind_int(stmt, idx, 1)
		} else {
			C.sqlite3_bind_int(stmt, idx, 0)
		}
	case string:
		cstr := C.CString(v)
		defer C.free(unsafe.Pointer(cstr))
		if rc := C.sqlite3_bind_text(stmt, idx, cstr, C.int(len(v)), C.goSqliteTransient()); rc != C.SQLITE_OK {
			return errorFromCode(rc, nil)
		}
	case []byte:
		var ptr unsafe.Pointer
		if len(v) > 0 {
			ptr = C.CBytes(v)
			defer C.free(ptr)
		}
		if rc := C.sqlite3_bind_blob(stmt, idx, ptr, C.int(len(v)), C.goSqliteTransient()); rc != C.SQLITE_OK {
			return errorFromCode(rc, nil)
		}
	default:
		if valuer, ok := val.(driver.Valuer); ok {
			converted, err := valuer.Value()
			if err != nil {
				return err
			}
			return bindValue(stmt, idx, converted)
		}
		return fmt.Errorf("unsupported bind value type %T", val)
	}
	return nil
}

func columnValue(stmt *C.sqlite3_stmt, idx C.int) driver.Value {
	switch C.sqlite3_column_type(stmt, idx) {
	case C.SQLITE_INTEGER:
		return int64(C.sqlite3_column_int64(stmt, idx))
	case C.SQLITE_FLOAT:
		return float64(C.sqlite3_column_double(stmt, idx))
	case C.SQLITE_TEXT:
		text := C.sqlite3_column_text(stmt, idx)
		size := C.sqlite3_column_bytes(stmt, idx)
		return C.GoStringN((*C.char)(unsafe.Pointer(text)), size)
	case C.SQLITE_BLOB:
		size := C.sqlite3_column_bytes(stmt, idx)
		if size == 0 {
			return []byte{}
		}
		data := C.sqlite3_column_blob(stmt, idx)
		return C.GoBytes(data, size)
	case C.SQLITE_NULL:
		return nil
	default:
		return nil
	}
}

func errorFromCode(rc C.int, db *C.sqlite3) error {
	if rc == C.SQLITE_OK {
		return nil
	}
	if db != nil {
		msg := C.sqlite3_errmsg(db)
		return errors.New(C.GoString(msg))
	}
	return fmt.Errorf("sqlite error code %d", int(rc))
}

func execSimple(db *C.sqlite3, query string) error {
	cquery := C.CString(query)
	defer C.free(unsafe.Pointer(cquery))
	if rc := C.sqlite3_exec(db, cquery, nil, nil, nil); rc != C.SQLITE_OK {
		return errorFromCode(rc, db)
	}
	return nil
}

func init() {
	Register()
}
