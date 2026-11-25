// +build !appengine

package plist

import (
	"reflect"
	"unsafe"
)

// copies8BitStrings reports whether zeroCopy8BitString copies the string out of buf.
const copies8BitStrings = false

func zeroCopy8BitString(buf []byte, off int, len int) string {
	if len == 0 {
		return ""
	}

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(unsafe.Pointer(&buf[off]))
	hdr.Len = len
	return s
}
