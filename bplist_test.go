package plist

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"math"
	"testing"
)

func BenchmarkBplistGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := newBplistGenerator(ioutil.Discard)
		d.generateDocument(plistValueTree)
	}
}

func BenchmarkBplistParse(b *testing.B) {
	buf := bytes.NewReader(plistValueTreeAsBplist)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		d := newBplistParser(buf)
		d.parseDocument()
		b.StopTimer()
		buf.Seek(0, 0)
	}
}

func TestBplistInt128(t *testing.T) {
	bplist := []byte{0x62, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x30, 0x30, 0x14, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x19}
	expected := uint64(0x090a0b0c0d0e0f10)
	buf := bytes.NewReader(bplist)
	d := newBplistParser(buf)
	pval, _ := d.parseDocument()
	if pinteger, ok := pval.(*cfNumber); !ok || pinteger.value != expected {
		t.Error("Expected", expected, "received", pval)
	}
}

func TestBplistSignedIntValues(t *testing.T) {
	bplist := []byte{
		'b', 'p', 'l', 'i', 's', 't', '0', '0',

		// Array (8 entries)
		0xA8,
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,

		// 0xFFFFFFFFFFFFFF80 (MinInt8, sign extended)
		0x13, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x80,

		// 0x7F (MaxInt8)
		0x10, 0x7f,

		// 0xFFFFFFFFFFFF8000 (MinInt16, sign extended)
		0x13, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x80, 0x00,

		// 0x7FFF (MaxInt16)
		0x11, 0x7f, 0xff,

		// 0xFFFFFFFF80000000 (MinInt32, sign extended)
		0x13, 0xff, 0xff, 0xff, 0xff, 0x80, 0x00, 0x00, 0x00,

		// 0x7FFFFFFF (MaxInt32)
		0x12, 0x7f, 0xff, 0xff, 0xff,

		// 0x8000000000000000 (MinInt64)
		0x13, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,

		// 0x7FFFFFFFFFFFFFFF (MaxInt64)
		0x13, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,

		// Offset table
		0x08, 0x11, 0x1a, 0x1c, 0x25, 0x28, 0x31, 0x36, 0x3f,

		// Trailer
		0x00, 0x00, 0x00, 0x00, 0x00,
		0x00,
		0x01,
		0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x48,
	}

	expectedValues := []int64{
		math.MinInt8,
		math.MaxInt8,
		math.MinInt16,
		math.MaxInt16,
		math.MinInt32,
		math.MaxInt32,
		math.MinInt64,
		math.MaxInt64,
	}

	buf := bytes.NewReader(bplist)
	d := newBplistParser(buf)
	pval, _ := d.parseDocument()
	parsedValues := pval.(*cfArray).values
	for i, cfv := range parsedValues {
		value := int64(cfv.(*cfNumber).value)
		if value != expectedValues[i] {
			t.Error("Expected", expectedValues[i], "received", value)
		}
	}
}

func TestBplistLatin1ToUTF16(t *testing.T) {
	expectedPrefix := []byte{0x62, 0x70, 0x6c, 0x69, 0x73, 0x74, 0x30, 0x30, 0xd1, 0x01, 0x02, 0x51, 0x5f, 0x6f, 0x10, 0x80}
	expectedPostfix := []byte{0x00, 0x08, 0x00, 0x0b, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x10}
	expectedBuf := bytes.NewBuffer(expectedPrefix)

	sBuf := &bytes.Buffer{}
	for i := uint16(0xc280); i <= 0xc2bf; i++ {
		binary.Write(sBuf, binary.BigEndian, i)
		binary.Write(expectedBuf, binary.BigEndian, i-0xc200)
	}

	for i := uint16(0xc380); i <= 0xc3bf; i++ {
		binary.Write(sBuf, binary.BigEndian, i)
		binary.Write(expectedBuf, binary.BigEndian, i-0xc300+0x0040)
	}

	expectedBuf.Write(expectedPostfix)

	var buf bytes.Buffer
	encoder := NewBinaryEncoder(&buf)

	data := map[string]string{
		"_": string(sBuf.Bytes()),
	}
	if err := encoder.Encode(data); err != nil {
		t.Error(err.Error())
	}

	if !bytes.Equal(buf.Bytes(), expectedBuf.Bytes()) {
		t.Error("Expected", expectedBuf.Bytes(), "received", buf.Bytes())
		return
	}
}

func TestBplistSharedArrays(t *testing.T) {
	// Five arrays: an empty one, then four that each hold 14 references to the one before.
	data := []byte("bplist00\xA0")
	offsets := []byte{8}
	for id := byte(0); id < 4; id++ {
		offsets = append(offsets, byte(len(data)))
		data = append(data, 0xAE)
		data = append(data, bytes.Repeat([]byte{id}, 14)...)
	}
	table := byte(len(data))
	data = append(data, offsets...)
	data = append(data, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, table)

	type arrays []arrays
	var decoded interface{}
	var typed arrays
	allocs := testing.AllocsPerRun(1, func() {
		decoded, typed = nil, nil
		for _, v := range []interface{}{&decoded, &typed} {
			if _, err := Unmarshal(data, v); err != nil {
				t.Fatal(err)
			}
		}
	})
	// Decoding each array once per reference takes over 14^4 allocations.
	if allocs > 1000 {
		t.Errorf("decoding %d bytes took %v allocations", len(data), allocs)
	}

	for level := 0; level < 4; level++ {
		array, _ := decoded.([]interface{})
		if len(array) != 14 || len(typed) != 14 {
			t.Fatalf("level %d: wanted arrays of 14 values, received %v and %v", level, decoded, typed)
		}
		decoded, typed = array[13], typed[13]
	}
}

func TestBplistSharedArrayNestingDepth(t *testing.T) {
	// Array 0 is empty, array i holds array i-1, and the top array holds them all in that order.
	// The parser finds each of them already parsed, and never has more than three arrays open.
	for _, count := range []int{maxCollectionDepth - 2, maxCollectionDepth - 1} {
		var data bytes.Buffer
		data.WriteString("bplist00\xA0")
		offsets := []uint16{8}
		top := []uint16{0xAF11, uint16(count)}
		for id := 1; id <= count; id++ {
			offsets = append(offsets, uint16(data.Len()))
			data.WriteByte(0xA1)
			binary.Write(&data, binary.BigEndian, uint16(id-1))
			top = append(top, uint16(id))
		}
		offsets = append(offsets, uint16(data.Len()))
		binary.Write(&data, binary.BigEndian, top)
		table := data.Len()
		binary.Write(&data, binary.BigEndian, offsets)
		binary.Write(&data, binary.BigEndian, []uint64{2<<8 | 2, uint64(count + 2), uint64(count + 1), uint64(table)})

		// The top array nests count+2 arrays.
		var decoded interface{}
		_, err := Unmarshal(data.Bytes(), &decoded)
		if wantErr := count+2 > maxCollectionDepth; (err != nil) != wantErr {
			t.Errorf("%d nested arrays: wanted error: %v, received: %v", count+2, wantErr, err)
		}
	}
}
