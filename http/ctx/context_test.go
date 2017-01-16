package ctx

import (
	"net/http"
	"reflect"
	"testing"
)

func generateRequest(qs string) *http.Request {
	r, err := http.NewRequest("GET", "?"+qs, nil)
	if err != nil {
		panic(err)
	}
	return r
}

func TestContext_QueryIntMultiple(t *testing.T) {
	tests := []struct {
		Request *http.Request
		arg     string
		want    []int
	}{
		{
			Request: generateRequest("t=1"),
			want:    []int{1},
		},
		{
			Request: generateRequest("t=1&t=3"),
			want:    []int{1, 3},
		},
		{
			Request: generateRequest("t=3&t=1"),
			want:    []int{3, 1},
		},
	}
	for i, tt := range tests {
		if tt.arg == "" {
			tt.arg = "t"
		}
		r := &Context{
			Request: tt.Request,
		}
		if got := r.QueryIntMultiple(tt.arg); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%d. Context.QueryIntMultiple() = %v, want %v", i, got, tt.want)
		}
	}
}
