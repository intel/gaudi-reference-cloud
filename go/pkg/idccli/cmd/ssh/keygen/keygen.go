// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package keygen

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	test_tools "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tools/ssh"
	"github.com/spf13/cobra"
)

type keyGenOpts struct {
	sshDir string
}

func NewCmdKeyGen() *cobra.Command {
	opts := keyGenOpts{}
	keyGenCmd := &cobra.Command{
		Short: "Generate sshkeys if not existed in the provided outputdir",
		Use:   "keygen",
		Example: heredoc.Doc(`
			# command for ssh keygen
			$ idccli ssh keygen --outputdir <output_dir>
			$ idccli ssh keygen -o <output_dir>
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSshKeyGen(opts)
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	keyGenCmd.Flags().StringVarP(&opts.sshDir, "outputdir", "o", sshDir, "output dir to generate ssh keys into")

	return keyGenCmd

}

// writePemToFile writes keys to a file
func writeKeyToFile(key string, saveFileTo string) error {
	keyByte := []byte(key)
	err := os.WriteFile(saveFileTo, keyByte, 0600)
	if err != nil {
		return err
	}
	return nil
}

func runSshKeyGen(opts keyGenOpts) error {
	sshDir := opts.sshDir
	_, err := os.Stat(sshDir)

	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(sshDir, 0700)
			if err != nil {
				return fmt.Errorf("error encountered while creating directory at path %v: %v", sshDir, err)
			}
		} else {
			return fmt.Errorf("error in runSshKeyGen: %v", err)
		}
	}

	publicKeyFile := filepath.Join(sshDir, "id_rsa.pub")
	privateKeyFile := filepath.Join(sshDir, "id_rsa")

	var publicKeyExists bool
	var createKeyPair string

	// Check if the public key file exists
	if _, err := os.Stat(publicKeyFile); err == nil {
		publicKeyExists = true
		fmt.Println("Using private key:", privateKeyFile)
		fmt.Println("Using public key:", publicKeyFile)
	}

	if !publicKeyExists {
		// Ask the user if they want to create a new key pair
		prompt := &survey.Select{
			Message: "SSH RSA key pair not found. Do you want to create a new key pair?",
			Options: []string{"Yes", "No"},
		}

		if err := survey.AskOne(prompt, &createKeyPair); err != nil {
			return fmt.Errorf("error occured when promoting for user's choice, err: %v", err)
		}

		if createKeyPair == "Yes" {
			bitSize := 4096
			// To get the current user's username
			currentUser, err := user.Current()
			if err != nil {
				return err
			}
			username := currentUser.Username
			// To get hostname of the user
			hostname, err := os.Hostname()
			if err != nil {
				return err
			}

			// identifier will append "username@hostname" at the end of public key contents
			identifier := fmt.Sprintf("%s@%s", username, hostname)
			privateKey, publicKey, err := test_tools.CreateSshRsaKeyPair(bitSize, identifier)
			if err != nil {
				return fmt.Errorf("error when crrsting ssh keys: %v", err)
			}

			err = writeKeyToFile(privateKey, privateKeyFile)
			if err != nil {
				return err
			}
			fmt.Println("\nPrivate key is saved to:", privateKeyFile)

			err = writeKeyToFile(publicKey, publicKeyFile)
			if err != nil {
				return err
			}
			fmt.Println("Public key is saved to:", publicKeyFile)

			fmt.Printf("\033[1;32m%s\033[0m\n", "\nSSH RSA key pair generated")

		} else if createKeyPair == "No" {
			fmt.Println("\nNo key pair will be generated. You must create a key pair to create an instance.")
			fmt.Println("Press Enter to exit")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			return nil
		}
	}

	fmt.Printf("\033[1m%s\033[0m\n", "\nPlease find your SSH RSA public key contents below\n")
	publicKeyContent, err := os.ReadFile(publicKeyFile)
	if err != nil {
		return err
	}

	fmt.Printf("\033[37;42m%s\033[0m\n\n", string(publicKeyContent))
	fmt.Printf("\033[1;32m%s\033[0m\n", "Please copy and paste the above contents to the console link below")
	fmt.Printf("\033[1;37m%s\033[0m\n", "https://console.cloud.intel.com/security/publickeys/import\n")
	return nil
}
