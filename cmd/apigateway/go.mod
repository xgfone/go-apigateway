module github.com/xgfone/go-apigateway/cmd/apigateway

require (
	github.com/xgfone/go-apigateway v0.0.0
	github.com/xgfone/go-atexit v0.10.0
	github.com/xgfone/go-defaults v0.12.0
	github.com/xgfone/go-loadbalancer v0.0.0
)

require (
	github.com/xgfone/go-atomicvalue v0.2.0 // indirect
	github.com/xgfone/go-binder v0.5.0 // indirect
	github.com/xgfone/go-cast v0.8.1 // indirect
	github.com/xgfone/go-checker v0.3.0 // indirect
	github.com/xgfone/go-structs v0.2.0 // indirect
)

replace github.com/xgfone/go-apigateway => ../..

replace github.com/xgfone/go-loadbalancer => github.com/xgfone/go-loadbalancer v0.0.0-20230903055907-6ed28118b546

go 1.21
