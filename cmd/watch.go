package cmd

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type watchConfig struct {
	Defaults struct {
		Userpath string `yaml:"userpath"`
	} `yaml:"defaults"`
	Watch struct {
		Users []struct {
			Name     string   `yaml:"name"`
			Sources  []string `yaml:"sources"`
			S3Target string   `yaml:"s3target"`
		} `yaml:"users"`
	} `yaml:"watch"`
}

var src []string // watch flag --source read as string

// versionCmd represents the version command
var cmdWatch = &cobra.Command{
	Use:   "watch",
	Short: "Start the fsnotify file system event watcher",
	Long: strings.TrimSpace(`
Use the watch command with a --source flag to indicate the 
directory, which is listened on for file events.`),
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
		log.Printf("DEBUG[+]: watch --source %#v \n", src)

		// TODO This should come from config.go (viper)
		// There should be 2 ways of setting it:
		// 1. Through config.yaml
		// 2. overwrite with SFTPPUSH_bucket
		var target string = "olmax-test-sftppush-126912"

		log.Printf("DEBUG[*] configWatch (from config): %q", &wC)
		// TODO Catch errors, implement a notification service
		// TODO Multiple targets - create a watcher for every target in target file
		e := event.FsEventOps{}
		conn := newS3Conn()

		f, err := wC.unmarshalWatchFlag(src)
		if err != nil {
			log.Fatalf("FATAL[*] decodeWatchFlag: %s", err)
		}

		log.Printf("DEBUG[*] cmdWatch (from flags): %q", f)

		// implements fsnotify.NewWatcher
		for _, s := range src {
			e.NewWatcher(s, conn, &target)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdWatch)
	cmdWatch.Flags().StringArrayVarP(&src, "source", "s", []string{}, "Source directories to watch (required)")
	// cmdWatch.MarkFlagRequired("source")
}

func (w *watchConfig) unmarshalWatchFlag(flagIn []string) (*watchConfig, error) {
	w.Watch = struct {
		Users []struct {
			Name     string   "yaml:\"name\""
			Sources  []string "yaml:\"sources\""
			S3Target string   "yaml:\"s3target\""
		} "yaml:\"users\""
	}{} // reset values set by config

	type results struct {
		name     string
		paths    []string
		s3bucket string
	}

	for _, entries := range flagIn {
		r := results{}
		entries := strings.Split(entries, ",")
		// verify entry format
		if len(entries) != 3 {
			log.Fatal("FATAL[-] ")
			return nil, errors.New("Ensure Name, paths, and s3target are set.")
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
			case "s3target":
				r.s3bucket = v
			default:
				return nil, errors.Errorf("Unknown entry: %s", p)
			}
		}

		w.Watch.Users = append(w.Watch.Users, struct {
			Name     string   "yaml:\"name\""
			Sources  []string "yaml:\"sources\""
			S3Target string   "yaml:\"s3target\""
		}{
			Name:     r.name,
			Sources:  r.paths,
			S3Target: r.s3bucket,
		})

	}
	return w, nil
}

func newS3Conn() *s3.S3 {
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

	// Only for testing
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.NewSharedCredentials("", "olmax"),
	})
	if err != nil {
		log.Fatalf("FATAL[-] cmdWatch, NewSession: %s\n", err)
	}
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		log.Printf("WARNING[-] cmdWatch, Credentials: %s\n", err)
	}

	svcs3 := s3.New(sess)
	log.Printf("INFO[+] NewSess: %s", svcs3.ClientInfo.Endpoint)
	return svcs3
}
