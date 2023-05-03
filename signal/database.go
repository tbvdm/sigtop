// Copyright (c) 2021, 2023 Tim van der Molen <tim@kariliq.nl>
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

package signal

import (
	"fmt"

	"github.com/tbvdm/sigtop/sqlcipher"
)

func (c *Context) CheckDatabase() ([]string, error) {
	results, err := runPragmaCheck(c.db, "cipher_integrity_check")
	if err != nil || len(results) > 0 {
		return results, err
	}

	integrityResults, err := runPragmaCheck(c.db, "integrity_check")
	if err != nil {
		return nil, err
	}

	foreignKeyResults, err := runPragmaCheck(c.db, "foreign_key_check")
	if err != nil {
		return nil, err
	}

	return append(integrityResults, foreignKeyResults...), nil
}

func runPragmaCheck(db *sqlcipher.DB, pragma string) ([]string, error) {
	stmt, err := db.Prepare("PRAGMA " + pragma)
	if err != nil {
		return nil, err
	}

	var results []string
	switch pragma {
	case "cipher_integrity_check":
		for stmt.Step() {
			results = append(results, stmt.ColumnText(0))
		}
	case "integrity_check":
		for stmt.Step() {
			results = append(results, stmt.ColumnText(0))
		}
		if len(results) == 1 && results[0] == "ok" {
			results = nil
		}
	case "foreign_key_check":
		for stmt.Step() {
			var s string
			if stmt.ColumnType(1) == sqlcipher.ColumnTypeNull {
				s = fmt.Sprintf("foreign key violation in table %s", stmt.ColumnText(0))
			} else {
				s = fmt.Sprintf("foreign key violation in row %d of table %s", stmt.ColumnInt64(1), stmt.ColumnText(0))
			}
			results = append(results, s)
		}
	default:
		stmt.Finalize()
		return nil, fmt.Errorf("invalid check: %s", pragma)
	}

	return results, stmt.Finalize()
}

func (c *Context) WriteDatabase(path string) error {
	// To decrypt an encrypted database to a plaintext database, the
	// SQLCipher documentation recommends to do the following:
	//
	// 1. Open the encrypted database
	// 2. Attach the plaintext database
	// 3. Use the sqlcipher_export() SQL function to decrypt
	//
	// This doesn't work in our case, because we insist on opening the
	// Signal Desktop database in read-only mode. The SQLite backup API
	// doesn't work either, because it does not support
	// encrypted-to-plaintext backups.
	//
	// However, since SQLCipher 4.3.0, the backup API does support
	// encrypted-to-encrypted backups. This allows us to do the following:
	//
	// 1. Open the Signal Desktop database in read-only mode
	// 2. Create a temporary, encrypted database in memory
	// 3. Back up the Signal Desktop database to the temporary database
	// 4. Attach a new plaintext database to the temporary database
	// 5. Use sqlcipher_export() to decrypt the temporary database to the
	//    plaintext database

	// Create a temporary, encrypted database in memory
	db, err := sqlcipher.Open(":memory:")
	if err != nil {
		return err
	}
	defer db.Close()
	// Set a dummy key to enable encryption
	if err := db.Key([]byte("x")); err != nil {
		return err
	}

	// Back up the Signal Desktop database to the temporary database
	backup, err := sqlcipher.NewBackup(db, "main", c.db, "main")
	if err != nil {
		return err
	}
	backup.Step(-1)
	if err := backup.Finish(); err != nil {
		return err
	}

	// Attach a new plaintext database to the temporary database
	stmt, err := db.Prepare("ATTACH DATABASE ? AS plaintext KEY ''")
	if err != nil {
		return err
	}
	if err := stmt.Bind(1, path); err != nil {
		stmt.Finalize()
		return err
	}
	stmt.Step()
	if err := stmt.Finalize(); err != nil {
		return err
	}

	// Decrypt the temporary database to the plaintext database
	if err := db.Exec("BEGIN TRANSACTION"); err != nil {
		return err
	}
	if err := db.Exec("SELECT sqlcipher_export('plaintext')"); err != nil {
		return err
	}
	if err := setDatabaseVersion(db, "plaintext", c.dbVersion); err != nil {
		return err
	}
	if err := db.Exec("END TRANSACTION"); err != nil {
		return err
	}
	if err := db.Exec("DETACH DATABASE plaintext"); err != nil {
		return err
	}

	return nil
}

func databaseVersion(db *sqlcipher.DB) (int, error) {
	stmt, err := db.Prepare("PRAGMA user_version")
	if err != nil {
		return 0, err
	}

	var version int
	if stmt.Step() {
		version = stmt.ColumnInt(0)
	}

	return version, stmt.Finalize()
}

func setDatabaseVersion(db *sqlcipher.DB, schema string, version int) error {
	return db.Execf("PRAGMA %s.user_version = %d", schema, version)
}
