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
    "os"
    "fmt"
    "bytes"
    "encoding/json"
    // external
    "github.com/spf13/cobra"
)


// updateVaultHSMSettingsCmd represents the update-vault-hsm-settings command
var updateVaultHSMSettingsCmd = &cobra.Command{
    Use:   "update-vault-hsm-settings",
    Short: "Update PASM Vault HSM settings",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

	hsmStateProvided := flags.Changed("hsm-state")
        dekCacheTimeoutProvided := flags.Changed("dek-cache-timeout")
        if (!hsmStateProvided && !dekCacheTimeoutProvided) {
            fmt.Println("No HSM settings provided to update")
            os.Exit(1)
        }
        if (hsmStateProvided && dekCacheTimeoutProvided) {
            fmt.Println("Cannot update both at once. Specify one at a time")
            os.Exit(1)
        }

        // HSM state
        if hsmStateProvided {
            hsmState, _ := flags.GetString("hsm-state")
            if (hsmState == "enable" || hsmState == "disable" || hsmState == "rekey") {
	        params["hsm_state"] = hsmState
	    } else {
                fmt.Printf("\nInvalid -s, --hsm-state option: %s. " +
	          "Supported: enable, disable (or) rekey\n", hsmState)
                os.Exit(1)
	    }
	}

        // DEK cache timeout
	if dekCacheTimeoutProvided {
            dekCacheTimeout, _ := flags.GetInt("dek-cache-timeout")
	    if (dekCacheTimeout < 0) {
                fmt.Println("\n-d, --dek-cache-timeout option cannot be less than 0")
                os.Exit(1)
	    }
	    params["hsm_dek_cache_timeout"] = dekCacheTimeout
	}

        // revision
        revision, _ := flags.GetInt("revision")
        params["revision"] = revision

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "UpdateVaultHSMSettings")
        ret, err := DoPost(endpoint,
                               GetCACertFile(),
                               AuthTokenKV(),
                               jsonParams,
                               "application/json")
        if err != nil {
            fmt.Printf("\nHTTP request failed: %s\n", err)
            os.Exit(4)
        } else {
            // type assertion
            retBytes := ret["data"].(*bytes.Buffer)
            retStatus := ret["status"].(int)
            retStr := retBytes.String()

            if (retStr == "" && retStatus == 404) {
                fmt.Println("\nPASM Vault Settings not found\n")
                os.Exit(5)
            }

            fmt.Println("\n" + retStr + "\n")

            // make a decision on what to exit with
            retMap := JsonStrToMap(retStr)
            if _, present := retMap["error"]; present {
                os.Exit(3)
            } else {
                os.Exit(0)
            }
        }
    },
}

func init() {
    rootCmd.AddCommand(updateVaultHSMSettingsCmd)
    updateVaultHSMSettingsCmd.Flags().StringP("hsm-state", "s", "",
                                 "PASM Vault HSM state. To be one of enable, disable, rekey. ")
    updateVaultHSMSettingsCmd.Flags().IntP("dek-cache-timeout", "d", 0,
                                 "DEK cache timeout in seconds. ")
    updateVaultHSMSettingsCmd.Flags().IntP("revision", "R", 0,
                              "Revision number of PASM Vault settings")

    // mark mandatory fields as required
    updateVaultHSMSettingsCmd.MarkFlagRequired("revision")
}
