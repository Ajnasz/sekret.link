package maxreads

import "testing"

func Test_Parse(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{"empty defaults to 1", "", 1, false},
		{"valid value", "5", 5, false},
		{"zero is invalid", "0", 0, true},
		{"negative is invalid", "-1", 0, true},
		{"non numeric is invalid", "abc", 0, true},
		{"max smallint is valid", "32767", 32767, false},
		{"above max smallint is invalid", "32768", 0, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Parse(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != c.want {
				t.Errorf("expected %d, got %d", c.want, got)
			}
		})
	}
}
