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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
)
/*type testConfig struct {
	apikey string
}*/

func testTraceData(t *testing.T) {
	f := NewFactory()
	cfg := &Config{
		DataCentre: "local",
		Host: "plusinsight.localsite24x7.in",
		APIKEY: "0123456789abcdef",
		Insecure: true,
	}

	se, err := f.CreateTracesExporter(context.Background(), componenttest.NewNopExporterCreateSettings(), cfg)
	assert.NoError(t, err)

	td := testdata.GenerateTracesTwoSpansSameResource()
	assert.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, se.ConsumeTraces(context.Background(), td))
	assert.NoError(t, se.Shutdown(context.Background()))

}