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
	"crypto/rand"
	"crypto/rsa"
	"encoding/csv"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
)

const (
	SecretType               = "SSH key endpoint"
	ServersCsvFlag           = "servers-csv"
	BoxNameFlag              = "box-name"
	SecretExpiryFlag         = "secret-expiry"
	TagFlag                  = "tag"
	LeaseDurationFlag        = "lease-duration"
	SecretNamePatternFlag    = "secret-name-pattern"
	SecretDescriptionFlag    = "secret-description"
	ExclusiveCheckoutFlag    = "exclusive-checkout"
	DefaultSecretNamePattern = "ssh-secret-*"
	SshKeysDirectory         = "~/.ssh"
	SshKeysFile              = SshKeysDirectory + "/authorized_keys"
)

type SshServerInfo struct {
	Server      string
	Port        int
	User        string
	Password    string
	Private_key string
	Public_key  string
	Passphrase  string
}

type CommonSecretInfo struct {
	BoxName           string
	Tag               string
	SecretExpiry      string
	SecretNamePattern string
	SecretDescription string
	LeaseDuration     string
	ExclusiveCheckout bool
}

var setupProxyCmd = &cobra.Command{
	Use:   "setup-ssh-proxy",
	Short: "Setup key based SSH access on servers and onboard them in PASM vault for SSH proxy.",
	Run:   setupSSHProxy,
}

func init() {
	rootCmd.AddCommand(setupProxyCmd)

	cmd_flags := setupProxyCmd.Flags()
	cmd_flags.String(ServersCsvFlag,
		"",
		"Path to csv file having details of the server to be setup for the proxy.")
	cmd_flags.StringP(BoxNameFlag,
		"b",
		"",
		"Name of the box in which the secrets should be uploaded.")
	cmd_flags.String(SecretNamePatternFlag,
		"ssh-secret-*",
		"Pattern that secret names must follow. Currently '*' is supported within pattern. This '*' will replaced by numbers while creating secrets.")
	cmd_flags.String(SecretDescriptionFlag,
		"",
		"Description to be used for the secrets. This will be common across secrets. Secrets will not have any description if this flag is not set.")
	cmd_flags.String(SecretExpiryFlag,
		"",
		"Expiry of the secret in \"YYYY-MM-DD\" format. Will be common across all secrets created if set.")
	cmd_flags.String(LeaseDurationFlag,
		"",
		"Duration of lease to be granted for SSH secrets in ISO8601 format. This will be common across secrets if set.")
	cmd_flags.String(TagFlag,
		"",
		"Tag to be assigned to ssh secrets. This will be common across secrets if set.")
	cmd_flags.Bool(ExclusiveCheckoutFlag,
		false,
		"Pass this flag if the exclusive checkout option must be enabled for the secret. This will be common across secrets if set. ")

	setupProxyCmd.MarkFlagRequired(ServersCsvFlag)
	setupProxyCmd.MarkFlagRequired(BoxNameFlag)
}

func parseCommonSecretInfoFromCommandFlags(flags *pflag.FlagSet) (*CommonSecretInfo, error) {
	var commonSecretInfo CommonSecretInfo
	box_name, _ := flags.GetString(BoxNameFlag)
	commonSecretInfo.BoxName = box_name

	if flags.Changed(SecretNamePatternFlag) {
		secretNamePattern, _ := flags.GetString(SecretNamePatternFlag)
		if strings.Count(secretNamePattern, "*") > 1 {
			return nil, fmt.Errorf("Only one '*' is allowed in %s flag.", SecretNamePatternFlag)
		}
		commonSecretInfo.SecretNamePattern = secretNamePattern
	} else {
		commonSecretInfo.SecretNamePattern = DefaultSecretNamePattern
	}

	if flags.Changed(SecretDescriptionFlag) {
		secretDescription, _ := flags.GetString(SecretDescriptionFlag)
		commonSecretInfo.SecretDescription = secretDescription
	}

	if flags.Changed(SecretExpiryFlag) {
		secretExpiry, _ := flags.GetString(SecretExpiryFlag)
		commonSecretInfo.SecretExpiry = secretExpiry
	}

	if flags.Changed(LeaseDurationFlag) {
		leaseDuration, _ := flags.GetString(LeaseDurationFlag)
		commonSecretInfo.LeaseDuration = leaseDuration
	}

	if flags.Changed(ExclusiveCheckoutFlag) {
		exclusiveCheckout, _ := flags.GetBool(ExclusiveCheckoutFlag)
		commonSecretInfo.ExclusiveCheckout = exclusiveCheckout
	}

	if flags.Changed(TagFlag) {
		tag, _ := flags.GetString(TagFlag)
		commonSecretInfo.Tag = tag
	}

	return &commonSecretInfo, nil
}

func getServerCsvReader(csv_file *os.File) *csv.Reader {
	csv_reader := csv.NewReader(csv_file)
	csv_reader.TrimLeadingSpace = true
	csv_reader.FieldsPerRecord = 4
	csv_reader.ReuseRecord = true
	return csv_reader
}

func readKeyFile(filePath string) (string, error) {
	file_content_bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(file_content_bytes), nil
}

func populateServerInfoFromCsvRecord(record []string, serverInfo *SshServerInfo) error {
	port, err := strconv.Atoi(record[1])
	if err != nil {
		return err
	}
	serverInfo.Server = record[0]
	serverInfo.Port = port
	serverInfo.User = record[2]
	serverInfo.Password = record[3]
	return nil
}

func initiateSshConnectionWithPassword(serverInfo *SshServerInfo) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: serverInfo.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(serverInfo.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", serverInfo.Server, serverInfo.Port), config)
}

func initiateSshConnectionWithKey(serverInfo *SshServerInfo) (*ssh.Client, error) {
	signer, _ := ssh.ParsePrivateKeyWithPassphrase([]byte(serverInfo.Private_key), []byte(serverInfo.Passphrase))
	config := &ssh.ClientConfig{
		User: serverInfo.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", fmt.Sprintf("%s:%d", serverInfo.Server, serverInfo.Port), config)
}

func setupKeybasedSshAccess(serverInfo *SshServerInfo) error {
	sshClientWithPw, clientErr := initiateSshConnectionWithPassword(serverInfo)
	if clientErr != nil {
		return clientErr
	}
	defer sshClientWithPw.Close()

	cmdErr := runCommandOnSshServer(sshClientWithPw, fmt.Sprintf("mkdir -p %s && chmod 700 %s", SshKeysDirectory, SshKeysDirectory))
	if cmdErr != nil {
		return cmdErr
	}

	cmdErr = runCommandOnSshServer(sshClientWithPw, fmt.Sprintf("echo \"%s\" >> %s && chmod 600 %s", serverInfo.Public_key, SshKeysFile, SshKeysFile))
	if cmdErr != nil {
		return cmdErr
	}

	verifyErr := verifyKeybasedSshAccess(serverInfo)
	if verifyErr != nil {
		runCommandOnSshServer(sshClientWithPw, fmt.Sprintf("sed -i.bak '/^%s/d'", serverInfo.Public_key))
		return verifyErr
	}

	return nil
}

func runCommandOnSshServer(sshClient *ssh.Client, cmd string) error {
	sshSession, sessionErr := sshClient.NewSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer sshSession.Close()

	cmdErr := sshSession.Run(cmd)
	if cmdErr != nil {
		return cmdErr
	}

	return nil
}

func verifyKeybasedSshAccess(serverInfo *SshServerInfo) error {
	sshClientWithKey, client_err := initiateSshConnectionWithKey(serverInfo)
	if client_err != nil {
		return client_err
	}
	defer sshClientWithKey.Close()
	sshSessionWithKey, session_error := sshClientWithKey.NewSession()
	if session_error != nil {
		return session_error
	}
	defer sshSessionWithKey.Close()
	return sshSessionWithKey.Run("hostname")
}

func generateAndPopulateSSHKeysInServerInfo(serverInfo *SshServerInfo) error {
	rsaPrivateKey, keyGenErr := rsa.GenerateKey(rand.Reader, 4096)
	if keyGenErr != nil {
		return keyGenErr
	}

	sshPubKey, pubKeyErr := ssh.NewPublicKey(rsaPrivateKey.Public())
	if pubKeyErr != nil {
		return pubKeyErr
	}

	passphrase := GenerateRandomString(14)
	sshPrivKey, keyMarshalErr := ssh.MarshalPrivateKeyWithPassphrase(rsaPrivateKey, "", []byte(passphrase))
	if keyMarshalErr != nil {
		return keyMarshalErr
	}

	serverInfo.Private_key = string(pem.EncodeToMemory(sshPrivKey))
	serverInfo.Public_key = string(ssh.MarshalAuthorizedKey(sshPubKey))
	serverInfo.Passphrase = passphrase

	return nil
}

func setupSSHProxy(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	csv_path, _ := flags.GetString(ServersCsvFlag)

	csv_file, file_err := os.Open(csv_path)
	if file_err != nil {
		fmt.Printf("Invalid value provided for parameter 'servers'. %s", file_err)
		os.Exit(2)
	}
	defer csv_file.Close()
	commonSecretInfo, parse_error := parseCommonSecretInfoFromCommandFlags(flags)

	if parse_error != nil {
		fmt.Printf("Error: %s\n", parse_error.Error())
		os.Exit(2)
	}

	var serverInfos []SshServerInfo
	var successfulServers []SshServerInfo
	var unsuccessfulServers []SshServerInfo
	csv_reader := getServerCsvReader(csv_file)
	for {
		record, csv_read_err := csv_reader.Read()
		if csv_read_err == io.EOF {
			break
		}
		if csv_read_err != nil {
			fmt.Printf("%s. This row will be skipped from processing.", csv_read_err.Error())
			continue
		}
		var serverInfo SshServerInfo

		populate_err := populateServerInfoFromCsvRecord(record, &serverInfo)
		if populate_err != nil {
			fmt.Printf("Error parsing record. %s. This row will be skipped from processing.\n", populate_err.Error())
			continue
		}

		key_err := generateAndPopulateSSHKeysInServerInfo(&serverInfo)
		if key_err != nil {
			fmt.Printf("Error occurred. %s. Server %s will be skipped from setup.\n", key_err.Error(), serverInfo.Server)
			continue
		}

		serverInfos = append(serverInfos, serverInfo)
	}
	for i := 0; i < len(serverInfos); i++ {
		serverInfo := serverInfos[i]
		setupErr := setupKeybasedSshAccess(&serverInfo)
		if setupErr != nil {
			fmt.Printf("Error occurred while setting up key based access for server %s. %s. This server will not be onboarded in PASM vault.\n", serverInfo.Server, setupErr.Error())
			unsuccessfulServers = append(unsuccessfulServers, serverInfo)
			continue
		}
		successfulServers = append(successfulServers, serverInfo)
	}
	printUnsuccessfulServerList(unsuccessfulServers)

	if len(successfulServers) > 0 {
		csvRecords := createCsvRecordsForServers(successfulServers, commonSecretInfo)
		csvFileToUpload, writeErr := writeCsvToDisk(*csvRecords)
		if writeErr != nil {
			fmt.Printf("Error writing ssh secrets csv to disk. %s\n", writeErr.Error())
			os.Exit(2)
		}

		defer os.Remove(csvFileToUpload)
		params := map[string]interface{}{}
		params["csv_file"] = csvFileToUpload
		params["secret_type"] = SecretType
		uploadCsv(params)
		fmt.Println("CSV upload for servers for which key based SSH access setup was successful. Check status using 'get-csv-import-status' command.")
	}
}

func printUnsuccessfulServerList(unsuccessfulServers []SshServerInfo) {
	if len(unsuccessfulServers) > 0 {
		fmt.Println("Could not setup and onboard following servers.")
		for i := 0; i < len(unsuccessfulServers); i++ {
			fmt.Println(unsuccessfulServers[i].Server)
		}
	}
}

func writeCsvToDisk(records [][]string) (string, error) {
	file, temp_err := os.CreateTemp("", "ssh-secret-csv")
	if temp_err != nil {
		return "", temp_err
	}
	defer file.Close()
	csvWriter := csv.NewWriter(file)
	write_error := csvWriter.WriteAll(records)
	if write_error != nil {
		return "", write_error
	}
	csvWriter.Flush()
	return file.Name(), nil
}

func createCsvRecordsForServers(successfulServers []SshServerInfo, commonSecretInfo *CommonSecretInfo) *[][]string {
	var csvRecords [][]string
	csvRecords = append(csvRecords, *createCsvHeaderRecord())

	for i := 0; i < len(successfulServers); i++ {
		csvRecords = append(csvRecords, *generateCsvRecordForServer(successfulServers[i], commonSecretInfo, i+1))
	}

	return &csvRecords
}

func createCsvHeaderRecord() *[]string {
	var record []string
	record = append(record,
		"box_name",
		"secret_name",
		"host",
		"user",
		"keyfile",
		"desc",
		"port",
		"keypwd",
		"lease-duration",
		"exclusive-checkout",
		"tags",
		"expires_at")
	return &record
}

func generateCsvRecordForServer(serverInfo SshServerInfo, commonSecretInfo *CommonSecretInfo, recordIndex int) *[]string {
	var record []string
	secretName := strings.Replace(commonSecretInfo.SecretNamePattern, "*", strconv.Itoa(recordIndex), 1)
	privateKey := strings.ReplaceAll(serverInfo.Private_key, "\"", "")
	record = append(record,
		commonSecretInfo.BoxName,
		secretName,
		serverInfo.Server,
		serverInfo.User,
		privateKey,
		commonSecretInfo.SecretDescription,
		strconv.Itoa(serverInfo.Port),
		serverInfo.Passphrase,
		commonSecretInfo.LeaseDuration,
		strconv.FormatBool(commonSecretInfo.ExclusiveCheckout),
		commonSecretInfo.Tag,
		commonSecretInfo.SecretExpiry,
	)
	return &record
}
