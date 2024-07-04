// Package cobra was partially generated with 'cobra-cli init'.
// It provides methods in the file clientMethods.go, structures in the file models.go,
// and variables for the cobra in the file cobraVars.go to implement the client app.
package cobra

import (
	"github.com/spf13/cobra"
)

// main command
var server = &cobra.Command{
	Use:              "s [flags] [commands]",
	Short:            "Choose server to connect.",
	Long:             "You should point which IP address and which port you want to connect to the server.",
	TraverseChildren: true,
	// Handle flags before running
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		client.init(cmd)
	},
	Run: func(cmd *cobra.Command, args []string) {
		client.newClient()
	},
}

// command to get some information from server: hostname or dns list
var get = &cobra.Command{
	Use:   "get",
	Short: "Get some information from server.",
	Long: "This command allows you to get some information from the server like a server's hostname or list of server's dns servers.\n" +
		"This command does nothing, you need to choose the next one.",
}

// command to get server's hostname
var hostname = &cobra.Command{
	Use:   "hostname",
	Short: "Get hostname from server.",
	Long:  "This command allows you to get hostname from the server.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.hostname()
		client.closeClient()
	},
}

// command to get server's dns list
var dnsList = &cobra.Command{
	Use:   "dns",
	Short: "Get dns list from server.",
	Long:  "This command allows you to get dns servers list from the server.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.dnsList()
		client.closeClient()
	},
}

// command to get token from server
var login = &cobra.Command{
	Use:   "login",
	Short: "Login to server to execute sudo commands.",
	Long:  "Login allows you to get a token to perform actions on the server that require sudo.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.login()
		client.closeClient()
	},
}

// command to set or change some info on server: hostname or dns list
var set = &cobra.Command{
	Use:   "set",
	Short: "Set some information on server.",
	Long: "This command allows you to set some information on the server like a server's hostname or new dns server.\n" +
		"This command does nothing, you need to choose the next one.",
}

// command to change server's hostname
var name = &cobra.Command{
	Use:   "name [option]",
	Short: "Set the new hostname on server.",
	Long: "This command allows you to set the new hostname on server.\n" +
		"You need to be logged in to use it.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.name(cmd, args)
		client.closeClient()
	},
}

// command to add new dns write
var newDns = &cobra.Command{
	Use:   "new_dns [option]",
	Short: "Add the new dns server.",
	Long: "This command allows you to add the new dns on server.\n" +
		"You need to be logged in to use it.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.newDns(cmd, args)
		client.closeClient()
	},
}

// command to remove dns write
var rmDns = &cobra.Command{
	Use:   "rm_dns [option]",
	Short: "Remove dns server.",
	Long: "This command allows you to remove dns on server.\n" +
		"You need to be logged in to use it.",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cmd, args)
		client.rmDns(cmd, args)
		client.closeClient()
	},
}
