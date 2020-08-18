module github.com/cocotyty/cmdb_driver

go 1.13

require (
	github.com/cocotyty/forceset v1.0.4
	github.com/golang/protobuf v1.4.1
	github.com/zhihu/cmdb v0.0.0-20200615031832-32d25b74d2cd
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/genproto v0.0.0-20200430143042-b979b6f78d84
	google.golang.org/grpc v1.29.1
)

replace github.com/zhihu/cmdb => github.com/cocotyty/cmdb v0.0.0-20200807104248-47c0014cbc10
