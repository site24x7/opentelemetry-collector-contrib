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
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-version"

	"go.opentelemetry.io/collector/pdata/ptrace"

)

func (e *site24x7exporter) CreateTelemetrySpan(span ptrace.Span,
	resourceAttr map[string]interface{},
	serviceName string,
	instLibrary string,
	instLibraryVersion string,
	telSDKLang string,
	telSDKName string,
	rootSpanID string,) TelemetrySpan {

	spanAttr := span.Attributes().AsRaw()
	startTime := (span.StartTimestamp().AsTime().UnixNano()) // int64(time.Millisecond))
	endTime := (span.EndTimestamp().AsTime().UnixNano())     // int64(time.Millisecond))
	spanEvts := span.Events()
	telEvents := make([]TelemetrySpanEvent, 0, spanEvts.Len())
	exceptionMessages := make([]string, 0, spanEvts.Len())
	exceptionStackTraces := make([]string, 0, spanEvts.Len())
	exceptionTypes := make([]string, 0, spanEvts.Len())
	for i := 0; i < spanEvts.Len(); i++ {
		spanEvt := spanEvts.At(i)
		telEvt := TelemetrySpanEvent{
			Timestamp:       (spanEvt.Timestamp().AsTime().UnixNano() / int64(time.Millisecond)),
			Name:            spanEvt.Name(),
			EventAttributes: spanEvt.Attributes().AsRaw(),
		}
		if telEvt.EventAttributes != nil {

			if exMsg, found := telEvt.EventAttributes["exception.message"]; found {
				exceptionMessages = append(exceptionMessages, exMsg.(string))
			}

			if exST, found := telEvt.EventAttributes["exception.stacktrace"]; found {
				exceptionStackTraces = append(exceptionStackTraces, exST.(string))
			}

			if exType, found := telEvt.EventAttributes["exception.type"]; found {
				exceptionTypes = append(exceptionTypes, exType.(string))
			}
		}
		telEvents = append(telEvents, telEvt)
	}
	spanLinks := span.Links()
	telLinks := make([]TelemetrySpanLink, 0, spanLinks.Len())
	for i := 0; i < spanLinks.Len(); i++ {
		spanLink := spanLinks.At(i)
		telLink := TelemetrySpanLink{
			LinkSpanID:  spanLink.SpanID().HexString(),
			LinkTraceID: spanLink.TraceID().HexString(),
		}
		telLinks = append(telLinks, telLink)
	}
	spanState := span.Status()
	spanStatus := spanState.Code().String()
	hasError := false
	switch spanStatus {
	case "STATUS_CODE_ERROR":
		hasError = true
	}

	spanKind := "UNSPECIFIED"
	switch span.Kind() {
	case ptrace.SpanKindInternal:
		spanKind = "INTERNAL"
	case ptrace.SpanKindServer:
		spanKind = "SERVER"
	case ptrace.SpanKindClient:
		spanKind = "CLIENT"
	case ptrace.SpanKindProducer:
		spanKind = "PRODUCER"
	case ptrace.SpanKindConsumer:
		spanKind = "CONSUMER"
	}

	// Host attributes
	var hostIP, hostName, threadname string
	var hostPort, threadid int64
	if attrval, found := spanAttr["net.peer.ip"]; found {
		hostIP = attrval.(string)
	}
	if attrval, found := spanAttr["net.peer.name"]; found {
		hostName = attrval.(string)
	}
	if attrval, found := spanAttr["net.peer.port"]; found {
		hostPort = attrval.(int64)
	}
	// Thread attributes
	if attrval, found := spanAttr["thread.id"]; found {
		threadid = attrval.(int64)
	}
	if attrval, found := spanAttr["thread.name"]; found {
		threadname = attrval.(string)
	}
	// DB attributes
	var dbsystem, dbstmt, dbname, dbconnstr string
	if attrval, found := spanAttr["db.system"]; found {
		dbsystem = attrval.(string)
	}
	if attrval, found := spanAttr["db.statement"]; found {
		dbstmt = attrval.(string)
	}
	if attrval, found := spanAttr["db.name"]; found {
		dbname = attrval.(string)
	}
	if attrval, found := spanAttr["db.connection_string"]; found {
		dbconnstr = attrval.(string)
	}
	// Http attributes
	var httpurl, httpmethod string
	var httpstatus int64
	v1,err := version.NewVersion(instLibraryVersion)
	if err != nil {
		fmt.Println("Error in reading versino: ", err)
	}
	v2,err := version.NewVersion("1.7.0")
	if err != nil {
		fmt.Println("Error in reading version: ", err)
	}
	if v1.LessThan(v2) {
		// prior to 1.7, transaction names were coming under http.url
		if attrval, found := spanAttr["http.url"]; found { 
			httpurl = attrval.(string)
		}
	} else {
		if attrval, found := spanAttr["http.target"]; found {
			httpurl = attrval.(string)
		}
	}
	if attrval, found := spanAttr["http.method"]; found {
		httpmethod = attrval.(string)
	}
	if attrval, found := spanAttr["http.status_code"]; found {
		httpstatus = attrval.(int64)
	}

	isRoot := span.ParentSpanID().IsEmpty()
	telemetryParams := make([]TelemetryCustomParam, 0, len(spanAttr))
	for k, v := range spanAttr {
		telemetrycustomParam := TelemetryCustomParam{
			Key:   k,
			Value: v,
		}
		telemetryParams = append(telemetryParams, telemetrycustomParam)
	}

	SpanID := span.SpanID().HexString()
	TraceID := span.TraceID().HexString()
	parentSpanID := span.ParentSpanID().HexString()
	spanName := span.Name()
	
	startTimeMs := (startTime / int64(time.Millisecond))

	tspan := TelemetrySpan{
		Timestamp:          	startTimeMs,
		S247UID:	            "otel-s247exporter",
		SpanID:                 SpanID,
		TraceID:                TraceID,
		ParentSpanID:           parentSpanID,
		RootSpanID: 			rootSpanID,
		Name:                   spanName,
		Kind:                   spanKind,
		StartTime:              startTime,
		EndTime:                endTime,
		Duration:               float64(endTime-startTime) / float64(time.Millisecond),
		ServiceName:            serviceName,
		resourceAttributes:     resourceAttr,
		spanAttributes:         spanAttr,
		traceState:             string(span.TraceState()),
		spanEvents:             telEvents,
		spanLinks:              telLinks,
		statusCode:             spanState.Code().String(),
		statusMsg:              spanState.Message(),
		droppedAttributesCount: span.DroppedAttributesCount(),
		droppedLinksCount:      span.DroppedLinksCount(),
		droppedEventsCount:     span.DroppedEventsCount(),

		ExceptionMessage:    exceptionMessages,
		ExceptionStackTrace: exceptionStackTraces,
		ExceptionType:       exceptionTypes,

		InstrumentationLibrary:        instLibrary,
		InstrumentationLibraryVersion: instLibraryVersion,
		TelemetrySDKLanguage:          telSDKLang,
		TelemetrySDKName:              telSDKName,

		HostIP:   hostIP,
		HostName: hostName,
		HostPort: hostPort,

		ThreadID:   threadid,
		ThreadName: threadname,

		DbSystem:    dbsystem,
		DbStatement: dbstmt,
		DbName:      dbname,
		DbConnStr:   dbconnstr,

		HTTPURL:        httpurl,
		HTTPMethod:     httpmethod,
		HTTPStatusCode: httpstatus,

		IsRoot:   isRoot,
		HasError: hasError,

		CustomParams: telemetryParams,
	}
	return tspan
}

func (e *site24x7exporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {

	e.mutex.Lock()
	defer e.mutex.Unlock()

	apiKey :=  ctx.Value("api-key")
	if apiKey == nil {
		fmt.Println("Error reading context header api-key. ")
		return errors.New("api-key header not sent")
	}
	e.apikey = apiKey.(string)
	resourcespans := td.ResourceSpans()
	spanCount := td.SpanCount()
	rootSpanList := make(map[string]string)

	spanList := make([]TelemetrySpan, 0, spanCount)
	
	for i := 0; i < resourcespans.Len(); i++ {
		rspans := resourcespans.At(i)
		instSpans := rspans.ScopeSpans()
		// processing root id before sending in arh
		for j := 0; j < instSpans.Len(); j++ {
			ispans := instSpans.At(j)
			ispanItems := ispans.Spans()

			for k := 0; k < ispanItems.Len(); k++ {
				span := ispanItems.At(k)
				var rootSpanID string
				var TraceID string
				if span.ParentSpanID().IsEmpty() {
					TraceID = span.TraceID().HexString()
					rootSpanID = span.Name()
					rootSpanList[TraceID] = rootSpanID
				} 
			}
		}
	}

	for i := 0; i < resourcespans.Len(); i++ {
		rspans := resourcespans.At(i)
		resource := rspans.Resource()
		resourceAttr := resource.Attributes().AsRaw()

		var serviceName, telSDKLang, telSDKName string
		if val, found := resourceAttr["service.name"]; found {
			serviceName = val.(string)
		}
		if val, found := resourceAttr["telemetry.sdk.name"]; found {
			telSDKName = val.(string)
		}
		if val, found := resourceAttr["telemetry.sdk.language"]; found {
			telSDKLang = val.(string)
		}

		instSpans := rspans.ScopeSpans()
		
		for j := 0; j < instSpans.Len(); j++ {
			ispans := instSpans.At(j)
			instLibName := ispans.Scope().Name()
			instLibVer := ispans.Scope().Version()
			ispanItems := ispans.Spans()

			for k := 0; k < ispanItems.Len(); k++ {
				span := ispanItems.At(k)
				
				rootSpanID := rootSpanList[span.TraceID().HexString()]
				
				s247span := e.CreateTelemetrySpan(span, resourceAttr,
					serviceName,
					instLibName, instLibVer,
					telSDKLang, telSDKName, rootSpanID)
				spanList = append(spanList, s247span)
			}
		}
	}
	//err := e.SendOtelTraces(spanList)
	err := e.SendRawTraces(spanList)

	if err != nil {
		fmt.Println("Error in exporting traces", err)
	}

	return err
}
