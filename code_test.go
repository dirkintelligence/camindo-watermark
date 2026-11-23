package main

import "testing"

func TestCodeRoundTrip(t *testing.T) {
	for _, c := range []string{"A", "Z", "-", "GS-4X", "GSNORD", "------", "ZZZZZZ", "ab-cd"} {
		id, err := codeToID(c)
		if c == "GS-4X" {
			if err == nil {
				t.Errorf("%q accepted", c)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", c, err)
		}
		back := idToCode(id)
		want := c
		if c == "ab-cd" {
			want = "AB-CD"
		}
		if back == nil || *back != want {
			t.Errorf("%q -> %d -> %v", c, id, back)
		}
	}
	if _, err := codeToID("GSNORDX"); err == nil {
		t.Error("7 chars accepted")
	}
	if idToCode(1) != nil || idToCode(4294967295) != nil {
		t.Error("gap/too large decoded")
	}
	n := 0
	for id := uint32(0); id < 28*28*28*28*28*28; id += 997 {
		if c := idToCode(id); c != nil {
			if back, _ := codeToID(*c); back != id {
				t.Fatalf("id %d -> %q -> %d", id, *c, back)
			}
			n++
		}
	}
	t.Logf("checked %d codes", n)
}

func TestCodec(t *testing.T) {
	id, err := code{}.Encode("dr-cam")
	if err != nil {
		t.Fatal(err)
	}
	if s, ok := (code{}).Decode(id); !ok || s != "DR-CAM" {
		t.Errorf("decode %d = %q %v", id, s, ok)
	}
	if _, ok := (code{}).Decode(4294967295); ok {
		t.Error("too large decoded")
	}
}
