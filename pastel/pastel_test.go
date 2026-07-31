package pastel

import "testing"

func TestAtEndpoints(t *testing.T) {
	if got := Hex(0); got != Start {
		t.Fatalf("Hex(0)=%s want %s", got, Start)
	}
	if got := Hex(1); got != End {
		t.Fatalf("Hex(1)=%s want %s", got, End)
	}
	mid := Hex(0.5)
	if mid == Start || mid == End {
		t.Fatalf("mid should differ from endpoints, got %s", mid)
	}
}
