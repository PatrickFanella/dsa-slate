package app

import "testing"

func TestServiceName(t *testing.T) {
	if got := ServiceName(); got != "dsa-slate" {
		t.Fatalf("unexpected service name: %s", got)
	}
}
