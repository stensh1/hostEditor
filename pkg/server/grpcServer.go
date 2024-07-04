// Package server provides methods in the file grpcMethods.go that implements EditorServer in 'api' package interface,
// structures in the file models.go and methods for server launch.
package server

import (
	"context"
	"dnsService/pkg/api"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// mustInit initializes the server, downloads and parses the configuration file,
// creates instances of the logger, server and proxy server
func (s *S) mustInit() {
	// new logger instance
	if err := s.logger(); err != nil {
		s.LogErr.Println("grpcServer:mustInit(): Failed to create log file:", err)
	} else {
		s.LogInfo.Println("grpcServer:mustInit(): Logging in file")
	}

	// new grpc Server
	s.grpcS = grpc.NewServer()
	s.srv = &grpcServer{}
	s.LogInfo.Println("grpcServer:mustInit(): GRPC the new server created")

	// parsing conf file
	if err := s.loadConfig(); err != nil {
		s.LogFatal.Panic("grpcServer:mustInit(): Server configuration uploaded status:", err)
	} else {
		s.LogInfo.Println("grpcServer:mustInit(): Server configuration uploaded")
	}

	// new proxy server
	// Creating a proxy server w/ config
	s.proxyS = &http.Server{
		Addr: s.c.Server.Host + ":" + s.c.Server.HttpProxyPort,
	}
	s.LogInfo.Println("grpcServer:mustInit(): Server initialized successfully")
}

// Start starts grpc server, proxy server and connects to database
func (s *S) Start() {
	// create server instance
	s.mustInit()
	// redis connection
	s.dbConnect()
	// grpc start server
	go s.mustGrpcStart()
	// proxy start server
	go s.httpStart()
}

// mustGrpcStart starts grpc server
func (s *S) mustGrpcStart() {
	var err error

	// register new server with implemented methods
	api.RegisterEditorServer(s.grpcS, s.srv)

	// starting listening port
	s.LogInfo.Println("grpcServer:mustGrpcStart(): Listening: ", s.c.Server.Host+":"+s.c.Server.Port)
	s.l, err = net.Listen("tcp", s.c.Server.Host+":"+s.c.Server.Port)
	if err != nil {
		s.LogFatal.Panic("grpcServer:mustGrpcStart(): Failed to start listening: ", s.c.Server.Host+":"+s.c.Server.Port, err)
	}

	// serve listener
	s.LogInfo.Println("grpcServer:mustGrpcStart(): Serving: ", s.c.Server.Host+":"+s.c.Server.Port)
	if err := s.grpcS.Serve(s.l); err != nil {
		s.LogFatal.Panic("grpcServer:mustGrpcStart(): Failed to serve: ", s.c.Server.Host+":"+s.c.Server.Port, err)
	}
}

// httpStart starts proxy server
func (s *S) httpStart() {
	// context to reg endpoint
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// register new server
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := api.RegisterEditorHandlerFromEndpoint(ctx, mux, s.c.Server.Host+":"+s.c.Server.Port, opts)
	if err != nil {
		s.LogErr.Println("grpcServer:httpStart(): Failed to start HTTP gateway:", err)
		return
	}

	// serving
	s.LogInfo.Println("grpcServer:httpStart(): HTTP proxy listening on:", s.proxyS.Addr)
	if err := http.ListenAndServe(s.proxyS.Addr, mux); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.LogInfo.Println("grpcServer:httpStart(): Normal interrupt operation", err)
		} else {
			s.LogFatal.Println("grpcServer:httpStart(): Server failed to listen: ", s.proxyS.Addr, " ", err)
		}
	}
}

// Stop all servers and close db connection
func (s *S) Stop() {
	// New context to stop server
	ctx := context.Background()

	// stop grpc
	s.grpcS.GracefulStop()
	s.LogInfo.Println("grpcServer:Stop(): GRPC server gracefully stopped")

	// stop proxy
	if err := s.proxyS.Shutdown(ctx); err != nil {
		s.LogErr.Println("grpcServer:Stop(): Failed to gracefully stop http proxy:", err)
	} else {
		s.LogInfo.Println("grpcServer:Stop(): HTTP proxy server gracefully stopped")
	}

	// close db
	if err := s.srv.db.Close(); err != nil {
		s.LogErr.Println("grpcServer:Stop(): Failed to close db connection:", err)
	}
}

// logger is a private S method that implements three levels of logging
func (s *S) logger() error {
	// Creating a new log file just for each server launch
	f, err := os.OpenFile("logs/"+fmt.Sprint(time.Now().Date())+" "+
		fmt.Sprint(time.Now().Clock())+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		s.LogInfo = log.New(os.Stdout, "INFO:", log.Ldate|log.Ltime)
		s.LogErr = log.New(os.Stdout, "ERROR:", log.Ldate|log.Ltime)
		s.LogFatal = log.New(os.Stdout, "FATAL ERROR:", log.Ldate|log.Ltime)
		return err
	} else {
		s.LogInfo = log.New(f, "INFO: ", log.Ldate|log.Ltime)
		s.LogErr = log.New(f, "ERROR: ", log.Ldate|log.Ltime)
		s.LogFatal = log.New(f, "FATAL ERROR: ", log.Ldate|log.Ltime)
		return nil
	}
}

// loadConfig is a private S method opens the yaml configuration file
// and parse the data from it into the internal Config structure
func (s *S) loadConfig() error {
	// Initialising config object
	s.c = config{}

	// Reading the configuration .yaml file
	if err := s.c.newConfig("cfg/config.yaml"); err != nil {
		return err
	}

	return nil
}

// newConfig is a config private method that downloads a yaml file and decodes it into a config structure
func (c *config) newConfig(configPath string) error {
	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close() // TODO: handle it

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(c); err != nil {
		return err
	}

	return nil
}

// dbConnect searches for the database password in the environment variables and connects to the database
func (s *S) dbConnect() {
	// search password in env vars
	dbPswd, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		s.LogErr.Println("grpcServer:dbConnect(): Unable to get database password")
		return
	}

	// new db client instance
	s.srv.db = redis.NewClient(&redis.Options{
		Addr:     s.c.DB.Host + ":" + s.c.DB.Port,
		Password: dbPswd,
	})

	s.LogInfo.Println("grpcServer:dbConnect(): Database connection established")
}
