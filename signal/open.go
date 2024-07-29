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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tbvdm/sigtop/safestorage"
	"github.com/tbvdm/sigtop/sqlcipher"
)

type Context struct {
	dir                        string
	db                         *sqlcipher.DB
	dbVersion                  int
	recipientsByConversationID map[string]*Recipient
	recipientsByPhone          map[string]*Recipient
	recipientsByACI            map[string]*Recipient
}

func Open(dir string) (*Context, error) {
	return OpenWithPassword(dir, nil)
}

func OpenWithPassword(dir string, password []byte) (*Context, error) {
	dbFile := filepath.Join(dir, DatabaseFile)

	// SQLite/SQLCipher doesn't provide a useful error message if the
	// database doesn't exist or can't be read
	f, err := os.Open(dbFile)
	if err != nil {
		return nil, err
	}
	f.Close()

	db, err := sqlcipher.OpenFlags(dbFile, sqlcipher.OpenReadOnly)
	if err != nil {
		return nil, err
	}

	key, err := databaseKey(dir, password)
	if err != nil {
		return nil, err
	}

	// Format the key as an SQLite blob literal
	key = []byte(fmt.Sprintf("x'%s'", string(key)))

	if err := db.Key(key); err != nil {
		db.Close()
		return nil, err
	}

	// Verify key
	if err := db.Exec("SELECT count(*) FROM sqlite_master"); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot verify key: %w", err)
	}

	dbVersion, err := databaseVersion(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	if dbVersion < 19 {
		db.Close()
		return nil, fmt.Errorf("database version %d not supported (yet)", dbVersion)
	}

	ctx := Context{
		dir:       dir,
		db:        db,
		dbVersion: dbVersion,
	}

	return &ctx, nil
}

func (c *Context) Close() {
	c.db.Close()
}

func databaseKey(dir string, password []byte) ([]byte, error) {
	configFile := filepath.Join(dir, ConfigFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config struct {
		LegacyKey *string `json:"key"`
		ModernKey *string `json:"encryptedKey"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", configFile, err)
	}

	if password == nil && config.LegacyKey != nil {
		return []byte(*config.LegacyKey), nil
	}

	if config.ModernKey == nil {
		return nil, fmt.Errorf("encrypted database key not found")
	}

	key, err := hex.DecodeString(*config.ModernKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted database key: %w", err)
	}

	var dbKey []byte
	if password != nil {
		dbKey, err = safestorage.DecryptWithPassword(key, password)
	} else {
		err = fmt.Errorf("not yet supported")
	}
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt database key: %w", err)
	}

	return dbKey, nil
}
