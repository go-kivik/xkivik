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
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
)

func TestInspectErrorCode(t *testing.T) {
	type tt struct {
		err  error
		want int
	}

	tests := testy.NewTable()
	tests.Add("standard", tt{
		err:  errors.New("foo"),
		want: 0,
	})
	tests.Add("codeErr", tt{
		err:  WithCode(errors.New("foo"), 123),
		want: 123,
	})
	tests.Add("wrapped", tt{
		err:  fmt.Errorf("%w", WithCode(errors.New("foo"), 123)),
		want: 123,
	})
	tests.Add("kivik 404", tt{
		err:  &kivik.Error{HTTPStatus: 400},
		want: 14,
	})
	tests.Add("kivik internal server error", tt{
		err:  &kivik.Error{HTTPStatus: 500},
		want: ErrInternalServerError,
	})
	tests.Add("kivik 501", tt{
		err:  &kivik.Error{HTTPStatus: 501},
		want: ErrUnknown,
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got := InspectErrorCode(tt.err)
		if got != tt.want {
			t.Errorf("want %d, got %d", tt.want, got)
		}
	})
}
