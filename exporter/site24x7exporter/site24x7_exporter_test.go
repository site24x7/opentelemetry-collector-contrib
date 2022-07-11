// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package site24x7exporter

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/model/otlp"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)
type testConfig struct {
	apikey string
}
func testTraceData(t *testing.T, expected []TelemetrySpan, td pdate.Trace, apiKey string, dc string) {
	ctx := context.Background()
	f := NewFactory()
	require.NoError(t, err)
	assert.Equal(t, expected, m.Batches)
}

func newTestTraces() pdata.Traces {
	td := pdata.NewTraces()
	sps := td.ResourceSpans().AppendEmpty().InstrumentationLibrarySpans().AppendEmpty().Spans()
	s1 := sps.AppendEmpty()
	s1.SetName("a")
	s1.SetTraceID(pdata.NewTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}))
	s1.SetSpanID(pdata.NewSpanID([8]byte{0, 0, 0, 0, 0, 0, 0, 1}))
	
	s2 := sps.AppendEmpty()
	s2.SetName("b")
	s2.SetTraceID(pdata.NewTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}))
	s2.SetSpanID(pdata.NewSpanID([8]byte{0, 0, 0, 0, 0, 0, 0, 2}))
	return td
}