module github.com/theMomax/notypo-backend

require (
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v1.4.0
	github.com/smartystreets/goconvey v0.0.0-20190306220146-200a235640ff // indirect
	github.com/stretchr/testify v1.3.0
	github.com/urfave/cli v1.20.0
	gopkg.in/ini.v1 v1.42.0
)

replace github.com/urfave/cli v1.20.0 => ../cli
