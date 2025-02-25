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


// putSecretValueCmd represents the put-secret-value command
var putSecretValueCmd = &cobra.Command{
    Use:   "put-secret-value",
    Short: "Puts new value for already existing Secret",
    Run: func(cmd *cobra.Command, args []string) {
        flags := cmd.Flags()
        params := map[string]interface{}{}

        
        

        // box id
        boxid, _ := flags.GetString("boxid")
        params["box_id"] = boxid

        // secret id
        secretid, _ := flags.GetString("secretid")
        params["secret_id"] = secretid

        // data
        dataProvided := flags.Changed("data")
        datakeyProvided := flags.Changed("datakey")
        datavalueProvided := flags.Changed("datavalue")
        datakvProvided := datakeyProvided || datavalueProvided

        // static secret
        if (!flags.Changed("managed-type")) {
            // check if managed secret type flags set. Abort if yes
            if (flags.Changed("ESXihost") || flags.Changed("ESXiuser") ||
                flags.Changed("ESXipasswd") || flags.Changed("ESXi-tls-version") ||
                flags.Changed("ESXicacert") || flags.Changed("master-boxid") ||
                flags.Changed("master-secretid")) {
                fmt.Println("Invalid flag(s)(one of -H, --ESXihost, -U, --ESXiuser, " +
                            "-P, --ESXipasswd, -T, --ESXi-tls-version, -c, --ESXicacert), " +
                            "-B, --master-boxid, -I, --master-secretid) specified " +
                            "for this secret type")
                os.Exit(1)
            }
            if (!dataProvided && !datakvProvided) {
                fmt.Println("Please specify --data or --datakey & --datavalue" +
                            " key-value pairs for static secret")
                os.Exit(1)
            }

            if (dataProvided && datakvProvided) {
                fmt.Println("Specify either secret data or secret data " +
                            "key-values. Not both.")
                os.Exit(1)
            }

            // get data string
            if (dataProvided) {
                data, _ := flags.GetString("data")
                params["secret_data"] = data
            // get data key-values
            } else {
                if ((datakeyProvided && !datavalueProvided) ||
                    (!datakeyProvided && datavalueProvided)) {
                        fmt.Println("Please provide both data key & values")
                        os.Exit(1)
                    }

                datakeyArray, _ := flags.GetStringArray("datakey")
                datavalueArray, _ := flags.GetStringArray("datavalue")
                if (len(datakeyArray) != len(datavalueArray)) {
                    fmt.Println("Please provide equal number of data keys & values")
                    os.Exit(1)
                }

                secretkvParams := map[string]interface{}{}
                for i := 0; i < len(datavalueArray); i +=1 {
                    if (IsJSON(datavalueArray[i])) {
                        secretkvParams[datakeyArray[i]] = JsonStrToMap(datavalueArray[i])
                    } else {
                        secretkvParams[datakeyArray[i]] = datavalueArray[i]
                    }
                }
                params["secret_data"] = secretkvParams
            }
        // managed secret
        } else {
            managedType, _ := flags.GetString("managed-type")
            if (managedType == "ESXiHostAccount") {
                // secret data
                if (dataProvided || datakvProvided) {
                    fmt.Println("Data arguments not required for managed-type ESXIHostAccount")
                    os.Exit(1)
                }

                if (!flags.Changed("ESXihost") || !flags.Changed("ESXiuser") ||
                    !flags.Changed("ESXipasswd")) {
                    fmt.Println("Please provide all of ESXihost, ESXiuser & ESXipasswd arguments")
                    os.Exit(1)
                }

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
            } else {
                fmt.Println("Unsupported managed-type specified")
                os.Exit(1)
            }
        }

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
    rootCmd.AddCommand(putSecretValueCmd)
    putSecretValueCmd.Flags().StringP("boxid", "b", "",
                                    "Id or name of the Box")
    putSecretValueCmd.Flags().StringP("secretid", "s", "",
                                    "Id or name of the Secret")
    putSecretValueCmd.Flags().StringP("data", "D", "",
                                    "Secret data")
    putSecretValueCmd.Flags().StringArrayP("datakey", "X", []string{},
                                         "The key to associate with the Secret data")
    putSecretValueCmd.Flags().StringArrayP("datavalue", "Y", []string{},
                                         "Value corresponding to specific Secret data")

    // managed type specific commands
    putSecretValueCmd.Flags().StringP("managed-type", "m", "",
                                    "Type of managed secret (supported: ESXiHostAccount)")
    putSecretValueCmd.Flags().StringP("ESXihost", "H", "",
                                    "ESXi host address if --managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("ESXiuser", "U", "",
                                    "ESXi username if --managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("ESXipasswd", "P", "",
                                    "ESXi password if --managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("ESXi-tls-version", "T", "",
                                    "TLS version to use while connecting to ESXi " +
                                    "if --managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("ESXicacert", "c", "",
                                    "CA Certificate to use while connecting to ESXi " +
                                    "if --managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("master-boxid", "B", "",
                                    "Box id or name of master secret, if " +
                                    "--managed-type set to ESXiHostAccount")
    putSecretValueCmd.Flags().StringP("master-secretid", "I", "",
                                    "Master Secret id or name, if " +
                                    "--managed-type set to ESXiHostAccount")

    // mark mandatory fields as required
    putSecretValueCmd.MarkFlagRequired("boxid")
    putSecretValueCmd.MarkFlagRequired("secretid")
}
