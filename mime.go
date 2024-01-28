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

package main

import (
	"errors"
	"io/fs"
	"mime"
	"strings"

	"github.com/tbvdm/go-openbsd"
)

// unveilMimeFiles unveils files that the mime package tries to read. See the
// mimeGlobs and typeFiles slices in $GOROOT/src/mime.
func unveilMimeFiles() error {
	files := [...]string{
		"/etc/apache/mime.types",
		"/etc/apache2/mime.types",
		"/etc/httpd/conf/mime.types",
		"/etc/mime.types",
		"/usr/local/share/mime/globs2",
		"/usr/share/mime/globs2",
		"/usr/share/misc/mime.types",
	}

	for _, file := range files {
		err := openbsd.Unveil(file, "r")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	return nil
}

func extensionFromContentType(contentType string) (string, error) {
	// Avoid silly results, such as .jpe for image/jpeg
	switch t, _, _ := strings.Cut(contentType, ";"); t {
	case "image/jpeg":
		return ".jpg", nil
	case "video/mp4":
		return ".mp4", nil
	case "video/mpeg":
		return ".mpg", nil
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return "", err
	}

	return exts[0], nil
}
