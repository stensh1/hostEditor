// Package cobra was partially generated with 'cobra-cli init'.
// It provides methods in the file clientMethods.go, structures in the file models.go,
// and variables for the cobra in the file cobraVars.go to implement the client app.
package cobra

import (
	"context"
	"dnsService/pkg/api"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	user2 "os/user"
	"strings"
	"syscall"
)

// c global object, that implements client
var client = &c{
	conf: config{},
}

// init parses cmd flags and writes it to c.conf
func (c *c) init(cmd *cobra.Command) {
	var err error

	// IP addr
	c.conf.u, err = cmd.Flags().GetString("u")
	if err != nil {
		log.Fatalf("Could not get flag 'u': %v", err)
	}
	// Port
	c.conf.p, err = cmd.Flags().GetString("p")
	if err != nil {
		log.Fatalf("Could not get flag 'p': %v", err)
	}
}

// newClient gets your current linux username and connects to server
func (c *c) newClient() {
	var err error

	// get linux username
	c.getUser()

	// make server connection
	c.conn, err = grpc.Dial(c.conf.u+":"+c.conf.p, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}

	// make object of EditorClient interface from 'api' package
	c.client = api.NewEditorClient(c.conn)
}

// closeClient destructor for c object
// closes connection
func (c *c) closeClient() {
	if err := c.conn.Close(); err != nil {
		log.Fatalf("Could not close client: %v", err)
	}
}

// getUser takes linux username and writes it in c
// or if asks user to type it manually if error
func (c *c) getUser() {
	user, err := user2.Current()
	if err != nil {
		// If you could not get linux username
		log.Printf("Could not take your username: %v\n", err.Error())
		fmt.Println("Enter your username: ")
		fmt.Scanf("%s", &c.username) //TODO: doesnt handles error
	} else {
		c.username = user.Username
	}
}

// getToken, knowing the name of the Linux user, tries to find his cookies data,
// if they cannot be found, calls the method of re-authorization on the server,
// if even after that cookies cannot be received, returns an error
func (c *c) getToken(cmd *cobra.Command, args []string) error {
	var err error

	// 1st try to find cookies
	c.token, err = os.ReadFile("/tmp/HostEditor/cookies/" + c.username + "/cookie.dat")
	if err != nil {
		fmt.Printf("%s, your token file not found\nTrying to log in...\n", c.username)
		c.login() // re-authorize
	}
	// 2nd try to find cookies
	c.token, err = os.ReadFile("/tmp/HostEditor/cookies/" + c.username + "/cookie.dat")
	if err != nil {
		return errors.New("failed to get token")
	}

	return nil
}

// hostname makes a request to get the server hostname using 'api' package
func (c *c) hostname() {
	// New server request
	req := &api.GetNameRequest{}
	// Getting a response back
	resp, err := c.client.GetName(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not get hostname: %v", err)
	}

	fmt.Println(c.conf.u, ":", resp.Message)
}

// dnsList makes a request to get the list of dns servers using 'api' package
func (c *c) dnsList() {
	// New server request
	req := &api.ListRequest{}
	// Getting a response back
	resp, err := c.client.List(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not get hostname: %v", err)
	}

	fmt.Println(c.conf.u, ":")
	fmt.Println(strings.Join(resp.Message, "\n"))
}

// login requests a password, sends a new authorization request to the server,
// processes the server's response and saves the received token in temporary files
func (c *c) login() {
	var pswd string

	// Password
	fmt.Println("Enter your password:")
	// Invisible typing
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	pswd = string(bytePassword)

	// New server request
	req := &api.AuthRequest{User: c.username, Password: pswd}
	// Getting a response back
	resp, err := c.client.Auth(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not get hostname: %v", err)
	}
	fmt.Println(c.conf.u, ":", "Logged in")

	// Creating directories in /tmp
	err = os.MkdirAll("/tmp/HostEditor/cookies/"+c.username, 0777)
	if err != nil {
		log.Fatalf("Could not save user's token")
	}

	// Creating new cookie file in new directory in /tmp
	file, err := os.Create("/tmp/HostEditor/cookies/" + c.username + "/cookie.dat")
	if err != nil {
		log.Fatalf("Failed to create file to store token: %v", err)
	}
	defer file.Close() // TODO: doesnt handles error

	// writing token in file
	_, err = file.WriteString(resp.Token)
	if err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	fmt.Println(c.conf.u, ":", "Cookies collected in: /tmp/HostEditor/cookies/"+c.username+"/cookie.dat")
}

// name creates a new request to change the server hostname
func (c *c) name(cmd *cobra.Command, args []string) {
	// we need a token for this operation
	if err := c.getToken(cmd, args); err != nil {
		log.Fatalf("%s, failed to get token: %v", c.username, err)
	}

	// New server request
	req := &api.SetNameRequest{Hostname: args[0], User: c.username, Token: string(c.token)}
	// Getting a response back
	resp, err := c.client.SetName(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not set hostname: %v", err)
	}

	fmt.Println(c.conf.u, ":", resp.Message)
}

// newDns creates a new request to add new write in dns servers list
func (c *c) newDns(cmd *cobra.Command, args []string) {
	// we need a token for this operation
	if err := c.getToken(cmd, args); err != nil {
		log.Fatalf("%s, failed to get token: %v", c.username, err)
	}

	// New server request
	req := &api.AddRequest{DnsServer: args[0], User: c.username, Token: string(c.token)}
	// Getting a response back
	resp, err := c.client.Add(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not add dns: %v", err)
	}

	fmt.Println(c.conf.u, ":", resp.Message)
}

// rmDns creates a new request to remove write from dns servers list
func (c *c) rmDns(cmd *cobra.Command, args []string) {
	// we need a token for this operation
	if err := c.getToken(cmd, args); err != nil {
		log.Fatalf("%s, failed to get token: %v", c.username, err)
	}

	// New server request
	req := &api.RemoveRequest{DnsServer: args[0], User: c.username, Token: string(c.token)}
	// Getting a response back
	resp, err := c.client.Remove(context.Background(), req)
	if err != nil {
		log.Fatalf("Could not remove dns: %v", err)
	}

	fmt.Println(c.conf.u, ":", resp.Message)
}
