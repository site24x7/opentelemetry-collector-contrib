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
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func convertLogToMap(lr plog.LogRecord) map[string]interface{} {
	out := map[string]interface{}{}

	if lr.Body().Type() == pcommon.ValueTypeString {
		out["log"] = lr.Body().StringVal()
	}

	lr.Attributes().Range(func(k string, v pcommon.Value) bool {
		switch v.Type() {
		case pcommon.ValueTypeString:
			out[k] = v.StringVal()
		case pcommon.ValueTypeInt:
			out[k] = strconv.FormatInt(v.IntVal(), 10)
		case pcommon.ValueTypeDouble:
			out[k] = strconv.FormatFloat(v.DoubleVal(), 'f', -1, 64)
		case pcommon.ValueTypeBool:
			out[k] = strconv.FormatBool(v.BoolVal())
		default:
			panic("missing case")
		}
		return true
	})

	return out
}

func (e *site24x7exporter) CreateLogItem(logrecord plog.LogRecord, resourceAttr map[string]interface{}) TelemetryLog {
	startTime := (logrecord.Timestamp().AsTime().UnixNano() / int64(time.Millisecond))
	//tlogBodyType := logrecord.Body().Type()
	tlogMsg := logrecord.Body().AsString()
	tlogTraceID := logrecord.TraceID().HexString()
	tlogSpanID := logrecord.SpanID().HexString()
	tlogFlags := logrecord.Flags()
	var droppedattr uint32

	var tlogInstanceName, tlogName string

	tlogAttr := convertLogToMap(logrecord)
	droppedattr = logrecord.DroppedAttributesCount()

	if attrVal, found := tlogAttr["msg"]; found {
		tlogMsg = attrVal.(string)
		delete(tlogAttr, "msg")
		droppedattr++
	}
	if tlogKvSpanID, found := tlogAttr["span_id"]; found {
		tlogSpanID = tlogKvSpanID.(string)
		delete(tlogAttr, "span_id")
		droppedattr++
	}

	if tlogKvTraceID, found := tlogAttr["trace_id"]; found {
		tlogTraceID = tlogKvTraceID.(string)
		delete(tlogAttr, "trace_id")
		droppedattr++
	}

	if tlogKvTraceFlags, found := tlogAttr["trace_flags"]; found {
		u64, err := strconv.ParseUint(tlogKvTraceFlags.(string),10,32)
		if err != nil {
			tlogFlags =  uint32(u64)
		} else {
			tlogFlags = 0
		}
		delete(tlogAttr, "trace_flags")
		droppedattr++
	}

	if tlogKvFileName, found := tlogAttr["log.file.name"]; found {
		tlogName = tlogKvFileName.(string)
		delete(tlogAttr, "log.file.name")
		delete(tlogAttr, "log")
		droppedattr++
	}

	if resourceInstance, found := resourceAttr["instance"]; found {
		tlogInstanceName = resourceInstance.(string)
	} else {
		hostname, err := os.Hostname()
		if err != nil {
			tlogInstanceName = "localhost"
		} else {
			tlogInstanceName = hostname
		}
	}

	tlog := TelemetryLog{
		Timestamp:          	startTime,
		S247UID:           		"otel-s247exporter",
		LogLevel:          		logrecord.SeverityText(),
		TraceID:           		tlogTraceID,
		SpanID:            		tlogSpanID,
		TraceFlag:         		tlogFlags,
		Instance:				tlogInstanceName,
		ResourceAttributes:		resourceAttr,
		LogAttributes:     		tlogAttr,
		Name:              		tlogName,
		Message:           		tlogMsg,
		DroppedAttributesCount: droppedattr,
	}
	return tlog
}

func (e *site24x7exporter) ConsumeLogs(_ context.Context, ld plog.Logs) error {
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

	err := e.SendOtelLogs(logList)
	//err := e.SendRawLogs(logList)

	if err != nil {
		fmt.Println("Error in exporting telemetry logs: ", err)
	}
	
	return err
}
