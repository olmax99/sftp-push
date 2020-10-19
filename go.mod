module github.com/olmax99/sftppush

require (
	github.com/aws/aws-sdk-go v1.34.33
	github.com/fsnotify/fsnotify v1.4.7
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.1.2
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	golang.org/x/tools v0.0.0-20201010145503-6e5c6d77ddcc // indirect
	golang.org/x/tools/gopls v0.5.1 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
)

go 1.15

replace github.com/fsnotify/fsnotify v1.4.7 => github.com/olmax99/fsnotify v1.5.0
