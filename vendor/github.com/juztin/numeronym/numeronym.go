// Copyright 2016 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Numeronyms parses ASCII text into numeronyms.

package numeronym

import (
	"bytes"
	"strconv"
)

func isAlphabetic(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z':
		return true
	case b >= 'a' && b <= 'z':
		return true
	}
	return false
}

// Parse returns numeronym(s) of the given ASCII bytes.
// For example:
//  "internationalization" => "i18n"
//  "ab_cdefg"             => "ab_c3g"
func Parse(b []byte) []byte {
	num := len(b) - 1
	buf := new(bytes.Buffer)
	var isAlpha bool
	for index, i := 0, 0; i < len(b); i++ {
		isAlpha = isAlphabetic(b[i])
		// Skip alpha characters, unless we've reached the end.
		if isAlpha && i < num {
			continue
		}
		switch {
		case isAlpha && i-index > 1: // Numeronym on last iteration.
			buf.WriteByte(b[index])
			buf.Write([]byte(strconv.Itoa(i - index - 1)))
		case i-index > 1: // Two characters OR a numeronym.
			buf.WriteByte(b[index])
			if i-index > 2 {
				buf.Write([]byte(strconv.Itoa(i - index - 2)))
			}
			buf.WriteByte(b[i-1])
		case i-index > 0: // Single character.
			buf.WriteByte(b[i-1])
		}
		buf.WriteByte(b[i]) // Current character.
		index = i + 1
	}
	return buf.Bytes()
}
