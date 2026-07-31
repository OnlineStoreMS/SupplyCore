package service

import "testing"

func TestParseRemarkPurchaseAmount(t *testing.T) {
	cases := []struct {
		in   string
		want float64
		ok   bool
	}{
		{"70", 70, true},
		{"70.5", 70.5, true},
		{"  63  ", 63, true},
		{"70元", 70, true},
		{"货款70", 70, true},
		{"运费5货款70", 70, true},
		{"", 0, false},
		{"无价格", 0, false},
	}
	for _, c := range cases {
		got, ok := ParseRemarkPurchaseAmount(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("ParseRemarkPurchaseAmount(%q)=(%v,%v) want (%v,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}
