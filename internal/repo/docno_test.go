package repo

import "testing"

func TestLooksLikeFullPONo(t *testing.T) {
	if !LooksLikeFullPONo("PO202610040001") {
		t.Fatal("full po")
	}
	if LooksLikeFullPONo("PO2026") {
		t.Fatal("partial should fuzzy")
	}
	if LooksLikeFullPONo("OC202610040001") {
		t.Fatal("not a po")
	}
}
