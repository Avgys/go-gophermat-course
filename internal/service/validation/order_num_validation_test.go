package validation

import "testing"

func TestLuhnNumVerify(t *testing.T) {
	tests := []struct {
		name    string
		num     string
		wantErr bool
	}{
		{name: "valid short", num: "79927398713", wantErr: false},
		{name: "valid long", num: "12345678903", wantErr: false},
		{name: "invalid checksum", num: "79927398710", wantErr: true},
		{name: "non digit", num: "1234a567", wantErr: true},
		{name: "empty", num: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LuhnNumVerify(tt.num)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
