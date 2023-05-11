/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var verbose bool

// folderrCmd represents the folderr command
var folderrCmd = &cobra.Command{
	Use:   "folderr-db [db_name] [path_for_private_key]",
	Short: "Set up Folderr",
	Long: `Set up Folderr's database structures and keys
Returns the private key in a file AND as output
db_name is the name of the database you'll use for your Folderr install
path is where the private key gets saved. Default: $HOME/.folderr/cli/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("provide db-name argument. \"db-name\" is the name of the database you'll use for your Folderr install")
		}

		save_dir, err := getConfigDir()
		if err != nil {
			panic(err)
		}
		if len(args) >= 2 {
			save_dir = args[1]
		} else {
			if verbose {
				fmt.Println("Using default config dir", save_dir, "to save private key in")
			}
		}
		uri := os.Getenv("MONGO_URI")
		if uri == "" {
			return fmt.Errorf("set environment variable \"MONGO_URI\" before running this command. thanks")
		}

		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			panic(err)
		}

		defer func() {
			if client.Disconnect(context.TODO()) != nil {
				panic(err)
			}
		}()

		db := client.Database(args[0])
		coll := db.Collection("folderrs")
		fldrr := coll.FindOne(context.TODO(), bson.D{})
		if fldrr.Err() != mongo.ErrNoDocuments {
			if fldrr.Err() != nil {
				panic(err)
			}
			fmt.Println("Folderr appears to be setup")
			return nil
		}
		// generate keys
		// this is for private keys
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		if privateKey.Validate() != nil {
			panic(privateKey.Validate())
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
		privBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   privBytes,
		}

		privatePem := pem.EncodeToMemory(&privBlock)

		// write private key
		err = ioutil.WriteFile(save_dir+"privateJWT.pem", privatePem, 0700)
		if err != nil {
			panic(err)
		}
		if verbose {
			fmt.Println("Saving private key to", save_dir+"privateJWT.pem")
		}

		// do public key fuckery
		pubBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

		pubBlock := pem.Block{
			Type:    "RSA PUBLIC KEY",
			Headers: nil,
			Bytes:   pubBytes,
		}

		publicPem := pem.EncodeToMemory(&pubBlock)
		// write public key in case something goes wrong
		err = ioutil.WriteFile(save_dir+"publicJWT.pem", publicPem, 0755)
		if err != nil {
			panic(err)
		}
		if verbose {
			fmt.Println("Saving public key to", save_dir+"publicJWT.pem", "in case anything goes wrong")
		}
		_, err = coll.InsertOne(context.TODO(), bson.D{
			{Key: "bans", Value: []string{}},
			{Key: "publicKeyJWT", Value: publicPem},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(string(privatePem))
		return nil
	},
}

func init() {
	setupCmd.AddCommand(folderrCmd)

	folderrCmd.Flags().BoolVarP(&verbose, "verbose", "l", false, "Shows information aside from key output.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// folderrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// folderrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
