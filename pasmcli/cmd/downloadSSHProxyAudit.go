/*
 Copyright 2020-2025 Entrust Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	// standard
	"fmt"
	"os"

	// external
	"github.com/spf13/cobra"
)

// downloadSSHProxyAuditCmd represents the download-ssh-proxy-audit command
var downloadSSHProxyAuditCmd = &cobra.Command{
	Use:   "download-ssh-proxy-audit",
	Short: "Download SSH Proxy audit log bundle",
	Run: func(cmd *cobra.Command, args []string) {

		// now GET
		endpoint := GetEndPoint("", "1.0", "GetSSHProxyAuditBundle")
		fname, err := DoGetDownload(endpoint, GetCACertFile(), AuthTokenKV())
		if err != nil {
			fmt.Printf("\nHTTP request failed: %s\n", err)
			os.Exit(4)
		} else {
			fmt.Println("\nSuccessfully downloaded ssh proxy audit log bundle " +
				"as - " + fname + "\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadSSHProxyAuditCmd)
}
