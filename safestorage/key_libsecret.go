// Copyright (c) 2024 Tim van der Molen <tim@kariliq.nl>
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

// Don't build on OpenBSD because libsecret requires extra pledge(2) promises
//
//go:build !(darwin || no_libsecret || openbsd || windows)

package safestorage

// #cgo pkg-config: libsecret-1
//
// #include <stdlib.h>
// #include <string.h>
//
// #include <libsecret/secret.h>
//
// gchar *
// secret_password_lookup_sync_wrapper(const SecretSchema *schema, GError **error, const char *name, const char *value)
// {
//	return secret_password_lookup_sync(schema, NULL, error, name, value, NULL);
// }
import "C"

import (
	"fmt"
	"unsafe"
)

func (a *App) rawEncryptionKeyFromLibsecret() ([]byte, error) {
	name := C.CString(libsecretSchema)
	defer C.free(unsafe.Pointer(name))

	attrName := C.CString("application")
	defer C.free(unsafe.Pointer(attrName))

	attrValue := C.CString(a.name)
	defer C.free(unsafe.Pointer(attrValue))

	schema := C.SecretSchema{
		name:  name,
		flags: C.SECRET_SCHEMA_NONE,
	}
	schema.attributes[0].name = attrName
	schema.attributes[0]._type = C.SECRET_SCHEMA_ATTRIBUTE_STRING
	schema.attributes[1].name = (*C.char)(C.NULL)
	schema.attributes[1]._type = 0

	gerr := (*C.GError)(C.NULL)
	result := C.secret_password_lookup_sync_wrapper(&schema, &gerr, attrName, attrValue)
	if gerr != (*C.GError)(C.NULL) {
		defer C.g_error_free(gerr)
		return nil, fmt.Errorf("cannot get encryption key: %s", C.GoString(gerr.message))
	}
	if result == (*C.char)(C.NULL) {
		return nil, fmt.Errorf("cannot find encryption key")
	}
	defer C.secret_password_free(result)

	return C.GoBytes(unsafe.Pointer(result), C.int(C.strlen(result))), nil
}
