/*
Copyright © 2023 Folderr <contact@folderr.net>
*/
package cmd

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Folderr/foldcli/utilities"
	"github.com/fossoreslp/go-uuid-v4"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/argon2"
)

var ownerUsername, ownerPassword, ownerEmail string

type User struct {
	Id                string         `bson:"id"`
	Password          string         `bson:"password"`
	Email             string         `bson:"email"`
	Username          string         `bson:"username"`
	Admin             bool           `bson:"admin"`
	Owner             bool           `bson:"owner"`
	CURLs             []string       `bson:"cURLs"`
	Files             int            `bson:"files"`
	Links             int            `bson:"links"`
	Notifs            []Notification `bson:"notifs"`
	CreatedAt         time.Time      `bson:"createdAt"`
	Privacy           UserPrivacy    `bson:"privacy"`
	MarkedForDeletion bool           `bson:"markedForDeletion"`
}

type Notification struct {
	Id        string    `bson:"id"`
	Title     string    `bson:"title"`
	Notify    string    `bson:"notify"`
	CreatedAt time.Time `bson:"createdAt"`
}

type UserPrivacy struct {
	DataCollection bool `bson:"dataCollection"`
}

// ownerCmd represents the owner command
var ownerCmd = &cobra.Command{
	Use:   "owner",
	Short: "Set up the owner for your Folderr instance",
	Long: `Set's up the owner account on your Folderr instance
Please run "` + utilities.Constants.RootCmdName + ` init db" before running this command`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// prechecks

		dir, err := utilities.GetConfigDir(dry)
		if err != nil {
			panic(err)
		}
		_, config, _, err := utilities.ReadConfig(dir, dry)
		uri := config.Database.Url
		if uri == "" || config.Database.DbName == "" {
			return fmt.Errorf("please run \"" + rootCmdName + " init db\" before running this command. thanks")
		}

		// P: Puncuation unicode
		// M: Mark (Accents)
		// Z: Separator
		regexfail := false
		passwdreg := regexp.MustCompile(`(.*[A-Za-z-_])(.*[\p{M}\p{Z}\p{P}]).{8,256}`)
		usernamereg := regexp.MustCompile(`\w{3,16}`)
		emailreg := regexp.MustCompile(`([\w.\-$%#!+/=^;&'*]{2,})?@[a-z\d$-_.+!*’(,;:@&=/]{2,}\.[a-z]{2,}(.[a-z]{2,})?`)
		// check if password is valid
		if len(passwdreg.FindString(ownerPassword)) != len(ownerPassword) {
			// TODO: Show password requirements
			fmt.Println(`Password is invalid
Password must be between 8 and 256 characters
Contain one special character
Contain one number
And contain a upper or lowercase A-Z letter`)
			regexfail = true
		}

		// check if username is valid
		if len(usernamereg.FindString(ownerUsername)) != len(ownerUsername) {
			fmt.Println(`Username is invalid
Username must include at least 3 characters of which all uppercase/lowercase alphanumeric numbers are allowed, and so is an underscore
Max size: 16 characters`)
			regexfail = true
		}

		// check if email is valid
		if len(emailreg.FindString(ownerEmail)) != len(ownerEmail) {
			fmt.Println(`Email is invalid. Standard email names are allowed.
If you believe this to be a bug please submit an issue at https://github.com/Folderr/folderr-cli/issues under "bug report"`)
			regexfail = true
		}

		if regexfail {
			os.Exit(1)
		}

		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetAppName("Folderr CLI"))
		if err != nil {
			panic(err)
		}

		defer func() {
			if client.Disconnect(context.TODO()) != nil {
				panic(err)
			}
		}()

		coll := client.Database(config.Database.DbName).Collection("users")

		var preUser User
		coll.FindOne(context.TODO(), bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{
					{Key: "username", Value: ownerUsername},
				},
				bson.D{
					{Key: "email", Value: ownerEmail},
				},
			},
			},
		}).Decode(&preUser)

		if preUser.Username == ownerUsername {
			fmt.Println("Username taken")
			os.Exit(1)
		}

		if preUser.Email == ownerEmail {
			fmt.Println("Email taken")
			os.Exit(1)
		}

		hashed, err := hashPassword(ownerPassword)
		if err != nil {
			fmt.Println("Encountered error while hashing password.")
			fmt.Println("If this is a bug please submit an issue at https://github.com/Folderr/folderr-cli/issues under \"bug report\"")
			fmt.Println("See error below")
			fmt.Println(err)
			os.Exit(1)
		}

		uid, err := uuid.NewString()

		if err != nil {
			fmt.Println("Encountered error while generating user ID")
			fmt.Println("Please submit issue with template \"bug report\" at https://github.com/Folderr/folderr-cli/issues and error below")
			fmt.Println(err)
			os.Exit(1)
		}

		ownerUser := User{
			Id:                uid,
			Username:          ownerUsername,
			Email:             ownerEmail,
			Password:          hashed,
			CreatedAt:         time.Now(),
			Owner:             true,
			Admin:             true,
			MarkedForDeletion: false,
		}

		_, err = coll.InsertOne(context.TODO(), ownerUser)
		if mongo.IsTimeout(err) {
			println("Server Timeout Error:", err.Error(), "\nThis can mean that the server is offline, you're offline, or there is (at least) a firewall in the way")
			return nil
		} else if mongo.IsNetworkError(err) {
			println("Network Error:", err.Error())
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
			fmt.Println("Encountered error while uploading your user data")
			fmt.Println("Please submit issue with template \"bug report\" at https://github.com/Folderr/folderr-cli/issues with the error below")
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("Generated owner account. See info below.")
		fmt.Println("Account ID:", ownerUser.Id)
		fmt.Println("Password:", ownerPassword)
		fmt.Println("Email:", ownerUser.Email)
		fmt.Println("Username:", ownerUser.Username)
		fmt.Println("Created At:", ownerUser.CreatedAt.Format("Monday January _2 2006 15:04:05"))
		return nil
	},
}

type params struct {
	memory     uint32
	time       uint32
	threads    uint8
	saltLength uint32
	keyLength  uint32
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func hashPassword(password string) (hash string, err error) {
	p := &params{
		memory:     64 * 1024,
		time:       10,
		keyLength:  32,
		threads:    4,
		saltLength: 16,
	}
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(key)

	// Return a string using the standard encoded hash representation.
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.time, p.threads, b64Salt, b64Hash)

	return encodedHash, nil
}

func init() {
	setupCmd.AddCommand(ownerCmd)
	ownerCmd.Flags().StringVarP(&ownerUsername, "username", "u", "", "Set's the username of the owner account")
	ownerCmd.MarkFlagRequired("username")
	ownerCmd.Flags().StringVarP(&ownerPassword, "password", "p", "", "Set's the password of the owner account")
	ownerCmd.MarkFlagRequired("password")
	ownerCmd.Flags().StringVarP(&ownerEmail, "email", "e", "", "Set's the email of the owner account")
	ownerCmd.MarkFlagRequired("email")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ownerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ownerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
