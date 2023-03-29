module sf-ui

go 1.19

require golang.org/x/net v0.6.0

require (
	github.com/creack/pty v1.1.18
	github.com/koding/websocketproxy v0.0.0-20181220232114-7ed82d81a28c
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/gorilla/websocket v1.5.0 // indirect

replace github.com/koding/websocketproxy => github.com/messede-degod/websocketproxy v0.0.0-20230329122220-e7e6605bf195

replace github.com/gorilla/websocket => github.com/messede-degod/websocket v0.0.0-20230329085455-f7bedd30414c
