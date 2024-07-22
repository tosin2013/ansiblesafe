package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"time"

	"path/filepath"

	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/vault/api"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
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
}

func findAnsibleVault() (string, error) {
	// Look for 'ansible-vault' in the PATH environment variable
	vaultPath, err := exec.LookPath("ansible-vault")
	if err != nil {
		// If the executable is not found, return an error
		return "", fmt.Errorf("ansible-vault not found in PATH. Please install it before using this script")
	}
	return vaultPath, nil
}

func main() {
	// Add the additional location to the PATH environment variable
	vaultPath, err := findAnsibleVault()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	fmt.Printf("ansible-vault found at: %s\n", vaultPath)
	var filePath string
	var choice int

	pflag.StringVarP(&filePath, "file", "f", "", "Path to YAML file (default: $HOME/vault.yml)")
	pflag.IntVarP(&choice, "operation", "o", 0, "Operation to perform (1: encrypt, 2: decrypt, 3: Write secrets to HashiCorp Vault, 4: Read secrets from HashiCorp Vault, 5: skip encrypting/decrypting)")
	pflag.Parse()

	if filePath == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Error getting current user: %s", err)
		}
		filePath = filepath.Join(usr.HomeDir, "vault.yml")
	}
	if choice == 4 {
		vaultAddress := os.Getenv("VAULT_ADDRESS")
		if vaultAddress == "" {
			log.Fatalf("Error: VAULT_ADDRESS environment variable is not set.")
		}
		// Get a token
		vaultToken := os.Getenv("VAULT_TOKEN")
		if vaultToken == "" {
			log.Fatalf("Error: VAULT_TOKEN environment variable is not set.")
		}
		secretPath := os.Getenv("SECRET_PATH")
		if secretPath == "" {
			log.Fatalf("Error: SECRET_PATH environment variable is not set.")
		}

		readSecretsFromVaultAndWriteToFile(filePath, vaultAddress, vaultToken, secretPath)
		os.Exit(0)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		var config Configuration

		var rootCmd = &cobra.Command{
			Use:   "ansiblesafe",
			Short: "Generate a YAML configuration file",
			Long: `ansiblesafe is a simple command line tool for generating a YAML configuration file with common Red Hat credentials.
				
				This tool provides an interactive menu option for the user to input the values for  common Red Hat credentials.`,
			Run: func(cmd *cobra.Command, args []string) {
				notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
				notice("Please enter the following to generate your vault.yml file:")
				fmt.Print("Enter your RHSM username: ")
				fmt.Scanln(&config.RhsmUsername)

				for {
					fmt.Print("Enter your RHSM password : ")
					rhsmPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}
					config.RhsmPassword = string(rhsmPassword)

					fmt.Print("Confirm RHSM password: ")
					confirmPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}

					if config.RhsmPassword == string(confirmPassword) {
						break
					}
					fmt.Println("Passwords do not match. Please try again.")
				}

				notice("See Creating Red Hat Customer Portal Activation Keys https://access.redhat.com/articles/1378093:")
				// Mix up multiple attributes
				//notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
				//notice("Don't forget this...")
				fmt.Print("Enter your RHSM ORG ID: ")
				fmt.Scanln(&config.RhsmOrg)

				fmt.Print("Enter your RHSM activation key: ")
				fmt.Scanln(&config.RhsmActivationKey)

				notice("Enter Admin password for VMs. This password will be used to access the VMs via SSH.")
				for {
					fmt.Print("Enter the admin password to be used to access VMs: ")
					adminPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}
					config.AdminUserPassword = string(adminPassword)

					fmt.Print("Confirm admin password: ")
					confirmPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}

					if config.AdminUserPassword == string(confirmPassword) {
						break
					}
					fmt.Println("Passwords do not match. Please try again.")
				}

				notice("Offline token not found you can find it at: https://access.redhat.com/management/api")
				// Mix up multiple attributes
				//notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
				//notice("Don't forget this...")
				fmt.Print("Enter your Offline Token: ")
				fmt.Scanln(&config.OfflineToken)

				fmt.Print("Would you like to enter an Automation Hub Offline Token? (y/n): ")
				var hub_response string
				fmt.Scanln(&hub_response)

				if strings.ToLower(hub_response) == "y" {
					notice("Automation Hub Offline Token can be found at https://console.redhat.com/ansible/automation-hub/token")
					fmt.Print("Enter Automation Hub Offline Token: ")
					fmt.Scanln(&config.AutomationHubOfflineToken)
				}

				//var pullSecret string

				fmt.Print("Would you like to enter an OpenShift pull secret? (y/n): ")
				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "y" {
					notice("To deploy and OpenShift enviornment enter the pull secret which can be found at: https://cloud.redhat.com/openshift/install/pull-secret")
					fmt.Print("Enter OpenShift pull secret: ")
					fmt.Scanln(&config.OpenShiftPullSecret)
				}

				fmt.Print("Would you like to enter an FreeIPA password? (y/n): ")
				var ipa_response string
				fmt.Scanln(&ipa_response)

				if strings.ToLower(response) == "y" {
					for {
						fmt.Print("Enter the admin password to be used to for FreeIPA: ")
						freeipa_adminPassword, err := gopass.GetPasswdMasked()
						if err != nil {
							log.Fatalf("Error reading password: %s", err)
						}
						config.FreeIpaServerPassword = string(freeipa_adminPassword)

						fmt.Print("Confirm FreeIPA admin password: ")
						confirm_freeipa_Password, err := gopass.GetPasswdMasked()
						if err != nil {
							log.Fatalf("Error reading password: %s", err)
						}

						if config.FreeIpaServerPassword == string(confirm_freeipa_Password) {
							break
						}
						fmt.Println("Passwords do not match. Please try again.")
					}

				}

				configData, err := yaml.Marshal(config)
				if err != nil {
					log.Fatalf("Error marshaling YAML data: %s", err)
				}
				err = ioutil.WriteFile(filePath, configData, 0644)
				if err != nil {
					log.Fatalf("Error writing vault file: %s", err)
				}

				fmt.Println("Configuration file generated successfully.")
			},
		}

		rootCmd.Execute()
	}
	if choice == 0 {
		fmt.Println("1. Encrypt vault.yml file")
		fmt.Println("2. Decrypt vault.yml file")
		fmt.Println("3. Write secrets to HashiCorp Vault")
		fmt.Println("4. Read secrets from HashiCorp Vault")
		fmt.Println("5. Skip file encryption/decryption")
		notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
		notice("Please choose an option: ")

		fmt.Scanln(&choice)
	}

	var password string
	var vaultpassword string
	var vaultPathMissing bool
	if vaultPath == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Error getting current user: %s", err)
		}
		vaultPath = filepath.Join(usr.HomeDir, ".vault_password")
		if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
			// The file path does not exist
			fmt.Printf("File path %s does not exist\n", vaultPath)
			vaultPathMissing = true
		} else if err != nil {
			// There was an error checking the file path
			fmt.Printf("Error checking file path: %s\n", err.Error())
			vaultPathMissing = true
		} else {
			// The file path exists
			fmt.Printf("File path %s exists\n", vaultPath)
			if _, err := os.Stat(vaultPath); err == nil {
				data, err := ioutil.ReadFile(vaultPath)
				if err != nil {
					fmt.Printf("Error reading file: %s\n", err.Error())
					os.Exit(1)
				}
				vaultpassword = string(data)
				password = string(vaultpassword)
			}
		}
	}

	if choice == 1 || choice == 2 {
		if vaultPathMissing == true {
			fmt.Print("Please enter the vault password: ")

			vaultpassword, err := gopass.GetPasswdMasked()
			if err != nil {
				log.Fatalf("Error reading password: %s", err)
			}
			password = string(vaultpassword)
		}
	}
	var vaultCommand string
	if choice == 1 {
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading file %s: %s", filePath, err)
		}
		if !strings.Contains(string(fileBytes), "ANSIBLE_VAULT") {
			if vaultPathMissing == true {
				fmt.Println("Encrypting file..." + password)
				vaultCommand = fmt.Sprintf("ansible-vault encrypt %s --vault-password-file=<(echo %q)", filePath, password)
			} else {
				vaultCommand = fmt.Sprintf("ansible-vault encrypt %s --vault-password-file=%s --encrypt-vault-id default", filePath, vaultPath)
			}
		} else {
			log.Fatalf("Error: %s is already encrypted.", filePath)
		}
	} else if choice == 2 {
		if vaultPathMissing == true {
			vaultCommand = fmt.Sprintf("ansible-vault decrypt %s --vault-password-file=<(echo %q)", filePath, password)
		} else {
			vaultCommand = fmt.Sprintf("ansible-vault decrypt %s --vault-password-file=%s", filePath, vaultPath)
		}
	} else if choice == 3 {
		// Write secrets to HashiCorp Vault
		configData, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading file %s: %s", filePath, err)
		}

		var config Configuration
		err = yaml.Unmarshal(configData, &config)
		if err != nil {
			log.Fatalf("Error unmarshalling YAML data: %s", err)
		}

		// Initialize a Vault client

		vaultConfig := api.DefaultConfig()
		vaultAddress := os.Getenv("VAULT_ADDRESS")
		if vaultAddress == "" {
			log.Fatalf("Error: VAULT_ADDRESS environment variable is not set.")
		}
		vaultConfig.Address = vaultAddress
		vaultConfig.Timeout = 10 * time.Second
		vaultClient, err := api.NewClient(vaultConfig)
		if err != nil {
			log.Fatalf("Error creating Vault client: %s", err)
		}

		// Get a token
		vaultToken := os.Getenv("VAULT_TOKEN")
		if vaultToken == "" {
			log.Fatalf("Error: VAULT_TOKEN environment variable is not set.")
		}

		// Authenticate with the token
		vaultClient.SetToken(vaultToken)

		// Write secrets to Vault
		secretData := make(map[string]interface{})
		secretData["rhsm_username"] = config.RhsmUsername
		secretData["rhsm_password"] = config.RhsmPassword
		secretData["rhsm_org"] = config.RhsmOrg
		secretData["rhsm_activationkey"] = config.RhsmActivationKey
		secretData["admin_user_password"] = config.AdminUserPassword
		secretData["offline_token"] = config.OfflineToken
		secretData["automation_hub_offline_token"] = config.AutomationHubOfflineToken
		secretData["openshift_pull_secret"] = config.OpenShiftPullSecret
		secretData["freeipa_server_admin_password"] = config.FreeIpaServerPassword

		secretPath := os.Getenv("SECRET_PATH")
		if secretPath == "" {
			log.Fatalf("Error: SECRET_PATH environment variable is not set.")
		}
		ctx := context.Background()
		kv2 := vaultClient.KVv2("ansiblesafe")

		_, err = kv2.Put(ctx, secretPath, secretData)
		if err != nil {
			log.Fatalf("Error writing secret to Vault: %s", err)
		}

		fmt.Println("Secrets written to Vault successfully.")
	} else if choice == 4 {
		vaultAddress := os.Getenv("VAULT_ADDRESS")
		if vaultAddress == "" {
			log.Fatalf("Error: VAULT_ADDRESS environment variable is not set.")
		}
		// Get a token
		vaultToken := os.Getenv("VAULT_TOKEN")
		if vaultToken == "" {
			log.Fatalf("Error: VAULT_TOKEN environment variable is not set.")
		}
		secretPath := os.Getenv("SECRET_PATH")
		if secretPath == "" {
			log.Fatalf("Error: SECRET_PATH environment variable is not set.")
		}
		readSecretsFromVaultAndWriteToFile(filePath, vaultAddress, vaultToken, secretPath)
	} else if choice == 5 {
		notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
		notice("Skipping file encryption.")
	} else {
		log.Fatalf("Invalid choice: %d", choice)
	}

	cmd := exec.Command("bash", "-c", vaultCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing vault command: %s", err)
		log.Printf("Command: %s", vaultCommand)
		log.Printf("Output: %s", string(output))
		log.Fatalf("Failed to execute vault command")
	}

	fmt.Println(string(output))

}

func readSecretsFromVaultAndWriteToFile(filePath, vaultAddress, vaultToken, secretPath string) error {
	// Read secrets from HashiCorp Vault
	fmt.Println("Current file path is: " + filePath)
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = vaultAddress
	vaultConfig.Timeout = 10 * time.Second
	vaultClient, err := api.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("Error creating Vault client: %w", err)
	}

	// Authenticate with the token
	vaultClient.SetToken(vaultToken)

	// Read secrets from Vault
	ctx := context.Background()
	kv2 := vaultClient.KVv2("ansiblesafe")

	secret, err := kv2.Get(ctx, secretPath)
	if err != nil {
		log.Fatal("Error retrieving secret from Vault: %w", err)
	}

	if secret == nil {
		log.Fatal("Error: Secret not found at path " + secretPath)
	}

	// Extract secrets from the retrieved data
	config := Configuration{
		RhsmUsername:          secret.Data["rhsm_username"].(string),
		RhsmPassword:          secret.Data["rhsm_password"].(string),
		RhsmOrg:               secret.Data["rhsm_org"].(string),
		RhsmActivationKey:     secret.Data["rhsm_activationkey"].(string),
		AdminUserPassword:     secret.Data["admin_user_password"].(string),
		OfflineToken:          secret.Data["offline_token"].(string),
		OpenShiftPullSecret:   secret.Data["openshift_pull_secret"].(string),
		FreeIpaServerPassword: secret.Data["freeipa_server_admin_password"].(string),
	}

	// Marshal secrets to YAML data
	configData, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal("Error marshaling YAML data: %w", err)
	}

	// Write secrets to YAML file
	err = ioutil.WriteFile(filePath, configData, 0644)
	if err != nil {
		log.Fatal("Error writing vault file: %w", err)
	}
	fmt.Println("Secrets read from Vault and written to file successfully.")

	return nil
}
