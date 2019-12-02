module github.com/dcos/dcos-core-cli

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/antihax/optional v0.0.0-20180407024304-ca021399b1a6
	github.com/dcos/client-go v0.0.0-20190910161559-e3e16c6d1484
	github.com/dcos/dcos-cli v0.0.0-20191121101114-91bc46caf036
	github.com/docker/docker v0.7.3-0.20190611184350-29829874d173
	github.com/dustin/go-humanize v1.0.0
	github.com/gambol99/go-marathon v0.7.2-0.20191118125545-9c4da387a2a4
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mesos/mesos-go v0.0.11-0.20190717023829-56ac038085ac
	github.com/olekukonko/tablewriter v0.0.1
	github.com/pkg/errors v0.8.1 // indirect
	github.com/pquerna/ffjson v0.0.0-20181028064349-e517b90714f7 // indirect
	github.com/r3labs/sse v0.0.0-20181217150409-243b7807c4c4
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	golang.org/x/crypto v0.0.0-20190911031432-227b76d455e7
	golang.org/x/net v0.0.0-20190918130420-a8b05e9114ab // indirect
	golang.org/x/sys v0.0.0-20190919044723-0c1ff786ef13 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/gambol99/go-marathon => github.com/mesosphere/go-marathon v0.7.2-0.20191202124444-58c41efabec5

go 1.12
