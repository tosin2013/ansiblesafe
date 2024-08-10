package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/hashicorp/vault/api"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
    RhsmUsername              string `yaml:"rhsm_username"`
    RhsmPassword              string `yaml:"rhsm_password"`
    RhsmOrg                   string `yaml:"rhsm_org"`
    RhsmActivationKey         string `yaml:"rhsm_activationkey"`
    AdminUserPassword         string `yaml:"admin_user_password"`
    OfflineToken              string `yaml:"offline_token"`
    AutomationHubOfflineToken string `yaml:"automation_hub_offline_token"`
    OpenShiftPullSecret       string `yaml:"openshift_pull_secret"`
    FreeIpaServerPassword     string `yaml:"freeipa_server_admin_password"`
    AwsAccessKey              string `yaml:"aws_access_key"`
    AwsSecretKey              string `yaml:"aws_secret_key"`
}

func main() {
    vaultPath, err := findAnsibleVault()
    if err != nil {
        log.Fatalf("Error: %s", err)
    }

    fmt.Printf("ansible-vault found at: %s\n", vaultPath)
    var filePath string
    var choice int

    pflag.StringVarP(&filePath, "file", "f", "", "Path to YAML file (default: $HOME/vault.yml)")
    pflag.IntVarP(&choice, "operation", "o", 0, "Operation to perform (1: encrypt, 2: decrypt, 3: Write secrets to HashiCorp Vault, 4: Read secrets from HashiCorp Vault, 5: Read secrets from HCP, 6: skip encrypting/decrypting)")
    pflag.Parse()

    if filePath == "" {
        usr, err := user.Current()
        if err != nil {
            log.Fatalf("Error getting current user: %s", err)
        }
        filePath = filepath.Join(usr.HomeDir, "vault.yml")
    }

    // Handle operations based on choice
    switch choice {
    case 1:
        encryptFile(filePath, vaultPath)
    case 2:
        decryptFile(filePath, vaultPath)
    case 3:
        writeSecretsToVault(filePath)
    case 4:
        readSecretsFromVault(filePath)
    case 5:
        readSecretsFromHCP(filePath)
    case 6:
        fmt.Println("Skipping file encryption/decryption")
    default:
        fmt.Println("1. Encrypt vault.yml file")
        fmt.Println("2. Decrypt vault.yml file")
        fmt.Println("3. Write secrets to HashiCorp Vault")
        fmt.Println("4. Read secrets from HashiCorp Vault")
        fmt.Println("5. Read secrets from HCP")
        fmt.Println("6. Skip file encryption/decryption")
        notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
        notice("Please choose an option: ")
        fmt.Scanln(&choice)
    }
}

func findAnsibleVault() (string, error) {
    vaultPath, err := exec.LookPath("ansible-vault")
    if err != nil {
        return "", fmt.Errorf("ansible-vault not found in PATH. Please install it before using this script")
    }
    return vaultPath, nil
}

func encryptFile(filePath, vaultPath string) {
    fileBytes, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Error reading file %s: %s", filePath, err)
    }
    if !strings.Contains(string(fileBytes), "ANSIBLE_VAULT") {
        vaultCommand := fmt.Sprintf("ansible-vault encrypt %s", filePath)
        executeCommand(vaultCommand)
    } else {
        log.Fatalf("Error: %s is already encrypted.", filePath)
    }
}

func decryptFile(filePath, vaultPath string) {
    vaultCommand := fmt.Sprintf("ansible-vault decrypt %s", filePath)
    executeCommand(vaultCommand)
}

func writeSecretsToVault(filePath string) {
    configData, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Error reading file %s: %s", filePath, err)
    }

    var config Configuration
    err = yaml.Unmarshal(configData, &config)
    if err != nil {
        log.Fatalf("Error unmarshalling YAML data: %s", err)
    }

    vaultAddress := os.Getenv("VAULT_ADDRESS")
    vaultToken := os.Getenv("VAULT_TOKEN")
    secretPath := os.Getenv("SECRET_PATH")
    validateEnvVars(vaultAddress, vaultToken, secretPath)

    vaultClient := createVaultClient(vaultAddress, vaultToken)
    secretData := map[string]interface{}{
        "rhsm_username":                config.RhsmUsername,
        "rhsm_password":                config.RhsmPassword,
        "rhsm_org":                     config.RhsmOrg,
        "rhsm_activationkey":           config.RhsmActivationKey,
        "admin_user_password":          config.AdminUserPassword,
        "offline_token":                config.OfflineToken,
        "automation_hub_offline_token": config.AutomationHubOfflineToken,
        "openshift_pull_secret":        config.OpenShiftPullSecret,
        "freeipa_server_admin_password": config.FreeIpaServerPassword,
        "aws_access_key":               config.AwsAccessKey,
        "aws_secret_key":               config.AwsSecretKey,
    }

    ctx := context.Background()
    kv2 := vaultClient.KVv2("ansiblesafe")
    _, err = kv2.Put(ctx, secretPath, secretData)
    if err != nil {
        log.Fatalf("Error writing secret to Vault: %s", err)
    }

    fmt.Println("Secrets written to Vault successfully.")
}

func readSecretsFromVault(filePath string) {
    vaultAddress := os.Getenv("VAULT_ADDRESS")
    vaultToken := os.Getenv("VAULT_TOKEN")
    secretPath := os.Getenv("SECRET_PATH")
    validateEnvVars(vaultAddress, vaultToken, secretPath)

    readSecretsFromVaultAndWriteToFile(filePath, vaultAddress, vaultToken, secretPath)
}

func readSecretsFromHCP(filePath string) {
    hcpClientID := os.Getenv("HCP_CLIENT_ID")
    hcpClientSecret := os.Getenv("HCP_CLIENT_SECRET")
    organizationID := os.Getenv("HCP_ORG_ID")
    projectID := os.Getenv("HCP_PROJECT_ID")
    appName := os.Getenv("APP_NAME")
    if hcpClientID == "" || hcpClientSecret == "" || organizationID == "" || projectID == "" || appName == "" {
        log.Fatalf("HCP_CLIENT_ID, HCP_CLIENT_SECRET, HCP_ORG_ID, HCP_PROJECT_ID, and APP_NAME must be set.")
    }

    token, err := getHCPAPIToken(hcpClientID, hcpClientSecret)
    if err != nil {
        log.Fatalf("Error retrieving HCP API token: %s", err)
    }

    secrets, err := getHCPSecrets(token, organizationID, projectID, appName)
    if err != nil {
        log.Fatalf("Error retrieving HCP secrets: %s", err)
    }

    err = ioutil.WriteFile(filePath, secrets, 0644)
    if err != nil {
        log.Fatalf("Error writing secrets to file: %s", err)
    }

    fmt.Println("Secrets read from HCP and written to file successfully.")
}


func executeCommand(command string) {
    cmd := exec.Command("bash", "-c", command)
    cmd.Env = append(os.Environ(), "ANSIBLE_VAULT_PASSWORD_FILE="+os.Getenv("HOME")+"/.vault_password")
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Error executing vault command: %s", err)
        log.Printf("Command: %s", command)
        log.Printf("Output: %s", string(output))
        log.Fatalf("Failed to execute vault command")
    }
    fmt.Println(string(output))
}

func validateEnvVars(vaultAddress, vaultToken, secretPath string) {
    if vaultAddress == "" {
        log.Fatalf("Error: VAULT_ADDRESS environment variable is not set.")
    }
    if vaultToken == "" {
        log.Fatalf("Error: VAULT_TOKEN environment variable is not set.")
    }
    if secretPath == "" {
        log.Fatalf("Error: SECRET_PATH environment variable is not set.")
    }
}

func createVaultClient(vaultAddress, vaultToken string) *api.Client {
    vaultConfig := api.DefaultConfig()
    vaultConfig.Address = vaultAddress
    vaultConfig.Timeout = 10 * time.Second
    vaultClient, err := api.NewClient(vaultConfig)
    if err != nil {
        log.Fatalf("Error creating Vault client: %s", err)
    }
    vaultClient.SetToken(vaultToken)
    return vaultClient
}

func readSecretsFromVaultAndWriteToFile(filePath, vaultAddress, vaultToken, secretPath string) error {
    vaultClient := createVaultClient(vaultAddress, vaultToken)

    ctx := context.Background()
    kv2 := vaultClient.KVv2("ansiblesafe")

    secret, err := kv2.Get(ctx, secretPath)
    if err != nil {
        log.Fatalf("Error retrieving secret from Vault: %v", err)
    }

    if secret == nil {
        log.Fatalf("Error: Secret not found at path %s", secretPath)
    }

    config := Configuration{
        RhsmUsername:          secret.Data["rhsm_username"].(string),
        RhsmPassword:          secret.Data["rhsm_password"].(string),
        RhsmOrg:               secret.Data["rhsm_org"].(string),
        RhsmActivationKey:     secret.Data["rhsm_activationkey"].(string),
        AdminUserPassword:     secret.Data["admin_user_password"].(string),
        OfflineToken:          secret.Data["offline_token"].(string),
        OpenShiftPullSecret:   secret.Data["openshift_pull_secret"].(string),
        FreeIpaServerPassword: secret.Data["freeipa_server_admin_password"].(string),
        AwsAccessKey:          secret.Data["aws_access_key"].(string),
        AwsSecretKey:          secret.Data["aws_secret_key"].(string),
    }

    configData, err := yaml.Marshal(config)
    if err != nil {
        log.Fatalf("Error marshaling YAML data: %v", err)
    }

    err = ioutil.WriteFile(filePath, configData, 0644)
    if err != nil {
        log.Fatalf("Error writing vault file: %v", err)
    }
    fmt.Println("Secrets read from Vault and written to file successfully.")

    return nil
}

func getHCPAPIToken(clientID, clientSecret string) (string, error) {
    url := "https://auth.idp.hashicorp.com/oauth2/token"
    data := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials&audience=https://api.hashicorp.cloud", clientID, clientSecret)

    resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
    if err != nil {
        return "", fmt.Errorf("error making request to get HCP API token: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("error response from HCP API token endpoint: %s", string(body))
    }

    var result map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&result)
    if err != nil {
        return "", fmt.Errorf("error decoding HCP API token response: %w", err)
    }

    token, ok := result["access_token"].(string)
    if !ok {
        return "", fmt.Errorf("access_token not found in HCP API token response")
    }
    return token, nil
}

func getHCPSecrets(token, organizationID, projectID, appID string) ([]byte, error) {
    url := fmt.Sprintf("https://api.cloud.hashicorp.com/secrets/2023-06-13/organizations/%s/projects/%s/apps/%s/open", organizationID, projectID, appID)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating HCP secrets request: %w", err)
    }
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request to HCP secrets endpoint: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("error response from HCP secrets endpoint: %s", string(body))
    }

    // Decode the JSON response into a map
    var secretResponse struct {
        Secrets []struct {
            Name    string `json:"name"`
            Version struct {
                Value string `json:"value"`
            } `json:"version"`
        } `json:"secrets"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&secretResponse); err != nil {
        return nil, fmt.Errorf("error decoding HCP secrets response: %w", err)
    }

    // Create the configuration object based on the received data
    config := Configuration{}
    for _, secret := range secretResponse.Secrets {
        switch secret.Name {
        case "rhsm_username":
            config.RhsmUsername = secret.Version.Value
        case "rhsm_password":
            config.RhsmPassword = secret.Version.Value
        case "rhsm_org":
            config.RhsmOrg = secret.Version.Value
        case "rhsm_activationkey":
            config.RhsmActivationKey = secret.Version.Value
        case "admin_user_password":
            config.AdminUserPassword = secret.Version.Value
        case "offline_token":
            config.OfflineToken = secret.Version.Value
        case "automation_hub_offline_token":
            config.AutomationHubOfflineToken = secret.Version.Value
        case "openshift_pull_secret":
            config.OpenShiftPullSecret = secret.Version.Value
        case "freeipa_server_admin_password":
            config.FreeIpaServerPassword = secret.Version.Value
        case "aws_access_key":
            config.AwsAccessKey = secret.Version.Value
        case "aws_secret_key":
            config.AwsSecretKey = secret.Version.Value
        }
    }

    yamlData, err := yaml.Marshal(config)
    if err != nil {
        return nil, fmt.Errorf("error marshaling secrets to YAML: %w", err)
    }

    return yamlData, nil
}

func getStringValue(data map[string]interface{}, key string) string {
    if value, ok := data[key]; ok {
        if strValue, ok := value.(string); ok {
            return strValue
        }
    }
    return ""
}

