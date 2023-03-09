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

// #include <stdlib.h>
//
// #include "sqlite3.h"
import "C"

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"unsafe"

	"golang.org/x/crypto/pbkdf2"
)

const (
	providerName    = "go"
	providerVersion = "0"

	cipherName      = "aes-256-cbc"
	cipherKeySize   = 32
	cipherBlockSize = aes.BlockSize
	cipherIVSize    = cipherBlockSize
)

var (
	initCount         int
	providerNameCS    *C.char
	providerVersionCS *C.char
	cipherNameCS      *C.char
)

//export sqlcipherGoInit
func sqlcipherGoInit() C.int {
	initCount++
	if initCount == 1 {
		providerNameCS = C.CString(providerName)
		providerVersionCS = C.CString(providerVersion)
		cipherNameCS = C.CString(cipherName)
	}
	return C.SQLITE_OK
}

//export sqlcipherGoFree
func sqlcipherGoFree() C.int {
	if initCount == 0 {
		return C.SQLITE_ERROR
	}
	initCount--
	if initCount == 0 {
		C.free(unsafe.Pointer(providerNameCS))
		C.free(unsafe.Pointer(providerVersionCS))
		C.free(unsafe.Pointer(cipherNameCS))
	}
	return C.SQLITE_OK
}

//export sqlcipherGoGetProviderName
func sqlcipherGoGetProviderName() *C.char {
	return providerNameCS
}

//export sqlcipherGoGetProviderVersion
func sqlcipherGoGetProviderVersion() *C.char {
	return providerVersionCS
}

//export sqlcipherGoAddRandom
func sqlcipherGoAddRandom(buf unsafe.Pointer, bufSize C.int) C.int {
	return C.SQLITE_OK
}

//export sqlcipherGoRandom
func sqlcipherGoRandom(buf unsafe.Pointer, bufSize C.int) C.int {
	bufSlice := unsafe.Slice((*byte)(buf), bufSize)
	if _, err := rand.Read(bufSlice); err != nil {
		return C.SQLITE_ERROR
	}
	return C.SQLITE_OK
}

//export sqlcipherGoHMACSHA1
func sqlcipherGoHMACSHA1(key *C.uchar, keySize C.int, in *C.uchar, inSize C.int, in2 *C.uchar, in2Size C.int, out *C.uchar) C.int {
	return sqlcipherGoHMAC(sha1.New, key, keySize, in, inSize, in2, in2Size, out)
}

//export sqlcipherGoHMACSHA256
func sqlcipherGoHMACSHA256(key *C.uchar, keySize C.int, in *C.uchar, inSize C.int, in2 *C.uchar, in2Size C.int, out *C.uchar) C.int {
	return sqlcipherGoHMAC(sha256.New, key, keySize, in, inSize, in2, in2Size, out)
}

//export sqlcipherGoHMACSHA512
func sqlcipherGoHMACSHA512(key *C.uchar, keySize C.int, in *C.uchar, inSize C.int, in2 *C.uchar, in2Size C.int, out *C.uchar) C.int {
	return sqlcipherGoHMAC(sha512.New, key, keySize, in, inSize, in2, in2Size, out)
}

func sqlcipherGoHMAC(h func() hash.Hash, key *C.uchar, keySize C.int, in *C.uchar, inSize C.int, in2 *C.uchar, in2Size C.int, out *C.uchar) C.int {
	keySlice := unsafe.Slice((*byte)(key), keySize)
	hmac := hmac.New(h, keySlice)

	inSlice := unsafe.Slice((*byte)(in), inSize)
	hmac.Write(inSlice)

	if unsafe.Pointer(in2) != C.NULL {
		in2Slice := unsafe.Slice((*byte)(in2), in2Size)
		hmac.Write(in2Slice)
	}

	outSlice := unsafe.Slice((*byte)(out), hmac.Size())
	copy(outSlice, hmac.Sum(nil))

	return C.SQLITE_OK
}

//export sqlcipherGoKDFSHA1
func sqlcipherGoKDFSHA1(pass *C.uchar, passSize C.int, salt *C.uchar, saltSize C.int, iter C.int, key *C.uchar, keySize C.int) C.int {
	return sqlcipherGoKDF(sha1.New, pass, passSize, salt, saltSize, iter, key, keySize)
}

//export sqlcipherGoKDFSHA256
func sqlcipherGoKDFSHA256(pass *C.uchar, passSize C.int, salt *C.uchar, saltSize C.int, iter C.int, key *C.uchar, keySize C.int) C.int {
	return sqlcipherGoKDF(sha256.New, pass, passSize, salt, saltSize, iter, key, keySize)
}

//export sqlcipherGoKDFSHA512
func sqlcipherGoKDFSHA512(pass *C.uchar, passSize C.int, salt *C.uchar, saltSize C.int, iter C.int, key *C.uchar, keySize C.int) C.int {
	return sqlcipherGoKDF(sha512.New, pass, passSize, salt, saltSize, iter, key, keySize)
}

func sqlcipherGoKDF(h func() hash.Hash, pass *C.uchar, passSize C.int, salt *C.uchar, saltSize C.int, iter C.int, key *C.uchar, keySize C.int) C.int {
	passSlice := unsafe.Slice((*byte)(pass), passSize)
	saltSlice := unsafe.Slice((*byte)(salt), saltSize)
	derivKey := pbkdf2.Key(passSlice, saltSlice, int(iter), int(keySize), h)

	keySlice := unsafe.Slice((*byte)(key), keySize)
	copy(keySlice, derivKey)

	return C.SQLITE_OK
}

//export sqlcipherGoCipher
func sqlcipherGoCipher(key *C.uchar, keySize C.int, iv *C.uchar, in *C.uchar, inSize C.int, out *C.uchar, encrypt C.int) C.int {
	if keySize != cipherKeySize {
		return C.SQLITE_ERROR
	}

	keySlice := unsafe.Slice((*byte)(key), keySize)
	aesCipher, err := aes.NewCipher(keySlice)
	if err != nil {
		return C.SQLITE_ERROR
	}

	ivSlice := unsafe.Slice((*byte)(iv), cipherIVSize)
	var cbcCipher cipher.BlockMode
	if encrypt != 0 {
		cbcCipher = cipher.NewCBCEncrypter(aesCipher, ivSlice)
	} else {
		cbcCipher = cipher.NewCBCDecrypter(aesCipher, ivSlice)
	}

	inSlice := unsafe.Slice((*byte)(in), inSize)
	outSlice := unsafe.Slice((*byte)(out), inSize)
	cbcCipher.CryptBlocks(outSlice, inSlice)

	return C.SQLITE_OK
}

//export sqlcipherGoGetCipher
func sqlcipherGoGetCipher() *C.char {
	return cipherNameCS
}

//export sqlcipherGoGetKeySize
func sqlcipherGoGetKeySize() C.int {
	return cipherKeySize
}

//export sqlcipherGoGetIVSize
func sqlcipherGoGetIVSize() C.int {
	return cipherIVSize
}

//export sqlcipherGoGetBlockSize
func sqlcipherGoGetBlockSize() C.int {
	return cipherBlockSize
}

//export sqlcipherGoGetHMACSizeSHA1
func sqlcipherGoGetHMACSizeSHA1() C.int {
	return sha1.Size
}

//export sqlcipherGoGetHMACSizeSHA256
func sqlcipherGoGetHMACSizeSHA256() C.int {
	return sha256.Size
}

//export sqlcipherGoGetHMACSizeSHA512
func sqlcipherGoGetHMACSizeSHA512() C.int {
	return sha512.Size
}

//export sqlcipherGoFIPSStatus
func sqlcipherGoFIPSStatus() C.int {
	return 0
}
