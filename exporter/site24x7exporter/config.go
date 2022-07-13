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
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for file exporter.
type Config struct {
	config.ExporterSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	// TimeoutSettings is the total amount of time spent attempting a request,
	// including retries, before abandoning and dropping data. Default is 5
	// seconds.
	TimeoutSettings exporterhelper.TimeoutSettings `mapstructure:",squash"`

	// RetrySettings defines configuration for retrying batches in case of export failure.
	// The current supported strategy is exponential backoff.
	RetrySettings exporterhelper.RetrySettings `mapstructure:"retry"`

	QueueSettings exporterhelper.QueueSettings `mapstructure:"queue"`

	// Data centre to push data: Supported config - us,eu,cn,au,in,local
	DataCentre string `mapstructure:"dc"`
	// host to which the opentelemetry data is pushed to. Only applicable when dc is set to local
	Host string `mapstructure:"host"`
	// API Key of site24x7.
	APIKEY string `mapstructure:"apikey"`
	// Is https to host insecure? Only applicable when dc is set to local
	Insecure bool `mapstructure:"insecure"`
}

var _ config.Exporter = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {

	switch cfg.DataCentre {
	case "us":
	case "eu":
	case "cn":
	case "au":
	case "in":
	case "jp":
		if cfg.Host != "" {
			return errors.New("host must not be given for non-local dc")
		}
		fmt.Println("Data centre validation passed")
	case "local":
		if cfg.Host == "" {
			return errors.New("Host must be non-empty")
		}
	}

	if cfg.APIKEY == "" {
		return errors.New("API Key must be non-empty")
	}

	return nil
}
