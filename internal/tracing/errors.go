/*
   Copyright 2023 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package tracing

import (
	"fmt"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
)

// DebugEnabled reports whether OTel SDK/exporter internals should print
// their diagnostics, controlled by the COMPOSE_OTEL_DEBUG environment
// variable. It defaults to false so tracing plumbing never leaks into
// ordinary CLI output.
func DebugEnabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("COMPOSE_OTEL_DEBUG"))
	return enabled
}

// errorHandler is the otel.ErrorHandler installed for the CLI: it discards
// errors unless DebugEnabled reports true, in which case it prints them to
// stderr.
type errorHandler struct{}

func (errorHandler) Handle(err error) {
	if DebugEnabled() {
		fmt.Fprintln(os.Stderr, "otel:", err)
	}
}

var _ otel.ErrorHandler = errorHandler{}
