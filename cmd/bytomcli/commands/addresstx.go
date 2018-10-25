package commands

import (
	"os"

	"github.com/bytom/util"
	"github.com/spf13/cobra"
)

var addAddressCallbackCmd = &cobra.Command{
	Use:   "add-address-callback <address> <callbackUrl>",
	Short: "Add a callback url to be called when a tx happens on an address",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var ins = struct {
			Address string `json:"address"`
			URL     string `json:"url"`
		}{Address: args[0], URL: args[1]}

		data, exitCode := util.ClientCall("/add-address-callback", &ins)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}
		printJSON(data)
	},
}
var listAddressCallbacksCmd = &cobra.Command{
	Use:   "list-address-callbacks <address>",
	Short: "Show all callback urls added to an address",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var ins = struct {
			Address string `json:"address"`
		}{Address: args[0]}

		data, exitCode := util.ClientCall("/list-address-callbacks", &ins)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}
		printJSONList(data)
	},
}
var removeAddressCallbackCmd = &cobra.Command{
	Use:   "remove-address-callback <address> <callbackUrl>",
	Short: "Remove callback url added to an address",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		var ins = struct {
			Address string `json:"address"`
			URL     string `json:"url"`
		}{Address: args[0], URL: args[1]}

		data, exitCode := util.ClientCall("/remove-address-callback", &ins)
		if exitCode != util.Success {
			os.Exit(exitCode)
		}
		printJSON(data)
	},
}
