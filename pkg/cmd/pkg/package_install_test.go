package pkg

import (
	"testing"

	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestPkgInstallMutuallyExclusiveOptionsShouldFail(t *testing.T) {
	err := pkgInstall(nil, "helloworld", pkgInstallOptions{cliOnly: true, appOnly: true})
	assert.EqualError(t, err, "--app and --cli are mutually exclusive")
}

func TestPkgInstallNotExistingOptionsPathShouldFail(t *testing.T) {
	err := pkgInstall(nil, "helloworld", pkgInstallOptions{optionsPath: "not existing path"})
	assert.Error(t, err)
}

func TestPkgInstallEmptyPackageNameShouldFail(t *testing.T) {
	err := pkgInstall(nil, "", pkgInstallOptions{})
	assert.EqualError(t, err, "package name must not be empty")
}

func TestPkgInstallShouldFailOnCosmosError(t *testing.T) {
	ctx := mock.NewContext(nil)
	err := pkgInstall(ctx, "helloworld", pkgInstallOptions{})
	assert.EqualError(t, err, `Post /package/describe: unsupported protocol scheme ""`)
}
