package frame

import (
	"net/http"
	"testing"
)

func TestFrame(t *testing.T) {
	frame, err := NewFrame("/Users/zhaizenghui/moresec/singchia/tigerbalm/js")
	if err != nil {
		t.Error(err)
		return
	}

	http.ListenAndServe("192.168.30.33:1202", frame)
}
