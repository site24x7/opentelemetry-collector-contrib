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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	//"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/sharedcomponent"
)

const (
	// The value of "type" key in configuration.
	typeStr = "site24x7"
)

// NewFactory creates a factory for OTLP exporter.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesExporter(createTracesExporter),
		component.WithLogsExporter(createLogsExporter))
		//component.WithMetricsExporter(createMetricsExporter), 
}

func createDefaultConfig() config.Exporter {
	return &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),
		TimeoutSettings:  exporterhelper.NewDefaultTimeoutSettings(),
		RetrySettings:    exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:    exporterhelper.NewDefaultQueueSettings(),
		APIKEY: "ab_123",
		DataCentre: "local",
		Host: "plusinsight.localsite24x7.in",
		Insecure: true,
	}
}

func getDCHost(
	cfg config.Exporter,
) (string) {
	dchost := getDataCentreHost(cfg.(*Config).DataCentre)
	if dchost == "" {
		dchost = cfg.(*Config).Host
	}

	return dchost
}

func getDCUrlSecurity(
	cfg config.Exporter,
) (bool) {
	switch cfg.(*Config).DataCentre {
	case "us":
	case "eu":
	case "cn":
	case "au":
	case "in":
	case "jp":
		return false
	}
	return cfg.(*Config).Insecure
}

func createTracesExporter(
	_ context.Context,
	set component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.TracesExporter, error) {
	s247exp :=  &site24x7exporter{
			dc: cfg.(*Config).DataCentre,
			host:  getDCHost(cfg),
			apikey: cfg.(*Config).APIKEY,
			insecure: getDCUrlSecurity(cfg),
		}

	return exporterhelper.NewTracesExporter(
		cfg,
		set,
		s247exp.ConsumeTraces,
		exporterhelper.WithStart(s247exp.Start),
		exporterhelper.WithShutdown(s247exp.Shutdown),
	)
}

func createMetricsExporter(
	_ context.Context,
	set component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.MetricsExporter, error) {
	s247exp :=  &site24x7exporter{
		dc: cfg.(*Config).DataCentre,
		host:  getDCHost(cfg),
		apikey: cfg.(*Config).APIKEY,
		insecure: getDCUrlSecurity(cfg),
	}
	return exporterhelper.NewMetricsExporter(
		cfg,
		set,
		s247exp.ConsumeMetrics,
		exporterhelper.WithStart(s247exp.Start),
		exporterhelper.WithShutdown(s247exp.Shutdown),
	)
}

func createLogsExporter(
	_ context.Context,
	set component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.LogsExporter, error) {
	s247exp :=  &site24x7exporter{
		dc: cfg.(*Config).DataCentre,
		host:  getDCHost(cfg),
		apikey: cfg.(*Config).APIKEY,
		insecure: getDCUrlSecurity(cfg),
	}
	return exporterhelper.NewLogsExporter(
		cfg,
		set,
		s247exp.ConsumeLogs,
		exporterhelper.WithStart(s247exp.Start),
		exporterhelper.WithShutdown(s247exp.Shutdown),
	)
}
