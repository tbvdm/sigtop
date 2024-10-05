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

const (
	keySize = 16 // AES-128
	salt    = "saltysalt"

	macosCiphertextPrefix = "v10"
	macosIterations       = 1003
	macosServiceSuffix    = " Safe Storage"

	linuxCiphertextPrefix = "v11"
	linuxIterations       = 1

	libsecretSchema = "chrome_libsecret_os_crypt_password_v2"

	kwallet4Service          = "org.kde.kwalletd"
	kwallet5Service          = "org.kde.kwalletd5"
	kwallet6Service          = "org.kde.kwalletd6"
	kwallet4Path             = "/modules/kwalletd"
	kwallet5Path             = "/modules/kwalletd5"
	kwallet6Path             = "/modules/kwalletd6"
	kwalletInterface         = "org.kde.KWallet"
	kwalletFolder            = "Chromium Keys"
	kwalletEntry             = "Chromium Safe Storage"
	kwalletInvalidHandle     = -1
	kwalletEntryTypePassword = 1

	windowsCiphertextPrefix = "v10"
	windowsDPAPIKeyPrefix   = "DPAPI"
	windowsKeySize          = 32 // AES-256
	windowsNonceSize        = 12

	localStateFile = "Local State"
)
