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

package safestorage

// #cgo LDFLAGS: -framework CoreFoundation -framework Security
//
// #include <CoreFoundation/CoreFoundation.h>
// #include <Security/Security.h>
import "C"

import (
	"fmt"
	"unsafe"
)

func (a *App) setEncryptionKeyFromSystem() error {
	query := C.CFDictionaryCreateMutable(C.kCFAllocatorDefault, 0, &C.kCFTypeDictionaryKeyCallBacks, &C.kCFTypeDictionaryValueCallBacks)
	if query == C.CFMutableDictionaryRef(C.NULL) {
		return fmt.Errorf("cannot create dictionary")
	}
	defer C.CFRelease(C.CFTypeRef(query))

	service := cfString(a.name + macosServiceSuffix)
	defer C.CFRelease(C.CFTypeRef(service))

	C.CFDictionaryAddValue(query, unsafe.Pointer(C.kSecClass), unsafe.Pointer(C.kSecClassGenericPassword))
	C.CFDictionaryAddValue(query, unsafe.Pointer(C.kSecAttrService), unsafe.Pointer(service))
	C.CFDictionaryAddValue(query, unsafe.Pointer(C.kSecReturnData), unsafe.Pointer(C.kCFBooleanTrue))

	var result C.CFTypeRef
	status := C.SecItemCopyMatching(C.CFDictionaryRef(query), &result)
	if status == C.errSecItemNotFound {
		return fmt.Errorf("cannot find encryption key")
	}
	if status != C.errSecSuccess {
		return fmt.Errorf("cannot get encryption key: error %d", status)
	}
	defer C.CFRelease(result)

	data := C.CFDataGetBytePtr(C.CFDataRef(result))
	dataLen := C.CFDataGetLength(C.CFDataRef(result))

	a.rawKey = RawEncryptionKey{
		Key: C.GoBytes(unsafe.Pointer(data), C.int(dataLen)),
		OS:  "macos",
	}
	a.key = deriveEncryptionKey(a.rawKey.Key, macosIterations)

	return nil
}

func cfString(s string) C.CFStringRef {
	b := []byte(s)
	cfs := C.CFStringCreateWithBytes(C.kCFAllocatorDefault, (*C.UInt8)(&b[0]), C.CFIndex(len(b)), C.kCFStringEncodingUTF8, C.false)
	if cfs == C.CFStringRef(C.NULL) {
		panic("cannot create CFString")
	}
	return cfs
}
