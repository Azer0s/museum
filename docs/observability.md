# Observability

## Tracing

Tracing is done via opentelemetry using gRPC to a Jaeger instance. The `Tracer` is always the application layer (e.g. `etcd persistence`, `nats eventing`, etc.) and the start tag should always correspond to the method name. Tracing can be enabled in prod, but be aware that there will be a lot of data collected.

## Logging

Logs are emitted to stdout in JSON. A log entry always includes the log level, the timestamp and a message. Additional fields can provide extra context (like `request_id`, `exhibitId`, etc.).

Logging in mūsēum is done as frequently as possible using the correct levels. In prod using the `WARN` level as a default is recommended.
