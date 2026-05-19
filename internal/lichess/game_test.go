package lichess

import "testing"

func TestParseGameURL(t *testing.T) {
	ref, err := ParseGameURL("https://lichess.org/F8M8OS71#53")
	if err != nil {
		t.Fatal(err)
	}
	if ref.ID != "F8M8OS71" || ref.Ply != 53 {
		t.Fatalf("got %+v", ref)
	}

	ref, err = ParseGameURL("https://lichess.org/787zsVup/black#48")
	if err != nil || ref.ID != "787zsVup" || ref.Ply != 48 {
		t.Fatalf("black path: %+v err=%v", ref, err)
	}
}
