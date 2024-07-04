// Package cobra was partially generated with 'cobra-cli init'.
// It provides methods in the file clientMethods.go, structures in the file models.go,
// and variables for the cobra in the file cobraVars.go to implement the client app.
package cobra

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cobra",
	Short: "hostEditor for view and edit data on server",
	Long:  "The hostEditor will help you connect to the server,\n view data about its hostname, the DNS server list and change this data.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// init ...
func init() {
	rootCmd.PersistentFlags().StringVar(&client.conf.u, "u", "", "Type server's IP address")
	rootCmd.PersistentFlags().StringVar(&client.conf.p, "p", "", "Type port")

	rootCmd.AddCommand(server)
	server.AddCommand(get)
	server.AddCommand(set)
	server.AddCommand(login)

	get.AddCommand(hostname)
	get.AddCommand(dnsList)

	set.AddCommand(name)
	set.AddCommand(newDns)
	set.AddCommand(rmDns)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
