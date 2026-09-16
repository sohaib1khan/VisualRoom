package humanize

import "testing"

func TestBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{12 * 1024 * 1024 * 1024, "12 GB"},
	}
	for _, tc := range cases {
		if got := Bytes(tc.in); got != tc.want {
			t.Errorf("Bytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPercent(t *testing.T) {
	if Percent(50, 100) != 50 {
		t.Fatal("expected 50%")
	}
	if Percent(1, 0) != 0 {
		t.Fatal("zero total should be 0")
	}
}

func TestHealthHeadline(t *testing.T) {
	if HealthHeadline(80) != "cruising fine" {
		t.Fatal(HealthHeadline(80))
	}
	if HealthHeadline(1) != "about to seize up" {
		t.Fatal(HealthHeadline(1))
	}
}
