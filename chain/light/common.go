package light

import "encoding/json"

type Result struct {
	Code int
	Rst  interface{}
}

func pkg(code int, rst interface{}) *[]byte {
	r := Result{code, rst}
	marshal, _ := json.Marshal(r)
	return &marshal
}
