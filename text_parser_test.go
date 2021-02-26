package plist

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_textPlistParser_parseDocument(t *testing.T) {
	tests := []struct {
		name     string
		parser   textPlistParser
		wantPval cfValue
		wantErr  bool
	}{
		{
			name: "Read string",
			parser: textPlistParser{
				reader: bytes.NewReader([]byte(`"abc"`)),
			},
			wantPval: cfString("abc"),
		},
		{
			name: "Map with empy key",
			parser: textPlistParser{
				reader: bytes.NewReader([]byte(`{""=Hello;}`)),
			},
			wantPval: &cfDictionary{
				keys:   []string{""},
				values: []cfValue{cfString("Hello")},
			},
		},
		{
			name: "Map with empy key",
			parser: textPlistParser{
				reader: bytes.NewReader([]byte(`{"A"="B";"C"="D";}`)),
			},
			wantPval: &cfDictionary{
				keys: []string{"A", "C", CustomAnnotationKey},
				values: []cfValue{
					cfString("B"),
					cfString("D"),
					&cfDictionary{
						keys: []string{"start", "end"},
						values: []cfValue{
							&cfNumber{value: 0, signed: true},
							&cfNumber{value: 18, signed: true},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPval, err := tt.parser.parseDocument()
			if (err != nil) != tt.wantErr {
				t.Errorf("textPlistParser.parseDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPval, tt.wantPval) {
				t.Errorf("textPlistParser.parseDocument() = %v, want %v", gotPval, tt.wantPval)
			}
		})
	}
}
