package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/carlmjohnson/requests"
)

type StatusError struct {
	Code  int
	Inner error
}

func (e StatusError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("unexpected status: %d: %v", e.Code, e.Inner)
	}
	return fmt.Sprintf("unexpected status: %d:", e.Code)
}

func (e StatusError) Unwrap() error {
	return e.Inner
}

// ErrorJSON validates the response has an acceptable status
// code and if it's bad, attempts to marshal the JSON
// into the error object provided.
func ErrorJSON(v error, acceptStatuses ...int) requests.ResponseHandler {
	return func(res *http.Response) error {
		for _, code := range acceptStatuses {
			if res.StatusCode == code {
				return nil
			}
		}

		if res.Body == nil {
			return StatusError{Code: res.StatusCode} //fmt.Errorf("%w: unexpected status: %d", (*ResponseError)(res), res.StatusCode)
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return StatusError{Code: res.StatusCode, Inner: err}
		}

		if err = json.Unmarshal(data, &v); err != nil {
			return StatusError{Code: res.StatusCode, Inner: err}
		}

		return StatusError{Code: res.StatusCode, Inner: v}
	}
}
