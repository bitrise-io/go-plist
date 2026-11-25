// +build appengine

package plist

const copies8BitStrings = true

func zeroCopy8BitString(buf []byte, off int, len int) string {
	return string(buf[off : off+len])
}
