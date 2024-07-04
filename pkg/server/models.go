// Package server provides methods in the file grpcMethods.go that implements EditorServer in 'api' package interface,
// structures in the file models.go and methods for server launch.
package server

import (
	"dnsService/pkg/api"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

// S implements server structure
type S struct {
	grpcS                     *grpc.Server
	proxyS                    *http.Server
	srv                       *grpcServer
	l                         net.Listener
	c                         config
	LogInfo, LogErr, LogFatal *log.Logger
}

// grpcServer implements grpc server structure
type grpcServer struct {
	api.UnimplementedEditorServer
	db *redis.Client
}

// config for start servers and connect DB
type config struct {
	Server struct {
		Host          string `yaml:"host"`
		Port          string `yaml:"port"`
		HttpProxyPort string `yaml:"http_proxy_port"`
	} `yaml:"server"`
	DB struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"redis"`
}
