package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

// Error codes
const (
	CodeInvalidToken = iota
	CodeInternalError
	CodeInvalidRequest
	CodeEncodeError
)

// Error model
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func newErr(c int, m string) *Error {
	return &Error{Code: c, Msg: m}
}

func fromErr(c int, err error) *Error {
	return newErr(c, err.Error())
}

// GetResponse creates response for error
func (err *Error) GetResponse() *Response {
	return &Response{Error: err}
}

// NewResponse returns new response
func NewResponse(result interface{}) *Response {
	return &Response{Success: true, Result: result}
}

// Response model
type Response struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// WriteResponse to response writer
func (r *Response) WriteResponse(w http.ResponseWriter) {
	data, err := json.Marshal(r)
	if err != nil {
		fmt.Printf("MarshalResponse: %v\n", spew.Sdump(r.Result))
		fmt.Println("MarshalResponse", err)
		data = []byte(`{"success":false,"error":{"error": 10, "error_msg":"cannot marshal response, see server logs for details"}}`)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
