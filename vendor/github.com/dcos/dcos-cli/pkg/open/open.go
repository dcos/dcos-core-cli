package open

import (
	"github.com/sirupsen/logrus"
)

// Opener opens a file or URL in the user's preferred application.
type Opener interface {
	Open(resource string) error
}

// OsOpener is an Opener using OS-specific commands.
// It is only relevant in desktop environments.
type OsOpener struct {
	logger *logrus.Logger
}

// NewOsOpener creates a new OsOpener,
func NewOsOpener(logger *logrus.Logger) *OsOpener {
	return &OsOpener{
		logger: logger,
	}
}
