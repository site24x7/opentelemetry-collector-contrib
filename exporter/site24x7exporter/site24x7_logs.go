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
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/collector/model/pdata"
)

func (e *site24x7exporter) CreateLogItem(logrecord pdata.LogRecord, resourceAttr map[string]interface{}) TelemetryLog {
	startTime := (logrecord.Timestamp().AsTime().UnixNano() / int64(time.Millisecond))
	tlogBodyType := logrecord.Body().Type()
	tlogMsg := logrecord.Name()
	tlogTraceId := logrecord.TraceID().HexString()
	tlogSpanId := logrecord.SpanID().HexString()
	var tlogInstanceName string
	switch tlogBodyType {
	case pdata.AttributeValueTypeString:
		tlogMsg = logrecord.Body().AsString()
		
	case pdata.AttributeValueTypeMap:
		tlogKvList := logrecord.Body().MapVal().AsRaw()
		// if kvlist gives "msg":"<logmsg>"
		if attrVal, found := tlogKvList["msg"]; found {
			//tLogValue := v1.KeyValueList(tlogKvList).GetValues()
			tlogMsg = attrVal.(string)
		} else {
			tlogMsg = logrecord.Body().AsString()
		}

		if tlogKvSpanId, found := tlogKvList["span_id"]; found {
			tlogSpanId = tlogKvSpanId.(string)
		}

		if tlogKvTraceId, found := tlogKvList["trace_id"]; found {
			tlogTraceId = tlogKvTraceId.(string)
		}
	}

	if resourceInstance, found := resourceAttr["instance"]; found {
		tlogInstanceName = resourceInstance.(string)
	} else {
		tlogInstanceName = "localhost"
	}

	tlog := TelemetryLog{
		Timestamp:          startTime,
		S247UID:            "otel-s247exporter",
		LogLevel:           logrecord.SeverityText(),
		TraceId:            tlogTraceId,
		SpanId:             tlogSpanId,
		TraceFlag:          logrecord.Flags(),
		Instance:			tlogInstanceName,
		ResourceAttributes: resourceAttr,
		LogAttributes:      logrecord.Attributes().AsRaw(),
		Name:               logrecord.Name(),
		Message:            tlogMsg,
	}
	return tlog
}

func (e *site24x7exporter) ConsumeLogs(_ context.Context, ld pdata.Logs) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	logCount := ld.LogRecordCount()
	logList := make([]TelemetryLog, 0, logCount)

	for i := 0; i < ld.ResourceLogs().Len(); i++ {
		rlogs := ld.ResourceLogs().At(i)
		resource := rlogs.Resource()
		resourceAttr := resource.Attributes().AsRaw()
		instLogs := rlogs.ScopeLogs()
		for j := 0; j < instLogs.Len(); j++ {
			ilogs := instLogs.At(j)
			ilogItems := ilogs.LogRecords()
			for k := 0; k < ilogItems.Len(); k++ {
				rawLogitem := ilogItems.At(k)
				logItem := e.CreateLogItem(rawLogitem, resourceAttr)
				logList = append(logList, logItem)
			}
		}
	}

	buf, err := json.Marshal(logList)
	if err != nil {
		errstr := err.Error()
		fmt.Println("Error in converting telemetry logs: ", errstr)
		
		return err
	}
	
	client := http.Client{}

	var gzbuf bytes.Buffer
	g := gzip.NewWriter(&gzbuf)
	g.Write(buf)
	g.Close()
	req , err := http.NewRequest("POST", e.url, &gzbuf)
	if err != nil {
		//Handle Error
		fmt.Println("Error initializing Url: ", err)
		return err
	}
	hostname, err := os.Hostname()
	req.Header = http.Header{
		"apikey": []string{e.apikey},
		"Content-Type": []string{"application/json"},
		"logtype": []string{"s247otellogs"},
		"x-service": []string{"MX"},
		"x-streammode": []string{"1"},
		"log-size": []string{strconv.Itoa(len(buf))},
		"upload-id": []string{uuid.New().String()},
		"agentuid": []string{hostname},
		"Content-Encoding": []string{"gzip"},
		"User-Agent": []string{"site24x7exporter"},
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: e.insecure}
	res , err := client.Do(req)
	if err != nil {
		//Handle Error
		fmt.Println("Error initializing Url: ", err)
		return err
	}
	fmt.Println("Uploaded logs information: ", res.Header)
	return err
}
