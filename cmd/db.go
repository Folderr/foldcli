/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var verbose, noCleanup bool

var FolderrDbInsertedId *mongo.InsertOneResult

// folderrDBCmd represents the folderr command
var folderrDBCmd = &cobra.Command{
	Use:   "db [db_name] (path_for_private_key)",
	Short: "Set up Folderr DB & Keys",
	Long: `Set up Folderr's database structures and security (encryption) keys
Returns the private key in a file AND as output
db_name is the name of the database you'll use for your Folderr install
path is where the keys get saved. Default: $HOME/.folderr/cli/

NOTES:
Does not have dry-run mode. Cannot accurately test with a dry run mode.
Test with "test" env variable. Do not use production database name/url when testing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := ReadConfig()
		if err != nil {
			println("Failed to read config. see below")
			return err
		}
		if dry {
			return fmt.Errorf(`command \"` + rootCmdName + ` setup db\" does not have dry-run mode
Run with test env var for automatic cleanup of files and database entries`)
		}
		if len(args) < 1 {
			return fmt.Errorf("provide db-name argument. \"db-name\" is the name of the database you'll use for your Folderr install")
		}

		save_dir := ConfigDir
		println(save_dir)
		if len(args) >= 2 {
			save_dir = args[1]
		} else {
			if verbose {
				println("Using default config dir", save_dir, "to save keys in")
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

		if os.Getenv("test") == "true" && !noCleanup {
			defer cleanupFolderrDbCmd(args[0], save_dir)
		}

		db := client.Database(args[0])
		coll := db.Collection("folderrs")
		fldrr := coll.FindOne(context.TODO(), bson.D{})
		if fldrr.Err() != mongo.ErrNoDocuments {
			if fldrr.Err() != nil {
				panic(err)
			}
			println("Folderr appears to be setup")
			return nil
		}
		// generate keys
		// this is for private keys
		println(save_dir)
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}
		if privateKey.Validate() != nil {
			return privateKey.Validate()
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return err
		}
		privBlock := pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   privBytes,
		}

		privatePem := pem.EncodeToMemory(&privBlock)

		if verbose {
			println("Saving private key to", save_dir+"/privateJWT.pem")
		}
		// write private key
		err = os.WriteFile(save_dir+"/privateJWT.pem", privatePem, 0700)
		if err != nil {
			return err
		}
		if verbose {
			println("Saved private key to", save_dir+"/privateJWT.pem")
		}

		// do public key fuckery
		pubBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

		pubBlock := pem.Block{
			Type:    "RSA PUBLIC KEY",
			Headers: nil,
			Bytes:   pubBytes,
		}

		publicPem := pem.EncodeToMemory(&pubBlock)
		if verbose {
			fmt.Println("Saving public key to", save_dir+"/publicJWT.pem", "in case anything goes wrong")
		}
		// write public key in case something goes wrong
		err = os.WriteFile(save_dir+"/publicJWT.pem", publicPem, 0755)
		if err != nil {
			return err
		}
		if verbose {
			fmt.Println("Saved public key to", save_dir+"/publicJWT.pem", "in case anything goes wrong")
		}
		FolderrDbInsertedId, err = coll.InsertOne(context.TODO(), bson.D{
			{Key: "bans", Value: []string{}},
			{Key: "publicKeyJWT", Value: publicPem},
		})
		if err != nil {
			panic(err)
		}

		formattedKey := string(privatePem)
		println(strings.TrimSpace(formattedKey))
		return nil
	},
}

func cleanupFolderrDbCmd(dbName, path string) {
	// make a new database connection
	uri := os.Getenv("MONGO_URI")

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		println("Cleanup failed. See panic.")
		panic(err)
	}

	defer func() {
		if client.Disconnect(context.TODO()) != nil {
			panic(err)
		}
	}()

	cleaned := []string{}

	// Database entry deleted first because if the db entry is not deleted first, why delete the local keys?
	if FolderrDbInsertedId != nil {
		coll := client.Database(dbName).Collection("folderrs")
		_, err = coll.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: FolderrDbInsertedId.InsertedID}})
		if err != nil {
			fmt.Println("Error occured while cleaning up DB, see below")
			panic(err)
		}
		cleaned = append(cleaned, "database")
	}

	// remove files now, thanks
	_, err = os.Stat(path + "/publicJWT.pem")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	} else if !os.IsNotExist(err) {
		err = os.Remove(path + "/publicJWT.pem")
		if err != nil {
			println("Error occured while cleaning up public key, see below")
			panic(err)
		}
		cleaned = append(cleaned, "public key")
	}

	// remove files now, thanks
	_, err = os.Stat(path + "/privateJWT.pem")
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	} else if !os.IsNotExist(err) {
		err = os.Remove(path + "/privateJWT.pem")
		if err != nil {
			println("Error occured while cleaning up private key, see below")
			panic(err)
		}
		cleaned = append(cleaned, "private key")
	}

	if verbose {
		println("Cleaned up", strings.Join(cleaned, ", "))
	}
}

func init() {
	folderrDBCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows information aside from key output.")
	folderrDBCmd.Flags().BoolVar(&noCleanup, "no-cleanup", false, "Does not cleanup if running in test mode. Only useful for data peekers and developers.")
	setupCmd.AddCommand(folderrDBCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// folderrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// folderrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
