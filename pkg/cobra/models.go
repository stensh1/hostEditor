// Package cobra was partially generated with 'cobra-cli init'.
// It provides methods in the file clientMethods.go, structures in the file models.go,
// and variables for the cobra in the file cobraVars.go to implement the client app.
package cobra

import (
	"dnsService/pkg/api"
	"google.golang.org/grpc"
)

// client
type c struct {
	username string
	token    []byte
	conf     config
	conn     *grpc.ClientConn
	client   api.EditorClient
}

// info about addr and port to connect
type config struct {
	u string // host addr
	p string // port
}
