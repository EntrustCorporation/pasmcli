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


// putEsxiHostSecretValueCmd represents the put-secret-value command
var putEsxiHostSecretValueCmd = &cobra.Command{
    Use:   "put-esxi-host-secret-value",
    Short: "Put new secret value for already existing ESXi Host Secret",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        ESXiHost, _ := flags.GetString("ESXihost")
        ESXiUser, _ := flags.GetString("ESXiuser")
        ESXiPasswd, _ := flags.GetString("ESXipasswd")
        secretData := map[string]interface{}{}
        secretData["host"] = ESXiHost
        secretData["uid"] = ESXiUser
        secretData["password"] = ESXiPasswd
        params["secret_data"] = secretData

        // secret config
        secretConfig := map[string]interface{}{}
        // secret config type
        secretConfig["type"] = "ESXi Host Account"

        // secret config master
        if ((flags.Changed("master-boxid") && !flags.Changed("master-secretid")) ||
            (!flags.Changed("master-boxid") && flags.Changed("master-secretid"))) {
            fmt.Println("Please provide both master-boxid as well as master-secretid")
            os.Exit(1)
        }

        if (flags.Changed("master-boxid") && flags.Changed("master-secretid")) {
            masterBoxId, _ := flags.GetString("master-boxid")
            masterSecretId, _ := flags.GetString("master-secretid")

            masterSecretConfig := map[string]interface{}{}
            masterSecretConfig["box_id"] = masterBoxId
            masterSecretConfig["secret_id"] = masterSecretId
            secretConfig["master_secret"] = masterSecretConfig
        }
        // secret config ESXi TLS version
        if (flags.Changed("ESXi-tls-version")) {
            ESXiTLSVersion, _ := flags.GetString("ESXi-tls-version")
            secretConfig["tls_version"] = ESXiTLSVersion
        }
        // secret config ESXi CA Certificate
        if (flags.Changed("ESXicacert")) {
            ESXiCACert, _ := flags.GetString("ESXicacert")
            secretConfig["ca_cert"] = ESXiCACert
        }
        params["secret_config"] = secretConfig

        // JSONify
        jsonParams, err := json.Marshal(params)
        if (err != nil) {
            fmt.Println("Error building JSON request: ", err)
            os.Exit(1)
        }

        // now POST
        endpoint := GetEndPoint("", "1.0", "PutSecretValue")
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
                fmt.Println("\nSecret not found\n")
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
    rootCmd.AddCommand(putEsxiHostSecretValueCmd)
    putEsxiHostSecretValueCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    putEsxiHostSecretValueCmd.Flags().StringP("secretid", "s", "",
                                    "Id or name of the Secret")

    putEsxiHostSecretValueCmd.Flags().StringP("ESXihost", "H", "",
                                    "ESXi host address if --managed-type set to ESXiHostAccount")
    putEsxiHostSecretValueCmd.Flags().StringP("ESXiuser", "U", "",
                                    "ESXi username if --managed-type set to ESXiHostAccount")
    putEsxiHostSecretValueCmd.Flags().StringP("ESXipasswd", "P", "",
                                    "ESXi password if --managed-type set to ESXiHostAccount")
    putEsxiHostSecretValueCmd.Flags().StringP("ESXi-tls-version", "T", "",
                                    "TLS version to use while connecting to ESXi")
    putEsxiHostSecretValueCmd.Flags().StringP("ESXicacert", "c", "",
                                    "CA Certificate to use while connecting to ESXi")
    putEsxiHostSecretValueCmd.Flags().StringP("master-boxid", "B", "",
                                    "(optional) Box id or name of master secret")
    putEsxiHostSecretValueCmd.Flags().StringP("master-secretid", "I", "",
                                    "(optional) Master Secret id or name")

    // mark mandatory fields as required
    putEsxiHostSecretValueCmd.MarkFlagRequired("boxid")
    putEsxiHostSecretValueCmd.MarkFlagRequired("secretid")

    putEsxiHostSecretValueCmd.MarkFlagRequired("ESXihost")
    putEsxiHostSecretValueCmd.MarkFlagRequired("ESXiuser")
    putEsxiHostSecretValueCmd.MarkFlagRequired("ESXipasswd")
}
