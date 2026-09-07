module github.com/Muxcore-Media/media-dlna

go 1.25.0

require (
	github.com/Muxcore-Media/core/pkg/contracts v0.5.8
	github.com/Muxcore-Media/core/sdk/go/client v0.5.8
	github.com/Muxcore-Media/core/sdk/go/module v0.5.8
	github.com/anacrolix/dms v1.8.0
	github.com/anacrolix/ffprobe v1.1.0
	google.golang.org/grpc v1.72.2
)

require (
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)

replace github.com/Muxcore-Media/core/pkg/contracts => ./stub/core/pkg/contracts

replace github.com/Muxcore-Media/core/sdk/go/client => ./stub/core/sdk/go/client

replace github.com/Muxcore-Media/core/sdk/go/module => ./stub/core/sdk/go/module
