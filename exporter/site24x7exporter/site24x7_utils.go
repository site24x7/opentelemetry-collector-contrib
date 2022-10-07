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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	
	"github.com/google/uuid"
	
)

// returns Data centre host names. Nil if not found. 
func getDataCentreHost(
	dc string,
) (string) {
	switch dc {
	case "us":
		return "plusinsight.site24x7.com"
	case "eu": 
		return "plusinsight.site24x7.eu"
	case "cn":
		return "plusinsight.site24x7.cn"
	case "au": 
		return "plusinsight.site24x7.net.au"
	case "in":
		return "plusinsight.site24x7.in"
	case "jp": 
		return "plusinsight.site24x7.jp"
	}
	return ""
}

func getDCConnectURL(
	dc string,
	host string,
	apikey string,
) (string) {
	var urlBuf bytes.Buffer
	dchost := getDataCentreHost(dc)
	if dchost == "" {
		dchost = host
	}
	fmt.Fprint(&urlBuf, "https://", dchost, "/otel/connect?license.key=", apikey)
	return urlBuf.String()
}

func getTraceURL(
	dc string, 
	host string,
	apikey string,
) (string) {
	var urlBuf bytes.Buffer
	dchost := getDataCentreHost(dc)
	if dchost == "" {
		dchost = host
	}
	fmt.Fprint(&urlBuf, "https://", dchost, "/otel/trace?license.key=", apikey)
	return urlBuf.String()
}

func getLogsURL(
	dc string, 
	host string,
	apikey string,
) (string) {
	var urlBuf bytes.Buffer
	dchost := getDataCentreHost(dc)
	if dchost == "" {
		dchost = host
	}
	fmt.Fprint(&urlBuf, "https://", dchost, "/otel/logs?license.key=", apikey)
	return urlBuf.String()
}

func getMetricsUrl(
	dc string, 
	host string,
	apikey string,
) (string) {
	var urlBuf bytes.Buffer
	dchost := getDataCentreHost(dc)
	if dchost == "" {
		dchost = host
	}
	fmt.Fprint(&urlBuf, "https://", dchost, "/otel/metrics?license.key=", apikey)
	return urlBuf.String()
}

// Sends traces in applogs format. Requires logtype to be initialized in connect
func (e *site24x7exporter) SendOtelTraces(spanList []TelemetrySpan) error {
	// Convert the spanlist to Json. 	
	buf, err := json.Marshal(spanList)
	if err != nil {
		fmt.Println("Error in converting traces data ", err)
		return err
	}
	// Applogs server requires gzip compression. 
	var gzbuf bytes.Buffer
	g := gzip.NewWriter(&gzbuf)
	g.Write(buf)
	g.Close()

	client := http.Client{}
	traceURL := getTraceURL(e.dc,e.host,e.apikey)

	req , err := http.NewRequest("POST", traceURL, &gzbuf)
	if err != nil {
		//Handle Error
		fmt.Println("Error initializing Url: ", err)
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		//Handle Error
		fmt.Println("Error getting hostname: ", err)
		return err
	}

	req.Header = http.Header{
		"apikey": []string{e.apikey},
		"Content-Type": []string{"application/json"},
		"logtype": []string{"s247apmopentelemetrytracing"},
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
		fmt.Println("Error in posting Telemetry traces to the server: ", err)
		return err
	}
	fmt.Println("Uploaded traces information: ", res.Header)
	return err
}

// Sends logs in applogs format. Requires logtype to be initialized in connect
func (e *site24x7exporter) SendOtelLogs(logRecords []TelemetryLog) error {
	// All data must be in gzipped json format. 
	buf, err := json.Marshal(logRecords)
	if err != nil {
		errstr := err.Error()
		fmt.Println("Error in converting telemetry logs: ", errstr)
		
		return err
	}

	var gzbuf bytes.Buffer
	g := gzip.NewWriter(&gzbuf)
	g.Write(buf)
	g.Close()
	client := http.Client{}
	logsURL := getLogsURL(e.dc,e.host,e.apikey)

	req , err := http.NewRequest("POST", logsURL, &gzbuf)
	if err != nil {
		//Handle Error
		fmt.Println("Error initializing Url: ", err)
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		//Handle Error
		fmt.Println("Error getting hostname: ", err)
		return err
	}
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
		fmt.Println("Error in posting telemetry logs to the server: ", err)
		return err
	}
	fmt.Println("Uploaded logs information: ", res.Header)
	
	return err
}
// Deprecated. Sends traces in json/file format.
func (e *site24x7exporter) SendRawTraces(spanList []TelemetrySpan) error {
	// Convert the spanlist to Json. 	
	buf, err := json.Marshal(spanList)
	if err != nil {
		fmt.Println("Error in converting traces data ", err)
		return err
	}
	traceURL := getTraceURL(e.dc,e.host,e.apikey)
	responseBody := bytes.NewBuffer(buf)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: e.insecure}
	resp, err := http.Post(traceURL, "application/json", responseBody)
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
	fmt.Println("Uploaded Traces Information! Response from server: ", body)
	return err
}
// Deprecated. Sends logs in json/file format. 
func (e *site24x7exporter) SendRawLogs(logRecords []TelemetryLog) error {
	// Convert Log data to json format. 
	buf, err := json.Marshal(logRecords)
	if err != nil {
		errstr := err.Error()
		fmt.Println("Error in converting telemetry logs: ", errstr)
		return err
	}
	logsURL := getLogsURL(e.dc,e.host,e.apikey)
	responseBody := bytes.NewBuffer(buf)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: e.insecure}
	resp, err := http.Post(logsURL, "application/json", responseBody)
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
	fmt.Println("Uploaded Logs Information! Response from server: ", body)
	return err
}
