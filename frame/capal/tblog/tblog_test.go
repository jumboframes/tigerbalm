package tblog

import "testing"

func TestTblog(t *testing.T) {
	tblog := NewTblog().WithLevel(Debug)
	tblog.Printf(Trace, "%s", "singchia watching 0")
	tblog.Printf(Debug, "%s", "singchia watching 1")
}
