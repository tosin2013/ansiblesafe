package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"

	"path/filepath"

	"strings"

	"github.com/fatih/color"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	RhsmUsername      string `yaml:"rhsm_username"`
	RhsmPassword      string `yaml:"rhsm_password"`
	RhsmOrg           string `yaml:"rhsm_org"`
	RhsmActivationKey string `yaml:"rhsm_activationkey"`
	AdminUserPassword string `yaml:"admin_user_password"`
	OfflineToken      string `yaml:"offline_token"`
}

func main() {
	_, err := exec.LookPath("ansible-vault")
	if err != nil {
		log.Fatalf("Error: ansible-vault CLI is not installed. Please install it before using this script.")
	}
	var filePath string

	var choice int

	pflag.StringVarP(&filePath, "file", "f", "", "Path to YAML file (default: $HOME/vault.yml)")
	pflag.IntVarP(&choice, "operation", "o", 0, "Operation to perform (1: encrypt, 2: decrypt)")
	pflag.Parse()

	if filePath == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Error getting current user: %s", err)
		}
		filePath = filepath.Join(usr.HomeDir, "vault.yml")
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		var config Configuration

		var rootCmd = &cobra.Command{
			Use:   "generate-config",
			Short: "Generate a YAML configuration file",
			Long: `generate-config is a simple command line tool for generating a YAML configuration file with specified values.
				
				This tool provides an interactive menu option for the user to input the values for the configuration.`,
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

				for {
					fmt.Print("RHSM password: ")
					adminPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}
					config.AdminUserPassword = string(adminPassword)

					fmt.Print("Confirm RHSM password: ")
					confirmPassword, err := gopass.GetPasswdMasked()
					if err != nil {
						log.Fatalf("Error reading password: %s", err)
					}

					if config.AdminUserPassword == string(confirmPassword) {
						break
					}
					fmt.Println("Passwords do not match. Please try again.")
				}

				notice("Offline token not found you can find it at https://access.redhat.com/management/api:")
				// Mix up multiple attributes
				//notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
				//notice("Don't forget this...")
				fmt.Print("Enter your Offline Token: ")
				fmt.Scanln(&config.OfflineToken)

				configData, err := yaml.Marshal(config)
				if err != nil {
					log.Fatalf("Error marshaling YAML data: %s", err)
				}
				err = ioutil.WriteFile("vault.yml", configData, 0644)
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
		notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
		notice("Please choose an option: ")

		fmt.Scanln(&choice)
	}

	var password string
	fmt.Print("Please enter the vault password: ")

	vaultpassword, err := gopass.GetPasswdMasked()
	if err != nil {
		log.Fatalf("Error reading password: %s", err)
	}
	password = string(vaultpassword)

	var vaultCommand string
	if choice == 1 {
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading file %s: %s", filePath, err)
		}
		if !strings.Contains(string(fileBytes), "ANSIBLE_VAULT") {
			vaultCommand = fmt.Sprintf("ansible-vault encrypt %s --vault-password-file=<(echo %q)", filePath, password)
		} else {
			log.Fatalf("Error: %s is already encrypted.", filePath)
		}
	} else if choice == 2 {
		vaultCommand = fmt.Sprintf("ansible-vault decrypt %s --vault-password-file=<(echo %q)", filePath, password)
	} else {
		log.Fatalf("Invalid choice: %d", choice)
	}

	cmd := exec.Command("bash", "-c", vaultCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing vault command: %s", err)
	}

	fmt.Println(string(output))

}
