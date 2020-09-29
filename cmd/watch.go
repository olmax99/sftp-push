package cmd

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/spf13/cobra"
)

var Src string

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
		log.Printf("DEBUG[+]: watch -source %s \n", Src)

		var target string = "olmax-test-sftppush-126912"
		// TODO Catch errors, implement a notification service
		// TODO Multiple targets - create a watcher for every target in target file
		e := event.FsEventOps{}
		conn := newS3Conn()

		// implements fsnotify.NewWatcher
		e.NewWatcher(Src, conn, &target)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdWatch)
	// cmdWatch.AddCommand(cmdTarget)
	cmdWatch.Flags().StringVarP(&Src, "source", "s", "", "Source directory to watch (required)")
	cmdWatch.MarkFlagRequired("source")
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
