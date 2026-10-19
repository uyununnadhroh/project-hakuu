package response

import (
	"encoding/json"
	"net/http"
)

// Response defines http response for the client.
//
// Format JSON sengaja dibuat sama dengan versi lama supaya frontend tidak
// perlu diubah: bila sukses, body = Data apa adanya; bila error, body =
// {"error": "pesan"} dengan HTTP status yang sesuai.
type Response struct {
	Data       interface{}
	Error      string
	StatusCode int
}

// SetError set the response to return the given error.
// code is http status code, http.StatusInternalServerError is the default value
func (res *Response) SetError(err error, code ...int) {
	if len(code) > 0 {
		res.StatusCode = code[0]
	} else {
		res.StatusCode = http.StatusInternalServerError
	}
	if err != nil {
		res.Error = err.Error()
	}
}

// RenderJSON writes the http response in JSON format to the client
func (res *Response) RenderJSON(w http.ResponseWriter, r *http.Request) {
	if res.StatusCode == 0 {
		res.StatusCode = http.StatusOK
	}

	var body interface{} = res.Data
	if res.Error != "" {
		body = map[string]string{"error": res.Error}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(res.StatusCode)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
