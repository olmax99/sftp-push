package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// watchConfig reflects the yaml config file parameters
type watchConfig struct {
	Defaults struct {
		Userpath   string `yaml:"userpath"`
		S3Target   string `yaml:"s3target"`
		Awsprofile string `yaml:"awsprofile"`
		Awsregion  string `yaml:"awsregion"`
	} `yaml:"defaults"`
	Watch struct {
		Users []struct {
			Name    string   `yaml:"name"`
			Sources []string `yaml:"sources"`
		} `yaml:"users"`
	} `yaml:"watch"`
}

// watchConfigOperations contains all methods needed to process input to cmdWatch
// type watchConfigOperations interface {
// 	createWatcher(eops event.FsEventOps, globalCfg *watchConfig) error
// 	checkDir(path string) (bool, error)
// 	unmarshalWatchFlag(flagIn []string, globalCfg *watchConfig) error
// 	newS3Conn(profile *string, region *string) *s3.S3
// }

// watchConfigOps implements the watchConfigOperations interface
type watchConfigOps struct{}

var src []string // watch flag --source read as string

// cmdWatch represents the watch command
var cmdWatch = &cobra.Command{
	Use:   "watch",
	Short: "Start the fsnotify file system event watcher",
	Long: strings.TrimSpace(`
The watch command starts the fsnotify file watcher, and triggers 
event tasks based on WRITE_CLOSE signals.

The --source flag is optional and can overwrite the arguments
provided by a config file.

Examples:

SFTPPUSH_DEFAULTS_USERPATH=/my/user/dir/ sftppush --config config.yaml watch

SFTPPUSH_DEFAULTS_AWSPROFILE=my-profile sftppush watch \
  --source="name=user1,paths=/device1/data /device2/data" \
  --source="name=user2,paths=/device1/data /device2/data"
`),
	// Args: func(cmd *cobra.Command, args []string) error {
	// 	if len(args) < 1 {
	// 		return errors.New("requires a color argument")
	// 	}
	// 	if myapp.IsValidColor(args[0]) {
	// 		return nil
	// 	}
	// 	return fmt.Errorf("invalid color specified: %s", args[0])
	// },
	RunE: func(cmd *cobra.Command, args []string) error {
		w := watchConfigOps{}

		if n := cmd.Flags().NFlag(); n < 1 {
			return errors.New("Use either '--source' flag or '--config'.")
		}

		log.Printf("DEBUG[*] cfgWatch (from config): %q", &gCfg)
		log.Printf("DEBUG[*] cmdWatch (from flag): %s", src)

		// Will overwrite config values if both --config and --sources are set
		if cmd.Flag("source").Changed {
			if err := w.unmarshalWatchFlag(src, &gCfg); err != nil {
				// log.Fatalf("FATAL[*] decodeWatchFlag: %s", err)
				return errors.Wrapf(err, "decodeWatchFlag: %q", src)
			}
		}

		// TODO Catch errors, implement a notification service
		e := event.FsEventOps{}
		if err := w.createWatcher(e, &gCfg); err != nil {
			return errors.Wrap(err, "createWatcher")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdWatch)
	cmdWatch.Flags().StringArrayVarP(&src, "source", "s", []string{}, "Source directories to watch (required)")
	// cmdWatch.MarkFlagRequired("source")
}

// newWatcher encapsulates the fsnotify *NewWatcher creation and provides all data
// needed for processing the events triggered by the new
func (w *watchConfigOps) createWatcher(e event.FsEventOps, g *watchConfig) error {
	id := g.Defaults.Awsprofile
	reg := g.Defaults.Awsregion
	c := w.newS3Conn(&id, &reg)

	srcD := &g.Defaults.Userpath
	trgB := &g.Defaults.S3Target
	arrU := &g.Watch.Users

	CheckedSrcDirs := make([]string, 0) // : value
	for _, u := range *arrU {
		targetD := *srcD + u.Name // <defaults.userpath> + <watch.source.name>
		for _, srcP := range u.Sources {
			tDir := targetD + srcP
			d, err := w.checkDir(tDir)
			if err != nil || !d {
				return errors.Wrapf(err, "e.NewWatcher: targetDir %s does not exist.", tDir)
			}
			// log.Printf("DEBUG[*] createWatcher,checkDir: %s", tDir)
			CheckedSrcDirs = append(CheckedSrcDirs, tDir)
		}
	}

	epi := &event.EventPushInfo{
		Session:   c,
		Userpath:  srcD,
		Watchdirs: CheckedSrcDirs,
		Bucket:    trgB,
		Key:       "",
		Results:   make(chan *event.ResultInfo), // Consumer Stage-4
	}
	e.NewWatcher(epi)
	return nil
}

// unmarshalWatchFlag will store the flag input into the global config instance and
// thereby overwriting the data received from the config file
func (w *watchConfigOps) unmarshalWatchFlag(flagIn []string, g *watchConfig) error {
	g.Watch = struct {
		Users []struct {
			Name    string   `yaml:"name"`
			Sources []string `yaml:"sources"`
		} "yaml:\"users\""
	}{} // reset values set by config file

	type results struct {
		name  string
		paths []string
	}

	for _, entries := range flagIn {
		r := results{}
		entries := strings.Split(entries, ",")
		// verify entry format
		if len(entries) != 2 {
			return errors.New("Ensure name, and paths are set. Run 'sftppush help'.")
		}
		for _, p := range entries {
			tokens := strings.Split(p, "=")
			k := strings.TrimSpace(tokens[0])
			v := strings.TrimSpace(tokens[1])
			switch k {
			case "name":
				r.name = v
			case "paths":
				r.paths = strings.Fields(v)
			default:
				return errors.Errorf("Unknown entry: %s", p)
			}
		}

		g.Watch.Users = append(g.Watch.Users, struct {
			Name    string   `yaml:"name"`
			Sources []string `yaml:"sources"`
		}{
			Name:    r.name,
			Sources: r.paths,
		})

	}
	return nil
}

// newS3Conn creates a new AWS Api session
func (w *watchConfigOps) newS3Conn(p *string, r *string) *s3.S3 {
	// TODO Use EC2 Instance Role

	// ####
	// # Secure Credentials
	// ####

	// Initial credentials loaded from SDK's default credential chain. Such as
	// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
	// Role. These credentials will be used to to make the STS Assume Role API.
	// sess := session.Must(session.NewSession())

	// Create the credentials from AssumeRoleProvider to assume the role
	// referenced by the "myRoleARN" ARN.
	//creds := stscreds.NewCredentials(sess, "myRoleArn")

	// Create service client value configured for credentials
	// from assumed role.
	//svc := s3.New(sess, &aws.Config{Credentials: creds})/

	profile := *p
	region := *r
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", profile),
	})
	if err != nil {
		log.Fatalf("FATAL[-] cmdWatch, NewSession: %s\n", err)
	}
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Printf("WARNING[-] cmdWatch, Credentials: %s\n", err)
	}

	svcS3 := s3.New(sess)
	log.Printf("INFO[+] NewSess: %s\n", svcS3.ClientInfo.Endpoint)
	return svcS3
}

// checkDir ensures that the source watch directories exist
func (w *watchConfigOps) checkDir(p string) (bool, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return false, err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
	case mode.IsRegular():
		return false, nil
	}
	return true, nil
}
