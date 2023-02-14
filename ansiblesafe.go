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
	RhsmUsername        string `yaml:"rhsm_username"`
	RhsmPassword        string `yaml:"rhsm_password"`
	RhsmOrg             string `yaml:"rhsm_org"`
	RhsmActivationKey   string `yaml:"rhsm_activationkey"`
	AdminUserPassword   string `yaml:"admin_user_password"`
	OfflineToken        string `yaml:"offline_token"`
	OpenShiftPullSecret string `yaml:"openshift_pull_secret"`
}

func main() {
	_, err := exec.LookPath("ansible-vault")
	if err != nil {
		log.Fatalf("Error: ansible-vault CLI is not installed. Please install it before using this script.")
	}
	var filePath string

	var choice int

	pflag.StringVarP(&filePath, "file", "f", "", "Path to YAML file (default: $HOME/vault.yml)")
	pflag.IntVarP(&choice, "operation", "o", 0, "Operation to perform (1: encrypt, 2: decrypt 3: skip encrypting/decrypting)")
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

				//var pullSecret string

				fmt.Print("Would you like to enter an OpenShift pull secret? (y/n): ")
				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "y" {
					notice("To deploy and OpenShift envioenment enter the pull secret which can be found at: https://cloud.redhat.com/openshift/install/pull-secret")
					fmt.Print("Enter OpenShift pull secret: ")
					fmt.Scanln(&config.OpenShiftPullSecret)
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
		fmt.Println("3. Skip file encryption/decryption")
		notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
		notice("Please choose an option: ")

		fmt.Scanln(&choice)
	}

	var password string
	if choice == 1 || choice == 2 {
		fmt.Print("Please enter the vault password: ")

		vaultpassword, err := gopass.GetPasswdMasked()
		if err != nil {
			log.Fatalf("Error reading password: %s", err)
		}
		password = string(vaultpassword)
	}
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
	} else if choice == 3 {
			notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
			notice("Skipping file encryption.")
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
