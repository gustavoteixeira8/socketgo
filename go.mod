module socketgo

go 1.20

replace localdb => ../localdb

require (
	github.com/google/uuid v1.5.0
	github.com/gorilla/websocket v1.5.1
	github.com/sirupsen/logrus v1.9.3
	github.com/gustavoteixeira8/localdb/repository v0.0.0
)

require (
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
