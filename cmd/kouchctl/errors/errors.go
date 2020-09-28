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

import "errors"

// Exit status codes
//
// Where possible, copied from curl: https://ec.haxx.se/usingcurl/usingcurl-returns
const (
	ErrUnsupportedProtocol = 1
	ErrFailedToInitialize  = 2
	ErrURLMalformed        = 3
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

func (e *curlErr) ErrCode() int {
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
	var codeErr interface {
		ErrCode() int
	}
	if errors.As(err, &codeErr) {
		return codeErr.ErrCode()
	}
	return 0
}
