package hid

import "testing"

func TestCell_SetChar(t *testing.T) {
	type fields struct {
		char     rune
		prevChar rune
		isDirty  bool
	}
	type args struct {
		char rune
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{"aaaaa",
			fields{'a', 'b', false},
			args{'a'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := &Cell{
				char:     tt.fields.char,
				prevChar: tt.fields.prevChar,
				isDirty:  tt.fields.isDirty,
			}
			cell.SetChar(tt.args.char)
		})
	}
}

