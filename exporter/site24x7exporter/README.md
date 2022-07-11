# Site24x7 Exporter

[Site24x7](https://www.site24x7.com/) is an all-in-one monitoring solution, and supports Opentelemetry instrumented applications/services with this exporter. 

This exporter supports sending traces, logs to Site24x7. 
> Please review the Collector's [security
> documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/security.md),
> which contains recommendations on securing sensitive information such as the
> apikey required by this exporter.

## Getting Started

The following settings are required:

- `dc` (no default): datacentre id, can be one of {us,eu,cn,au,in}.
- `apikey` (no default): your Site24x7 License key. 

Example:
## Basic Example
```yaml
exporters:
  site24x7:
    dc: us
    apikey: site24x7_license_key

```

## Configuration

The following configurations are supported:
* `apikey` : Your [Site24x7 License key](https://www.site24x7.com/help/admin/developer/device-key.html)
* `dc` : Your DataCentre region. Supported regions are {us, eu, cn, au, in}. You can set `dc` to local if you wish to send data to your custom end point. 

Please make sure to include a batch processor in your pipeline. 

Example:
```yaml
processors:
  batch:
    timeout: 10s
    send_batch_size: 1000

exporters:
  site24x7:
    dc: us
    apikey: site24x7_license_key

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [site24x7]
    logs:
      receivers: [filelog,fluentforward, otlp]
      processors: [batch]
      exporters: [site24x7]

```