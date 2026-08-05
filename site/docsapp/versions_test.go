package main

import "testing"

func TestMajorVersionMenu(t *testing.T) {
	in := []versionEntry{
		{Version: "v0.2.0", Path: "/"},
		{Version: "v0.1.0", Path: "/v0.1.0/"},
	}
	got := majorVersionMenu(in)
	if len(got) != 1 || got[0].Version != "v0.2.0" {
		t.Fatalf("v0-only menu: got %#v, want single v0.2.0", got)
	}

	in = []versionEntry{
		{Version: "v1.0.0", Path: "/"},
		{Version: "v0.9.1", Path: "/v0.9.1/"},
		{Version: "v0.8.0", Path: "/v0.8.0/"},
	}
	got = majorVersionMenu(in)
	if len(got) != 2 {
		t.Fatalf("two majors: got %#v", got)
	}
	if got[0].Version != "v1.0.0" || got[1].Version != "v0.9.1" {
		t.Fatalf("order/pick: got %#v", got)
	}
}

func TestVersionNewer(t *testing.T) {
	if !versionNewer("v0.2.0", "v0.1.9") {
		t.Fatal("0.2.0 should be newer than 0.1.9")
	}
	if versionNewer("v0.1.0", "v0.2.0") {
		t.Fatal("0.1.0 should not be newer than 0.2.0")
	}
}
