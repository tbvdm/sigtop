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
	encKey                     *safestorage.RawEncryptionKey
	dbKey                      []byte
	db                         *sqlcipher.DB
	dbVersion                  int
	recipientsByConversationID map[string]*Recipient
	recipientsByPhone          map[string]*Recipient
	recipientsByACI            map[string]*Recipient
}

func Open(betaApp bool, dir string, encKey *safestorage.RawEncryptionKey) (*Context, error) {
	appName := AppName
	if betaApp {
		appName = AppNameBeta
	}
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

	dbKey, encKey, err := databaseAndEncryptionKeys(appName, dir, encKey)
	if err != nil {
		return nil, err
	}

	// Format the key as an SQLite blob literal
	dbKeyBlob := []byte(fmt.Sprintf("x'%s'", string(dbKey)))

	if err := db.Key(dbKeyBlob); err != nil {
		db.Close()
		return nil, err
	}

	if err := db.Exec("PRAGMA cipher_log = stderr"); err != nil {
		db.Close()
		return nil, err
	}

	if err := db.Exec("SELECT count(*) FROM sqlite_master"); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot verify database key: %w", err)
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
		encKey:    encKey,
		dbKey:     dbKey,
		db:        db,
		dbVersion: dbVersion,
	}

	return &ctx, nil
}

func (c *Context) Close() {
	c.db.Close()
}

func databaseAndEncryptionKeys(appName, dir string, encKey *safestorage.RawEncryptionKey) ([]byte, *safestorage.RawEncryptionKey, error) {
	configFile := filepath.Join(dir, ConfigFile)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, err
	}

	var config struct {
		LegacyDBKey        *string `json:"key"`
		ModernDBKey        *string `json:"encryptedKey"`
		SafeStorageBackend *string `json:"safeStorageBackend"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("cannot parse %s: %w", configFile, err)
	}

	if config.LegacyDBKey != nil && encKey == nil {
		return []byte(*config.LegacyDBKey), nil, nil
	}

	if config.ModernDBKey == nil {
		return nil, nil, fmt.Errorf("encrypted database key not found")
	}

	encryptedDBKey, err := hex.DecodeString(*config.ModernDBKey)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid encrypted database key: %w", err)
	}

	app := safestorage.NewApp(appName, dir)
	if encKey != nil {
		if err := app.SetEncryptionKey(*encKey); err != nil {
			return nil, nil, fmt.Errorf("cannot set encryption key: %w", err)
		}
	} else if config.SafeStorageBackend != nil {
		if err := app.SetBackend(*config.SafeStorageBackend); err != nil {
			return nil, nil, fmt.Errorf("cannot set safeStorage backend: %w", err)
		}
	}

	dbKey, err := app.Decrypt(encryptedDBKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot decrypt database key: %w", err)
	}

	encKey, err = app.EncryptionKey()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get encryption key: %w", err)
	}

	return dbKey, encKey, nil
}

func (c *Context) EncryptionKey() ([]byte, error) {
	if c.encKey == nil || c.encKey.Key == nil {
		return nil, fmt.Errorf("encryption key not available")
	}
	return c.encKey.Key, nil
}

func (c *Context) DatabaseKey() ([]byte, error) {
	if c.dbKey == nil {
		return nil, fmt.Errorf("database key not available")
	}
	return c.dbKey, nil
}

func ImportEncryptionKey(betaApp bool, dir string, key []byte) error {
	appName := AppName
	if betaApp {
		appName = AppNameBeta
	}

	app := safestorage.NewApp(appName, dir)

	if err := app.StoreEncryptionKey(key); err != nil {
		return fmt.Errorf("cannot import encryption key: %w", err)
	}

	return nil
}
