module github.com/go-anyway/framework-errors

go 1.25.4

require (
	github.com/go-anyway/framework-log v1.0.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/go-anyway/framework-log => ../log
