package init

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Folderr/foldcli/utilities"
	"github.com/spf13/cobra"
)

func init() {
	dbInitCmd.Flags().BoolVarP(&override, "override", "o", false, "Override previous settings")
	initCmd.AddCommand(dbInitCmd)
}

var dbInitCmd = &cobra.Command{
	Use:     "db <database uri> <database name>",
	Short:   "Initalize config for database related commands",
	Long:    `Initalize config for database related commands`,
	Aliases: []string{"database"},
	Example: "  " + utilities.Constants.RootCmdName + " " + strings.Split(initCmd.Use, " ")[0] + " db mongodb://localhost/folderrV2?tls=true folderr\n  " +
		utilities.Constants.RootCmdName + " " + strings.Split(initCmd.Use, " ")[0] + " db mongodb://username:password@127.0.0.1/folderrV2 folderr",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		var directive cobra.ShellCompDirective
		if len(args) == 0 {
			comps = cobra.AppendActiveHelp(comps, "Connection url to your database")
			directive = cobra.ShellCompDirectiveDefault
		} else if len(args) == 1 {
			comps = cobra.AppendActiveHelp(comps, "The name of the database you wish to use")
			directive = cobra.ShellCompDirectiveDefault
		} else {
			comps = cobra.AppendActiveHelp(comps, "ERROR: You have provided too many arguments")
			directive = cobra.ShellCompDirectiveNoFileComp
		}
		return comps, directive
	},
	// Args: cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	RunE: func(command *cobra.Command, args []string) error {
		dryRun, err := command.Flags().GetBool("dry")
		if err != nil {
			return fmt.Errorf("Unexpected Error" + err.Error())
		}
		if len(args) == 0 {
			return fmt.Errorf("need the connection url and the database name. expected 2 arguments, got 0")
		} else if len(args) == 1 {
			return fmt.Errorf("need the database name. expected 2 arguments, got 1")
		}
		dir, err := utilities.GetConfigDir(dry)
		if err != nil {
			return err
		}
		vip, config, _, err := utilities.ReadConfig(dir, dry)
		if err != nil {
			panic(err)
		}

		if config.Database.Url != "" && !override {
			command.Println("Your database information is already saved. To overwrite it pass the `override` flag")
			return nil
		} else if config.Database.DbName != "" && !override {
			command.Println("Your database information is already saved. To overwrite it pass the `override` flag")
			return nil
		}
		uri, err := url.Parse(args[0])
		if err != nil {
			return fmt.Errorf("Failed to parse the database URI you gave.\nError: " + err.Error())
		} else if uri.Scheme != "mongodb" || uri.Host == "" || uri.Path == "" {
			invalidURIStr := []string{
				"The URI you gave is not in the format needed.\n",
				"  Format needed is \"mongodb://[user:password@]$host/$databaseAndExtras\"",
				"  Replace \"$host\" with the database host and \"$databaseAndExtras\" with the database and other connection options\n",
				"  If you want to supply a username & password please format it like \"username:password@\" and place it before the host",
				"  Example URI (with user): mongodb://example:1234@127.0.0.1/folderrV2",
				"  Example URI (without user): mongodb://127.0.0.1/folderrV2\n",
			}
			return fmt.Errorf(strings.Join(invalidURIStr, "\n"))
		}
		if config.Database.DbName != "" || config.Database.Url != "" {
			command.Println("Overwriting database config.")
		}

		// Actually set the database values
		vip.Set("db.url", args[0])
		vip.Set("db.dbName", args[1])
		if dryRun {
			command.Println("Saved database information\nNOTICE: Did NOT save, due to dry run")
			return nil
		}
		err = vip.WriteConfig()
		if err == nil {
			command.Println("Saved database information")
		}
		return err
	},
}
