// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package errors

import (
	"errors"
	"fmt"
)

// Exit status codes
//
// Where possible, copied from curl: https://ec.haxx.se/usingcurl/usingcurl-returns
const (
	ErrUnsupportedProtocol  = 1
	ErrFailedToInitialize   = 2
	ErrURLMalformed         = 3
	ErrFailedToConnect      = 7
	ErrHTTPPageNotRetrieved = 22
)

type curlErr struct {
	error
	code int
}

func (e *curlErr) Error() string {
	return e.error.Error()
}

func (e *curlErr) Unwrap() error {
	return e.error
}

func (e *curlErr) ExitStatus() int {
	return e.code
}

func WithCode(err error, code int) error {
	return &curlErr{
		error: err,
		code:  code,
	}
}

// New calls errors.New.
func New(text string) error {
	return errors.New(text)
}

func InspectErrorCode(err error) int {
	if err == nil {
		return 0
	}
	var codeErr interface {
		ExitStatus() int
	}
	if errors.As(err, &codeErr) {
		return codeErr.ExitStatus()
	}
	var kivikErr interface {
		StatusCode() int
	}
	if errors.As(err, &kivikErr) {
		return kivikErr.StatusCode()
	}

	return 0
}

// Code returns a new error with an error code. If err is an existing error, it
// is wrapped with the error code. All other values are passed to fmt.Sprint.
func Code(code int, err ...interface{}) error {
	if len(err) == 0 {
		if e, ok := err[0].(error); ok {
			return &curlErr{
				error: e,
				code:  code,
			}
		}
	}
	return &curlErr{
		error: errors.New(fmt.Sprint(err...)),
		code:  code,
	}
}

// Codef wraps the output of fmt.Errorf with a code.
func Codef(code int, format string, args ...interface{}) error {
	return &curlErr{
		error: fmt.Errorf(format, args...),
		code:  code,
	}
}

// As calls errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is calls errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap calls errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
