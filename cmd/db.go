/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var verbose, noCleanup bool

var FolderrDbInsertedId *mongo.InsertOneResult

// folderrDBCmd represents the folderr command
var folderrDBCmd = &cobra.Command{
	Use:   "db (path_for_private_key)",
	Short: "Set up Folderr DB & Keys",
	Long: `Set up Folderr's database structures and security (encryption) keys
Returns the private key in a file AND as output
db_name is the name of the database you'll use for your Folderr install
path is where the keys get saved. Default: $HOME/.folderr/cli/

NOTES:
Does not have dry-run mode. Cannot accurately test with a dry run mode.
Test with "test" env variable. Do not use production database name/url when testing.
REQUIRES Folderr to be installed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := utilities.GetConfigDir(dry)
		if err != nil {
			panic(err)
		}
		_, config, _, err := utilities.ReadConfig(dir, dry)
		if err != nil {
			cmd.Println("Failed to read config. see below")
			return err
		}
		if dry {
			return fmt.Errorf(`command \"` + rootCmdName + ` setup db\" does not have dry-run mode
Run with test env var for automatic cleanup of files and database entries`)
		}

		checkInit := utilities.CheckInitialization(&config)
		if !checkInit.Folderr {
			cmd.Println("Run \"" + utilities.Constants.RootCmdName + " init folderr\" before running this command. thanks")
			return nil
		}

		isFolderrInstalled, err := utilities.IsFolderrInstalled(config.Directory)
		if err != nil {
			return err
		}
		if !isFolderrInstalled {
			cmd.Println("Please install Folderr. Here's a command to do so.")
			return nil
		}
		uri := os.Getenv(utilities.Constants.EnvPrefix + "MONGO_URI")
		if (config.Database.Url == "" && uri != "") || config.Database.DbName == "" {
			cmd.Println("run \"" + utilities.Constants.RootCmdName + " init db\" before this command. thanks")
			initDbCmd, _, err := cmd.Root().Find([]string{"init db"})
			if err != nil {
				return err
			}
			return initDbCmd.Help()
		}
		if uri == "" {
			uri = config.Database.Url
		}
		save_dir := filepath.Join(dir, "/keys")
		if len(args) >= 1 {
			save_dir = args[0]
		} else {
			if verbose {
				cmd.Println("Using default config dir", save_dir, "to save keys in")
			}
		}

		if _, err = os.Stat(save_dir); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				os.MkdirAll(save_dir, 0700)
			} else {
				return err
			}
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
			defer cleanupFolderrDbCmd(cmd.OutOrStdout(), config, args[0], save_dir)
		}

		db := client.Database(config.Database.DbName)
		coll := db.Collection("folderrs")
		fldrr := coll.FindOne(context.TODO(), bson.D{})
		err = fldrr.Err()
		if err != mongo.ErrNoDocuments {
			if mongo.IsTimeout(err) {
				cmd.Println("Server Timeout Error:", err.Error(), "\nThis can mean that the server is offline, you're offline, or there is (at least) a firewall in the way")
				return nil
			} else if mongo.IsNetworkError(err) {
				cmd.Println("Network Error:", err.Error())
				return nil
			} else if err != nil {
				if strings.Contains(err.Error(), "Unauthorized") || strings.Contains(err.Error(), "unauthorized") {
					println(
						"Authorization error. Please provide authentication information in the string, before the host.\n",
						"Alternatively the error could mean you do not have permissions on this database.\n",
						"Error:",
						err.Error(),
					)
					return nil
				}
				cmd.Println("Encountered error while uploading your user data")
				cmd.Println("Please submit issue with template \"bug report\" at https://github.com/Folderr/folderr-cli/issues with the error below")
				cmd.Println(err)
			}
			cmd.Println("Folderr appears to be setup")
			return nil
		}

		privatePem, publicPem, err := utilities.GenKeys()
		if err != nil {
			return err
		}

		if verbose {
			cmd.Println("Saving private key to", save_dir+"/privateJWT.pem")
		}
		// write private key
		err = os.WriteFile(save_dir+"/privateJWT.pem", privatePem, 0700)
		if err != nil {
			return err
		}
		if verbose {
			cmd.Println("Saved private key to", save_dir+"/privateJWT.pem")
		}

		if verbose {
			cmd.Println("Saving public key to", save_dir+"/publicJWT.pem", "in case anything goes wrong")
		}
		// write public key in case something goes wrong
		err = os.WriteFile(save_dir+"/publicJWT.pem", publicPem, 0755)
		if err != nil {
			return err
		}
		if verbose {
			cmd.Print("Saved public key to", save_dir+"/publicJWT.pem", "in case anything goes wrong\n\n")
		}
		cmd.Println("The keys were saved in", save_dir, "under 'privateJWT.pem' and 'publicJWT.pem'")
		err = saveKeyToFolderr(save_dir, config, privatePem)
		if err != nil {
			cmd.Println(err.Error())
		} else {
			cmd.Println("Installed key to Folderr")
		}
		FolderrDbInsertedId, err = coll.InsertOne(context.TODO(), bson.D{
			{Key: "bans", Value: []string{}},
			{Key: "publicKeyJWT", Value: publicPem},
		})
		if err != nil {
			panic(err)
		} else {
			cmd.Println("Saved public key to database")
		}

		// formattedKey := string(privatePem)
		// println(strings.TrimSpace(formattedKey))
		return nil
	},
}

type locationJSON struct {
	Keys          string `json:"keys"`
	KeyConfigured bool   `json:"keyConfigured"`
}

func saveKeyToFolderr(save_dir string, config utilities.Config, privateKey []byte) error {
	dir := filepath.Join(config.Directory, "internal/keys")
	privatePath := filepath.Join(dir, "privateJWT.pem")
	_, err := os.Stat(filepath.Join(config.Directory, "internal/keys"))
	locationsPath := filepath.Join(config.Directory, "internal/locations.json")
	example := "{\"keys\": \"internal\", \"keyConfigured\": true}"
	if err != nil {
		return fmt.Errorf(
			`failed to write the private key to Folderr, you need to copy it from "%v" to "%v".
			You also need to write %v to "%v"
			Original error: %w`,
			privatePath,
			save_dir,
			example,
			locationsPath,
			err,
		)
	}
	err = os.WriteFile(privatePath, privateKey, 0600)
	if err != nil {
		return fmt.Errorf(
			`failed to write the private key to Folderr, you need to copy it from "%v" to "%v".
			You also need to write %v to "%v"
			Original error: %w`,
			privatePath,
			save_dir,
			example,
			locationsPath,
			err,
		)
	}

	contents := locationJSON{Keys: "internal", KeyConfigured: true}

	marshal, err := json.Marshal(contents)

	if err != nil {
		return fmt.Errorf(
			"failed to write \"%v\". You must do it yourself\nIt should look like %v\nOriginal error: %w",
			locationsPath,
			example,
			err,
		)
	}

	err = os.WriteFile(locationsPath, marshal, 0600)
	if err != nil {
		return fmt.Errorf(
			"failed to write \"%v\". You must do it yourself\nIt should look like %v\nOriginal error: %w",
			locationsPath,
			example,
			err,
		)
	}
	return nil
}

func cleanupFolderrDbCmd(w io.Writer, config utilities.Config, dbName, path string) {
	// make a new database connection
	uri := config.Database.Url
	fmt.Fprintln(w, path)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		println("DB Connection failed. See panic.")
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
			fmt.Fprintf(w, "Error occured while cleaning up DB, see below\n%v\n", err.Error())
		} else {
			cleaned = append(cleaned, "database")
		}
	}

	// remove files now, thanks
	_, err = os.Stat(filepath.Join(path + "/publicJWT.pem"))
	if err != nil {
		fmt.Fprintf(w, "Couldn't remove the public key\n%v\n", err.Error())
	} else {
		err = os.Remove(filepath.Join(path + "/publicJWT.pem"))
		if err != nil {
			fmt.Fprintf(w, "Error occured while cleaning up public key, see below\n%v\n", err.Error())
		} else {
			cleaned = append(cleaned, "public key")
		}
	}

	// remove files now, thanks
	_, err = os.Stat(filepath.Join(path, "/privateJWT.pem"))
	if err != nil {
		fmt.Fprintf(w, "Couldn't remove the private key\n%v\n", err.Error())
	} else {
		err = os.Remove(filepath.Join(path, "/privateJWT.pem"))
		if err != nil {
			fmt.Fprintf(w, "Error occured while cleaning up private key, see below\n%v\n", err.Error())
		} else {
			cleaned = append(cleaned, "private key")
		}
	}

	_, err = os.Stat(filepath.Join(config.Directory, "internal/keys/privateJWT.pem"))
	if err != nil {
		fmt.Fprintf(w, "Couldn't remove the private key from Folderr, see below\n%v\n", err.Error())
	} else {
		err = os.Remove(filepath.Join(config.Directory, "internal/keys/privateJWT.pem"))
		if err != nil {
			fmt.Fprintf(w, "Error occured while cleaning up private key, see below\n%v\n", err.Error())
		} else {
			cleaned = append(cleaned, "Folderr / keys / private key")
		}
	}

	_, err = os.Stat(filepath.Join(config.Directory, "internal/locations.json"))
	if err != nil {
		fmt.Fprintf(w, "Couldn't reset the locations.json from Folderr, see below\n%v\n", err.Error())
	} else {
		marshal, err := json.Marshal(locationJSON{Keys: "internal", KeyConfigured: false})
		if err != nil {
			fmt.Fprintf(w, "340: Error occured while cleaning up locations.json, see below\n%v\n", err.Error())
		} else {
			err = os.WriteFile(filepath.Join(config.Directory, "internal/locations.json"), marshal, 0600)
			if err != nil {
				fmt.Fprintf(w, "347: Error occured while cleaning up locations.json, see below\n%v\n", err.Error())
			} else {
				cleaned = append(cleaned, "Folderr / keys / locations.json")
			}
		}
	}

	if verbose {
		fmt.Fprintln(w, "Cleaned up", strings.Join(cleaned, ", "))
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
