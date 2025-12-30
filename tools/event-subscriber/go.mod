module github.com/minwook/battery-optimization/tools/event-subscriber

go 1.23.2

replace github.com/minwook/battery-optimization/pkg/events => ../../pkg/events

require github.com/minwook/battery-optimization/pkg/events v0.0.0-00010101000000-000000000000

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)
