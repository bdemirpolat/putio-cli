package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/putdotio/go-putio"
	"github.com/vigo/putio-cli/cli/version"
	"golang.org/x/oauth2"
)

var (
	// ErrPutioTokenRequired is an error for required put.io token
	ErrPutioTokenRequired = errors.New("please provide put.io token via flags or PUTIO_TOKEN env-var")

	// ErrPutioTokenInvalid is an error for invalid put.io token
	ErrPutioTokenInvalid = errors.New("invalid put.io token")

	// ErrMissingArgs is an error for missing arguments
	ErrMissingArgs = errors.New("you need to provide args")

	subCommands []*flag.FlagSet

	colorBoldBlue   = color.New(color.Bold, color.FgBlue).SprintFunc()
	colorBoldYellow = color.New(color.Bold, color.FgYellow).SprintFunc()
	colorYellow     = color.New(color.FgYellow).SprintFunc()
	colorRed        = color.New(color.FgRed).SprintFunc()
	colorWhite      = color.New(color.FgWhite).SprintFunc()
	colorBoldWhite  = color.New(color.Bold, color.FgWhite).SprintFunc()

	// FlagColorEnabled hold color output switcher
	FlagColorEnabled *bool

	// FlagVersion holds boolean for displaying version information
	FlagVersion *bool
	// FlagToken holds put.io token, default value is set from PUTIO_TOKEN env-var
	FlagToken *string

	// FlagSetList is a list subcommand
	FlagSetList *flag.FlagSet
	// FlagListShowDownloadURL is a boolean option to display list as download url
	FlagListShowDownloadURL *bool
	// FlagListShowOnlyID is a boolean option to display list as file id
	FlagListShowOnlyID *bool
	// FlagListDeleteByID is a flag for file id to delete before listing
	FlagListDeleteByID *int64

	// FlagSetUpload is a download subcommand
	FlagSetUpload *flag.FlagSet

	// FlagSetMove is a move subcommand
	FlagSetMove *flag.FlagSet

	// FlagSetDelete is a delete subcommand
	FlagSetDelete *flag.FlagSet

	usage = `
usage: putio-cli [-flags] [subcommand] [args]

  You can set your token via PUTIO_TOKEN environment variable! unless
  you need to pass token via -t or -token.

flags:

  -help, -h                           display help
  -version, -v                        display version (%s)
  -color, -c                          enable/disable color (default: disabeld)
  -token, -t                          set put.io token

subcommands:

  list FOLDERID                       list files under given FOLDERID (default: 0 which is root folder)
  list -delete FILEID FOLDERID        first delete given FILEID then list files under given FOLDERID
  list -id FOLDERID                   list files under given FOLDERID as file id
  list -url FOLDERID                  list files under given FOLDERID as downloadable URL

  upload url URL URL URL...           tell put.io to download given URL(s)

  delete FILEID FILEID...             delete given FILEIDs
  move FILEID FILEID... FOLDERID      move given files (FILEIDs) to target folder (FOLDERID)

examples:

  $ putio-cli -t YOURTOKEN list
  $ putio-cli -t YOURTOKEN -c list

  # list files under given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c list FOLDERID

  # first delete given FILEID then list files for given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c list -delete FILEID FOLDERID

  # tell putio to upload given urls
  $ putio-cli -t YOURTOKEN upload url https://www.youtube.com/watch?v=eeiTP69qc9Q https://www.youtube.com/watch?v=SrBUaaNsZzg

  # tell putio to move given FILEIDs to given FOLDERID in color!
  $ putio-cli -t YOURTOKEN -c move FILEID FILEID FILEID... FOLDERID

  # delete files in color!
  $ putio-cli -t YOURTOKEN -c delete FILEID FILEID

  # move files in color!
  $ putio-cli -t YOURTOKEN -c move FILEID FILEID... FOLDERID

`
)

// Application represents application structure
type Application struct {
	In     io.Reader
	Out    io.Writer
	Client *putio.Client
}

// printSubCommandsUsage is a helper function
func printSubCommandsUsage(invalidCMD *string) {
	if invalidCMD != nil {
		fmt.Fprintf(os.Stdout, "invalid subcommand: %s\n\nvalid commands:\n\n", *invalidCMD)
	}
	for _, subCommand := range subCommands {
		fmt.Fprintf(os.Stdout, "  %s\n", subCommand.Name())
		subCommand.PrintDefaults()
	}
	fmt.Fprintf(os.Stdout, "\n")
}

// NewApplication returns new Application instance
func NewApplication() *Application {
	FlagVersion = flag.Bool("version", false, fmt.Sprintf("display version information (%s)", version.Version))
	flag.BoolVar(FlagVersion, "v", false, "")

	FlagColorEnabled = flag.Bool("color", false, "enable/disable color")
	flag.BoolVar(FlagColorEnabled, "c", false, "")

	helpFlagToken := "put.io token"
	FlagToken = flag.String("token", "", helpFlagToken)
	flag.StringVar(FlagToken, "t", "", helpFlagToken+" (short)")

	FlagSetList = flag.NewFlagSet("list", flag.ExitOnError)
	FlagListShowDownloadURL = FlagSetList.Bool("url", false, "display list as download url")
	FlagListShowOnlyID = FlagSetList.Bool("id", false, "display list as file id")
	FlagListDeleteByID = FlagSetList.Int64("delete", -1, "delete given file id")

	FlagSetUpload = flag.NewFlagSet("upload", flag.ExitOnError)
	FlagSetDelete = flag.NewFlagSet("delete", flag.ExitOnError)
	FlagSetMove = flag.NewFlagSet("move", flag.ExitOnError)

	subCommands = []*flag.FlagSet{
		FlagSetList,
		FlagSetUpload,
		FlagSetDelete,
		FlagSetMove,
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, usage, version.Version)
	}

	return &Application{
		In:  os.Stdin,
		Out: os.Stdout,
	}
}

// Validate validates required params
func (a *Application) Validate() error {
	flag.Parse()

	if !*FlagColorEnabled {
		color.NoColor = true
	}

	if *FlagVersion {
		fmt.Fprintln(a.Out, version.Version)
		return nil
	}

	token, ok := os.LookupEnv("PUTIO_TOKEN")
	if ok {
		*FlagToken = token
	}

	if *FlagToken == "" {
		return ErrPutioTokenRequired
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *FlagToken})
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	a.Client = putio.NewClient(oauthClient)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := a.Client.ValidateToken(ctx)
	if err != nil {
		return err
	}
	if userID == nil {
		return ErrPutioTokenInvalid
	}

	if len(flag.Args()) > 0 {
		subcommand := flag.Args()[0]

		switch subcommand {
		case "list":
			_ = FlagSetList.Parse(flag.Args()[1:])
		case "upload":
			_ = FlagSetUpload.Parse(flag.Args()[1:])
		case "delete":
			_ = FlagSetDelete.Parse(flag.Args()[1:])
		case "move":
			_ = FlagSetMove.Parse(flag.Args()[1:])
		default:
			printSubCommandsUsage(&subcommand)
			os.Exit(1)
		}
	}

	return nil
}

// Run executes main application
func (a *Application) Run() error {
	if err := a.Validate(); err != nil {
		return err
	}

	if FlagSetList.Parsed() {
		return a.CommandList()
	}

	if FlagSetDelete.Parsed() {
		return a.CommandDelete()
	}

	if FlagSetUpload.Parsed() {
		return a.CommandUpload()
	}

	if FlagSetMove.Parsed() {
		return a.CommandMove()
	}

	return nil
}

// ByteCountSI returns formatted size
func (a *Application) ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
