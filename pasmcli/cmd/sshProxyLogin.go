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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var sshProxyLoginCmd = &cobra.Command{
	Use:   "ssh-proxy-login",
	Short: "Login to a server using SSH Proxy",
	Run:   loginToProxy,
}

func loginToProxy(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	serverName, _ := flags.GetString("server")

	resp, err := DoGet(getEndpointUrlWithServerQueryParam(serverName),
		GetCACertFile(),
		AuthTokenKV(),
		make([]byte, 0),
		"application/json")
	if err != nil {
		fmt.Printf("Error occurred while fetching server details from the PASM vault. %s\n", err)
		os.Exit(2)
	}
	validateResponse(resp)
	proxyInfoResponse := parseResponse(resp)
	infoArray := proxyInfoResponse.ProxyInfos
	var serverToConnect *proxyPortInfo

	switch {
	case len(infoArray) == 0:
		fmt.Printf("No Secret related to server %s found or the secret may not be accessible due to access policy.", serverName)
		os.Exit(3)
	case len(infoArray) == 1:
		serverToConnect = &infoArray[0]
	default:
		serverToConnect = getServerChoiceFromUser(infoArray, serverName)
	}
	initiateProxy(serverToConnect, proxyInfoResponse.PasmUserName)
}

func getEndpointUrlWithServerQueryParam(server string) string {
	endpoint_url, _ := url.Parse(GetEndPoint("", "1.0", "GetSSHProxyPortInfoForServer"))
	query := endpoint_url.Query()
	query.Add("server", server)
	endpoint_url.RawQuery = query.Encode()
	return endpoint_url.String()
}

func validateResponse(response map[string]interface{}) {
	response_status := response["status"].(int)
	if response_status != 200 {
		fmt.Printf("Bad response while fetching server details from PASM vault. Response Code: %d\n", response_status)
		os.Exit(4)
	}
}

func parseResponse(response map[string]interface{}) proxyInfo {
	body := response["data"]
	var proxyInfos proxyInfo
	err := json.NewDecoder(body.(*bytes.Buffer)).Decode(&proxyInfos)
	if err != nil {
		fmt.Printf("Error occurred while decoding response. %s\n", err)
		os.Exit(5)
	}
	return proxyInfos
}

func getServerChoiceFromUser(infoArray []proxyPortInfo, serverName string) *proxyPortInfo {
	fmt.Printf("Multiple secrets found with %s. Choose an option from the given list.\n", serverName)
	for i := 0; i < len(infoArray); i++ {
		proxyDetail := infoArray[i]
		fmt.Printf("%d. login as %s on proxy port %d\n", i+1, proxyDetail.UserName, proxyDetail.Port)
	}
	fmt.Printf("Your Choice (Enter option number or 'q' to quit): ")
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSuffix(choice, "\n")
	choice = strings.TrimSuffix(choice, "\r")
	if strings.ToLower(choice) == "q" {
		os.Exit(0)
	}
	choice_int, err := strconv.Atoi(choice)
	if err != nil || choice_int < 1 || choice_int > len(infoArray) {
		fmt.Println("Wrong choice provided.")
		os.Exit(6)
	}
	return &infoArray[choice_int-1]
}

func initiateProxy(serverDetail *proxyPortInfo, pasmUserName string) {
	cmd := exec.Command("ssh",
		fmt.Sprintf("-p %d", serverDetail.Port),
		fmt.Sprintf("%s@%s", pasmUserName, GetServer()))

	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error occurred while starting SSH command. %s\n", err)
		os.Exit(7)
	}

	cmd.Wait()
}

func init() {
	rootCmd.AddCommand(sshProxyLoginCmd)
	sshProxyLoginCmd.Flags().StringP("server", "s", "",
		"FQDN/IP of the server to which the SSH connection via proxy must be established.")
	sshProxyLoginCmd.MarkFlagRequired("server")
}
