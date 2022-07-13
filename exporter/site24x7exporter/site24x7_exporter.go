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

package site24x7exporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/site24x7exporter"

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	//"github.com/open-telemetry/opentelemetry-proto/gen/go/common/v1"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// Marshaler configuration used for marhsaling Protobuf to JSON.
var metricsMarshaler = pmetric.NewJSONMarshaler()

type site24x7exporter struct {
	host		string
	apikey 		string
	insecure 	bool
	dc			string
	mutex 		sync.Mutex
}

func (e *site24x7exporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (e *site24x7exporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	buf, err := metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	return exportMessageAsLine(e, buf)
}

func exportMessageAsLine(e *site24x7exporter, buf []byte) error {
	// Ensure only one write operation happens at a time.
	e.mutex.Lock()
	defer e.mutex.Unlock()
	var urlBuf bytes.Buffer
	responseBody := bytes.NewBuffer(buf)
	//fmt.Fprint(&urlBuf, e.url, "?license.key=",e.apikey);
	
	fmt.Fprint(&urlBuf, "https://", e.host, "/otel/metrics?license.key=",e.apikey);

	resp, err := http.Post(urlBuf.String(), "application/json", responseBody)
	fmt.Println("Posting telemetry data to url. ")
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("Metrics data exported ", body)

	return nil
}

func (e *site24x7exporter) Start(context.Context, component.Host) error {
	// Todo: Send arh/otel/connect and check for response. 
	var responseBody bytes.Buffer
	connectUrl := getDCConnectUrl(e.dc, e.host, e.apikey)
	
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: e.insecure}
	resp, err := http.Post(connectUrl, "application/json", &responseBody)
	if err != nil {
		fmt.Println("Error in posting data to url: ", err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response: ", err)
		return err
	}
	fmt.Println("Response data: ", body)
	return err
}

// Shutdown stops the exporter and is invoked during shutdown.
func (e *site24x7exporter) Shutdown(context.Context) error {
	return nil
}
