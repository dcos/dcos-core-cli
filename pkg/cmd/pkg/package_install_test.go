package pkg

import (
	"testing"

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