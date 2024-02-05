package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/spf13/cobra"
)

var keygenOverride = false

var keygenCmd = &cobra.Command{
	Use:     "keygen <path_for_private_key> <path_for_public_key>",
	Short:   "Generate a private/public keypair according to Folderr's standards",
	Example: "foldcli keygen /home/fldrr/keys/private.pem /home/fldrr/keys/public.pem",
	RunE: func(cmd *cobra.Command, args []string) error {
		privKey, pubKey, err := utilities.GenKeys()
		if err != nil {
			return err
		}

		err = os.WriteFile(args[0], privKey, 0600)
		if err != nil {
			return err
		}
		err = os.WriteFile(args[1], pubKey, 0600)
		if err != nil {
			return err
		}
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("please provide the private key and public key paths")
		}
		pubDir := filepath.Dir(args[1])
		privDir := filepath.Dir(args[0])

		_, err := os.Stat(pubDir)
		if os.IsNotExist(err) {
			return errors.New("public key directory doesn't exist. Please make it")
		}
		if err != nil {
			return err
		}
		_, err = os.Stat(privDir)
		if os.IsNotExist(err) {
			return errors.New("private key directory doesn't exist. Please make it")
		}
		if err != nil {
			return err
		}

		// Ensure the keys are output as pem

		pubExt := filepath.Ext(args[1])
		privExt := filepath.Ext(args[0])
		extErrs := []string{}
		if pubExt != ".pem" {
			extErrs = append(extErrs, "public key extension incorrect. set it to \"pem\", i.e public.pem")
		}
		if privExt != ".pem" {
			extErrs = append(extErrs, "Error: private key extension incorrect. set it to \"pem\", i.e private.pem")
		}

		if len(extErrs) > 0 {
			return errors.New(strings.Join(extErrs, "\n"))
		}

		// Let's just ensure that these keys don't already exist

		keysExist := []string{}
		_, err = os.Stat(args[0])
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if !os.IsNotExist(err) && !keygenOverride {
			keysExist = append(keysExist, "Private key exists, please delete it or choose a different name")
		}
		_, err = os.Stat(args[1])
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if !os.IsNotExist(err) && !keygenOverride {
			keysExist = append(keysExist, "Error: Public key exists, please delete it or choose a different name")
		}

		if len(keysExist) > 0 {
			return fmt.Errorf(strings.Join(keysExist, "\n"))
		}
		return nil
	},
}

func init() {
	keygenCmd.Flags().BoolVarP(&keygenOverride, "force", "f", false, "Override the current keys, if they eixst")
	rootCmd.AddCommand(keygenCmd)
}
