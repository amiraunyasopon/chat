package protocol

import "testing"

func TestParseDMRequiresMessage(t *testing.T) {
	for _, args := range []string{"alice", "alice ", "alice   "} {
		if _, ok := ParseDM(args); ok {
			t.Errorf("ParseDM(%q) accepted a missing message", args)
		}
	}
}

func TestParseDM(t *testing.T) {
	dm, ok := ParseDM("alice hello there")
	if !ok {
		t.Fatal("ParseDM rejected a valid direct message")
	}
	if dm.To != "alice" || dm.Message != "hello there" {
		t.Fatalf("ParseDM returned %#v", dm)
	}
}
