package web

import "encoding/json"

// jsonUnmarshal is a tiny indirection so the test in this package can
// mock it. The interface is identical to encoding/json.Unmarshal.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
