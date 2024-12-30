// Copyright (c) 2023 Tim van der Molen <tim@kariliq.nl>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package sqlcipher

// #cgo CPPFLAGS: -DSQLCIPHER_CRYPTO_GO -DSQLITE_HAS_CODEC -DSQLITE_OMIT_LOAD_EXTENSION -DSQLITE_TEMP_STORE=2
// #cgo openbsd CPPFLAGS: -DOMIT_MEMLOCK
//
// #include <stdlib.h>
//
// #include "sqlite3.h"
//
// typedef void (*bind_destructor)(void *);
// typedef int (*exec_callback)(void *, int, char **, char **);
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

const (
	OpenReadOnly     = C.SQLITE_OPEN_READONLY
	OpenReadWrite    = C.SQLITE_OPEN_READWRITE
	OpenCreate       = C.SQLITE_OPEN_CREATE
	OpenURI          = C.SQLITE_OPEN_URI
	OpenMemory       = C.SQLITE_OPEN_MEMORY
	OpenNoMutex      = C.SQLITE_OPEN_NOMUTEX
	OpenFullMutex    = C.SQLITE_OPEN_FULLMUTEX
	OpenSharedCache  = C.SQLITE_OPEN_SHAREDCACHE
	OpenPrivateCache = C.SQLITE_OPEN_PRIVATECACHE
	OpenNoFollow     = C.SQLITE_OPEN_NOFOLLOW
	OpenExResCode    = C.SQLITE_OPEN_EXRESCODE
)

type DB struct {
	db *C.sqlite3
}

func Open(path string) (*DB, error) {
	pathCS := C.CString(path)
	defer C.free(unsafe.Pointer(pathCS))

	var db DB
	if C.sqlite3_open(pathCS, &db.db) != C.SQLITE_OK {
		defer C.sqlite3_close(db.db)
		return nil, db.errorf("cannot open database: %s", path)
	}

	return &db, nil
}

func OpenFlags(path string, flags int) (*DB, error) {
	pathCS := C.CString(path)
	defer C.free(unsafe.Pointer(pathCS))

	var db DB
	if C.sqlite3_open_v2(pathCS, &db.db, C.int(flags), (*C.char)(C.NULL)) != C.SQLITE_OK {
		defer C.sqlite3_close(db.db)
		return nil, db.errorf("cannot open database: %s", path)
	}

	return &db, nil
}

func (db *DB) Close() error {
	if C.sqlite3_close(db.db) != C.SQLITE_OK {
		return db.errorf("cannot close database")
	}
	return nil
}

func (db *DB) Key(key []byte) error {
	if C.sqlite3_key(db.db, unsafe.Pointer(&key[0]), C.int(len(key))) != C.SQLITE_OK {
		return db.errorf("cannot set key")
	}
	return nil
}

func (db *DB) KeyDatabase(dbName string, key []byte) error {
	dbNameCS := C.CString(dbName)
	defer C.free(unsafe.Pointer(dbNameCS))

	if C.sqlite3_key_v2(db.db, dbNameCS, unsafe.Pointer(&key[0]), C.int(len(key))) != C.SQLITE_OK {
		return db.errorf("cannot set key for database %s", dbName)
	}
	return nil
}

func (db *DB) Rekey(key []byte) error {
	if C.sqlite3_rekey(db.db, unsafe.Pointer(&key[0]), C.int(len(key))) != C.SQLITE_OK {
		return db.errorf("cannot change key")
	}
	return nil
}

func (db *DB) RekeyDatabase(dbName string, key []byte) error {
	dbNameCS := C.CString(dbName)
	defer C.free(unsafe.Pointer(dbNameCS))

	if C.sqlite3_rekey_v2(db.db, dbNameCS, unsafe.Pointer(&key[0]), C.int(len(key))) != C.SQLITE_OK {
		return db.errorf("cannot change key for database %s", dbName)
	}
	return nil
}

func (db *DB) Exec(sql string) error {
	sqlCS := C.CString(sql)
	defer C.free(unsafe.Pointer(sqlCS))

	var errMsgCS *C.char
	if C.sqlite3_exec(db.db, sqlCS, C.exec_callback(C.NULL), C.NULL, &errMsgCS) != C.SQLITE_OK {
		defer C.sqlite3_free(unsafe.Pointer(errMsgCS))
		errMsg := C.GoString(errMsgCS)
		return errors.New("cannot execute SQL statement: " + errMsg)
	}
	return nil
}

func (db *DB) Execf(format string, a ...any) error {
	return db.Exec(fmt.Sprintf(format, a...))
}

func (db *DB) errorf(format string, a ...any) error {
	msg := C.GoString(C.sqlite3_errmsg(db.db))
	return errors.New(fmt.Sprintf(format, a...) + ": " + msg)
}

type Stmt struct {
	db   *DB
	stmt *C.sqlite3_stmt
	err  error
}

func (db *DB) Prepare(sql string) (*Stmt, string, error) {
	sqlCS := C.CString(sql)
	defer C.free(unsafe.Pointer(sqlCS))

	stmt := Stmt{db: db}
	var tailCS *C.char
	if C.sqlite3_prepare_v2(db.db, sqlCS, -1, &stmt.stmt, &tailCS) != C.SQLITE_OK {
		return nil, "", db.errorf("cannot prepare SQL statement")
	}

	off := uintptr(unsafe.Pointer(tailCS)) - uintptr(unsafe.Pointer(sqlCS))
	tail := sql[off:]

	return &stmt, tail, nil
}

func (s *Stmt) Bind(idx int, val any) error {
	switch val := val.(type) {
	case nil:
		return s.BindNull(idx)
	case int:
		return s.BindInt(idx, val)
	case int64:
		return s.BindInt64(idx, val)
	case float64:
		return s.BindDouble(idx, val)
	case string:
		return s.BindText(idx, val)
	case []byte:
		return s.BindBlob(idx, val)
	default:
		return fmt.Errorf("cannot bind variable of type %T", val)
	}
}

func (s *Stmt) BindNull(idx int) error {
	if C.sqlite3_bind_null(s.stmt, C.int(idx)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind null parameter")
	}
	return nil
}

func (s *Stmt) BindInt(idx, val int) error {
	if C.sqlite3_bind_int(s.stmt, C.int(idx), C.int(val)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind int parameter")
	}
	return nil
}

func (s *Stmt) BindInt64(idx int, val int64) error {
	if C.sqlite3_bind_int64(s.stmt, C.int(idx), C.sqlite3_int64(val)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind int64 parameter")
	}
	return nil
}

func (s *Stmt) BindDouble(idx int, val float64) error {
	if C.sqlite3_bind_double(s.stmt, C.int(idx), C.double(val)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind double parameter")
	}
	return nil
}

func (s *Stmt) BindText(idx int, val string) error {
	valCS := C.CString(val)
	if C.sqlite3_bind_text(s.stmt, C.int(idx), valCS, -1, C.bind_destructor(C.free)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind text parameter")
	}
	return nil
}

func (s *Stmt) BindBlob(idx int, val []byte) error {
	valCB := C.NULL
	if len(val) > 0 {
		valCB = C.CBytes(val)
	}
	if C.sqlite3_bind_blob(s.stmt, C.int(idx), valCB, C.int(len(val)), C.bind_destructor(C.free)) != C.SQLITE_OK {
		return s.db.errorf("cannot bind blob parameter")
	}
	return nil
}

func (s *Stmt) Step() bool {
	switch C.sqlite3_step(s.stmt) {
	case C.SQLITE_ROW:
		return true
	case C.SQLITE_DONE:
		return false
	default:
		s.err = s.db.errorf("cannot execute SQL statement")
		return false
	}
}

func (s *Stmt) Finalize() error {
	if s.err != nil {
		C.sqlite3_finalize(s.stmt)
		return s.err
	}
	if ret := C.sqlite3_finalize(s.stmt); ret != C.SQLITE_OK {
		msg := C.GoString(C.sqlite3_errstr(ret))
		return errors.New("cannot finalize SQL statement: " + msg)
	}
	return nil
}

type ColumnType int

const (
	ColumnTypeInteger ColumnType = iota
	ColumnTypeFloat
	ColumnTypeText
	ColumnTypeBlob
	ColumnTypeNull
)

func (s *Stmt) ColumnType(idx int) ColumnType {
	switch t := C.sqlite3_column_type(s.stmt, C.int(idx)); t {
	case C.SQLITE_INTEGER:
		return ColumnTypeInteger
	case C.SQLITE_FLOAT:
		return ColumnTypeFloat
	case C.SQLITE_TEXT:
		return ColumnTypeText
	case C.SQLITE_BLOB:
		return ColumnTypeBlob
	case C.SQLITE_NULL:
		return ColumnTypeNull
	default:
		panic(fmt.Sprintf("sqlite: unexpected column type: %d", int(t)))
	}
}

func (s *Stmt) ColumnInt(idx int) int {
	return int(C.sqlite3_column_int(s.stmt, C.int(idx)))
}

func (s *Stmt) ColumnInt64(idx int) int64 {
	return int64(C.sqlite3_column_int64(s.stmt, C.int(idx)))
}

func (s *Stmt) ColumnDouble(idx int) float64 {
	return float64(C.sqlite3_column_double(s.stmt, C.int(idx)))
}

func (s *Stmt) ColumnText(idx int) string {
	if s.ColumnType(idx) == ColumnTypeNull {
		return ""
	}

	text := C.sqlite3_column_text(s.stmt, C.int(idx))
	if text == (*C.uchar)(C.NULL) {
		// sqlite3_column_text() returns NULL if the column type is
		// NULL or an error occurred. We already checked the column
		// type, so an error must have occurred.
		//
		// If an error occurred, it should be an out-of-memory error,
		// so panic.
		msg := C.GoString(C.sqlite3_errstr(C.sqlite3_errcode(s.db.db)))
		panic("sqlite: cannot get column text: " + msg)
	}

	// The C string returned by sqlite3_column_text() might contain
	// embedded null characters, so don't use C.GoString
	n := C.sqlite3_column_bytes(s.stmt, C.int(idx))
	return string(C.GoBytes(unsafe.Pointer(text), n))
}

func (s *Stmt) ColumnBlob(idx int) []byte {
	blob := C.sqlite3_column_blob(s.stmt, C.int(idx))
	if blob == C.NULL {
		// sqlite3_column_blob() returns NULL if the column type is
		// NULL, the blob has zero length or an error occurred. If
		// sqlite3_errcode() returns a result code other than
		// SQLITE_OK, we assume that an error occurred during the call
		// to sqlite3_column_blob().
		//
		// If an error occurred, it should be an out-of-memory error,
		// so panic.
		if code := C.sqlite3_errcode(s.db.db); code != C.SQLITE_OK {
			msg := C.GoString(C.sqlite3_errstr(code))
			panic("sqlite: cannot get column blob: " + msg)
		}
	}

	// C.GoBytes handles NULL if n is 0. See gobytes in
	// $GOROOT/src/runtime/string.go, which is called by _Cfunc_GoBytes in
	// $GOROOT/src/cmd/cgo/out.go.
	n := C.sqlite3_column_bytes(s.stmt, C.int(idx))
	return C.GoBytes(blob, n)
}

func (s *Stmt) ColumnCount() int {
	return int(C.sqlite3_column_count(s.stmt))
}

type Backup struct {
	db     *DB
	backup *C.sqlite3_backup
	err    error
}

func NewBackup(dstDB *DB, dstName string, srcDB *DB, srcName string) (*Backup, error) {
	dstNameCS := C.CString(dstName)
	srcNameCS := C.CString(srcName)
	defer C.free(unsafe.Pointer(dstNameCS))
	defer C.free(unsafe.Pointer(srcNameCS))

	backup := C.sqlite3_backup_init(dstDB.db, dstNameCS, srcDB.db, srcNameCS)
	if backup == (*C.sqlite3_backup)(C.NULL) {
		return nil, dstDB.errorf("cannot initialize backup")
	}

	return &Backup{db: dstDB, backup: backup}, nil
}

func (b *Backup) Step(nPages int) bool {
	switch ret := C.sqlite3_backup_step(b.backup, C.int(nPages)); ret {
	case C.SQLITE_OK:
		return true
	case C.SQLITE_DONE:
		return false
	default:
		msg := C.GoString(C.sqlite3_errstr(ret))
		b.err = errors.New("cannot backup database: " + msg)
		return false
	}
}

func (b *Backup) Finish() error {
	if b.err != nil {
		C.sqlite3_backup_finish(b.backup)
		return b.err
	}
	if ret := C.sqlite3_backup_finish(b.backup); ret != C.SQLITE_OK {
		msg := C.GoString(C.sqlite3_errstr(ret))
		return errors.New("cannot finish backup: " + msg)
	}
	return nil
}
