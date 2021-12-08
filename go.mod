module github.com/foomo/squadron

go 1.16

require (
	github.com/1Password/connect-sdk-go v1.2.0
	github.com/miracl/conflate v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/pterm/pterm v0.12.33
	github.com/sergi/go-diff v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.2
)

replace github.com/miracl/conflate v1.2.1 => github.com/runz0rd/conflate v1.2.2-0.20210920145208-fa48576ef06d
