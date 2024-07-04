// Package server provides methods in the file grpcMethods.go that implements EditorServer in 'api' package interface,
// structures in the file models.go and methods for server launch.
package server

import (
	"context"
	"crypto/md5"
	"dnsService/pkg/api"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// cmdExec а common function for executing cmd commands, accepts a command to execute
func cmdExec(req string) ([]byte, error) {
	writeCmd := exec.Command("/bin/sh", "-c", req)

	resp, err := writeCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// isAuthorized checks if the user with the 'name' on the server is authorized
func (s *grpcServer) isAuthorized(ctx context.Context, name string, t string) bool {
	token, err := s.db.Get(ctx, name).Result()
	if err != nil || token != t {
		return false
	}

	return true
}

// SetName checks if the user is logged in, sends a command to change the hostname and writes the changes to the database
func (s *grpcServer) SetName(ctx context.Context, req *api.SetNameRequest) (*api.SetNameResponse, error) {
	// check auth
	if !s.isAuthorized(ctx, req.User, req.Token) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// search password in env vars
	pass, ok := os.LookupEnv("ROOT_PSWD")
	if !ok {
		log.Println("Add(): ROOT_PSWD not found in environment")
	}

	// change hostname
	if _, err := cmdExec(fmt.Sprintf("echo %s | sudo -S sh -c 'hostname %s'", pass, req.Hostname)); err != nil {
		return &api.SetNameResponse{Ok: false, Message: err.Error()}, err
	}

	// cache data
	if err := s.db.Set(ctx, "hostname", req.Hostname, 0).Err(); err != nil {
		log.Println("SetName(): Failed to set hostname:", err)
	}

	return &api.SetNameResponse{Ok: true, Message: fmt.Sprintf("Hostname %s updated successfully", req.Hostname)}, nil
}

// GetName checks if there is cached data and sends it to the client,
// if there is no data, then executes a command to get the hostname and caches the response in the database
func (s *grpcServer) GetName(ctx context.Context, req *api.GetNameRequest) (*api.GetNameResponse, error) {
	// get cached data
	resp, err := s.db.Get(ctx, "hostname").Result()
	// if no cached data
	if err != nil {
		// get hostname
		val, err := cmdExec(fmt.Sprintf("hostname"))
		if err != nil {
			return &api.GetNameResponse{Ok: false, Message: err.Error()}, err
		}
		resp = string(val)

		// cache data
		if err := s.db.Set(ctx, "hostname", resp, 0).Err(); err != nil {
			log.Println("SetName(): Failed to get hostname:", err)
		}
	}

	return &api.GetNameResponse{Ok: true, Message: resp}, nil
}

// Add checks if the user is logged in to the server, searches for the sudo password in the environment variables,
// executes the command to add the dns server and caches the changes in the database
func (s *grpcServer) Add(ctx context.Context, req *api.AddRequest) (*api.AddResponse, error) {
	// check auth
	if !s.isAuthorized(ctx, req.User, req.Token) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// search password in env vars
	pass, ok := os.LookupEnv("ROOT_PSWD")
	if !ok {
		log.Println("Add(): ROOT_PSWD not found in environment")
	}

	// add dns server
	if _, err := cmdExec(fmt.Sprintf("echo %s | sudo -S sh -c 'cat >> /etc/resolv.conf << EOF\nnameserver %s\nEOF'",
		pass, req.DnsServer)); err != nil {
		return &api.AddResponse{Ok: false, Message: err.Error()}, err
	}

	// add data to cache
	if err := s.db.RPush(ctx, "dns_list", req.DnsServer).Err(); err != nil {
		log.Println("Add(): Failed to add dns server:", err)
	}

	return &api.AddResponse{Ok: true, Message: fmt.Sprintf("DNS server %s added successfully", req.DnsServer)}, nil
}

// Remove checks if the user is logged in to the server, searches for the sudo password in the environment variables,
// executes the command to remove the dns server and caches the changes in the database
func (s *grpcServer) Remove(ctx context.Context, req *api.RemoveRequest) (*api.RemoveResponse, error) {
	// check auth
	if !s.isAuthorized(ctx, req.User, req.Token) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	// search password in env vars
	pass, ok := os.LookupEnv("ROOT_PSWD")
	if !ok {
		log.Println("Add(): ROOT_PSWD not found in environment")
	}

	// remove dns server
	// make temporary without deleting dns file
	if _, err := cmdExec(fmt.Sprintf("echo %s | sudo -S grep -v 'nameserver %s' /etc/resolv.conf > tmp.conf",
		pass, req.DnsServer)); err != nil {
		return &api.RemoveResponse{Ok: false, Message: err.Error()}, nil
	}
	// move tmp file to /etc/
	if _, err := cmdExec(fmt.Sprintf("echo %s | sudo -S sh -c 'cat tmp.conf > /etc/resolv.conf'", pass)); err != nil {
		return &api.RemoveResponse{Ok: false, Message: err.Error()}, nil
	}

	// remove data from cache
	_, err := s.db.LRem(ctx, "dns_list", 0, req.DnsServer).Result()
	if err != nil {
		log.Println("Remove(): Failed to remove dns server:", err)
	}

	return &api.RemoveResponse{Ok: true, Message: fmt.Sprintf("DNS server %s removed successfully", req.DnsServer)}, nil
}

// List Takes information about the list of DNS servers from the database, if there is no record,
// then executes a command to get the list of DNS servers on the server and caches it to the database
func (s *grpcServer) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	// get cached data
	resp, err := s.db.LRange(ctx, "dns_list", 0, -1).Result()
	// if no data
	if len(resp) == 0 || err != nil {
		// get dns list
		val, err := cmdExec(fmt.Sprintf("grep 'nameserver' /etc/resolv.conf | awk '{print $2}'"))
		if err != nil {
			return &api.ListResponse{Ok: false, Message: []string{err.Error()}}, err
		}
		resp = strings.Split(string(val), "\n")

		// cache data
		if err := s.db.RPush(ctx, "dns_list", resp).Err(); err != nil {
			log.Println("Add(): Failed to add dns server:", err)
		}
	}

	return &api.ListResponse{Ok: true, Message: resp}, nil
}

// Auth downloads the secret key of the jwt token from the environment variables,
// compares md5 amounts to confirm the user's password,
// signs the jwt token with payload: username and expiration time = 24 hours, using HS256,
// records the user with the token in the database, issues the token to the user
func (s *grpcServer) Auth(ctx context.Context, req *api.AuthRequest) (*api.AuthResponse, error) {
	var t string
	var err error

	// user md5 sum password from env vars
	pswdHash, ok := os.LookupEnv("USER_PSWD")
	if !ok {
		log.Println("Auth(): USER_PSWD not found in environment")
		return &api.AuthResponse{Ok: false, Token: ""}, errors.New("USER_PSWD not found in environment")
	}
	// jwt secret key from env vars
	jwtSecretKey, ok := os.LookupEnv("JWT_KEY")
	if !ok {
		log.Println("Auth(): JWT_KEY not found in environment")
		return &api.AuthResponse{Ok: false, Token: ""}, errors.New("JWT_KEY not found in environment")
	}

	// if password entered correctly
	if fmt.Sprintf("%x", md5.Sum([]byte(req.Password))) != pswdHash {
		log.Println("Auth(): Invalid password")
		return &api.AuthResponse{Ok: false, Token: ""}, errors.New("invalid password")
	}

	// payload for signing jwt token
	payload := jwt.MapClaims{
		"sub": req.User,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	// create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// sign token
	t, err = token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		log.Println("Auth(): Failed to sign token:", err)
		return &api.AuthResponse{Ok: false, Token: ""}, err
	}

	// make note in db
	if err := s.db.Set(ctx, req.User, t, time.Hour*24).Err(); err != nil {
		log.Println("Auth(): Failed to set user token", err)
		return &api.AuthResponse{Ok: false, Token: ""}, err
	}

	return &api.AuthResponse{Ok: true, Token: t}, nil
}
