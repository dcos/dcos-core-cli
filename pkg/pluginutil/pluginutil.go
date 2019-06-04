package pluginutil

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-cli/pkg/log"
	"github.com/sirupsen/logrus"
)

// HTTPClient returns an HTTP client from a plugin runtime.
func HTTPClient(baseURL string, opts ...httpclient.Option) *httpclient.Client {
	if baseURL == "" {
		baseURL, _ = os.LookupEnv("DCOS_URL")
	}
	var baseOpts []httpclient.Option

	if acsToken, _ := os.LookupEnv("DCOS_ACS_TOKEN"); acsToken != "" {
		baseOpts = append(baseOpts, httpclient.ACSToken(acsToken))
	}

	if verbosity, _ := os.LookupEnv("DCOS_VERBOSITY"); verbosity != "" {
		baseOpts = append(baseOpts, httpclient.Logger(Logger()))
	}

	tlsInsecure, _ := os.LookupEnv("DCOS_TLS_INSECURE")
	if tlsInsecure == "1" {
		baseOpts = append(baseOpts, httpclient.TLS(&tls.Config{InsecureSkipVerify: true}))
	} else {
		tlsCAPath, _ := os.LookupEnv("DCOS_TLS_CA_PATH")
		if tlsCAPath != "" {
			rootCAsPEM, err := ioutil.ReadFile(tlsCAPath)
			if err == nil {
				certPool := x509.NewCertPool()
				if certPool.AppendCertsFromPEM(rootCAsPEM) {
					baseOpts = append(baseOpts, httpclient.TLS(&tls.Config{RootCAs: certPool}))
				}
			}
		}
	}
	return httpclient.New(strings.TrimRight(baseURL, "/"), append(baseOpts, opts...)...)
}

// NewHTTPClient returns an HTTP client from a plugin runtime.
// It is created with the `client-go` package and will eventually
// replace the httpClient of the CLI.
func NewHTTPClient(baseURL string) *http.Client {
	dcosConfig := dcos.NewConfig(nil)

	if baseURL == "" {
		baseURL, _ = os.LookupEnv("DCOS_URL")
	}

	if acsToken, _ := os.LookupEnv("DCOS_ACS_TOKEN"); acsToken != "" {
		dcosConfig.SetACSToken(acsToken)
	}

	var tls dcos.TLS
	tlsInsecure, _ := os.LookupEnv("DCOS_TLS_INSECURE")
	if tlsInsecure == "1" {
		tls.Insecure = true
	} else {
		tls.RootCAsPath, _ = os.LookupEnv("DCOS_TLS_CA_PATH")
		if tls.RootCAsPath != "" {
			rootCAsPEM, err := ioutil.ReadFile(tls.RootCAsPath)
			if err == nil {
				certPool := x509.NewCertPool()
				if certPool.AppendCertsFromPEM(rootCAsPEM) {
					tls.RootCAs = certPool
				}
			}
		}
	}
	dcosConfig.SetTLS(tls)

	client := dcos.NewHTTPClient(dcosConfig)
	client.Transport.(*dcos.DefaultTransport).Logger = Logger()
	return client
}

// Logger returns a logger for a given plugin runtime.
func Logger() *logrus.Logger {
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &log.Formatter{},
		Hooks:     make(logrus.LevelHooks),
	}
	verbosity, _ := os.LookupEnv("DCOS_VERBOSITY")
	if verbosity == "1" {
		logger.SetLevel(logrus.InfoLevel)
	} else if verbosity == "2" {
		logger.SetLevel(logrus.DebugLevel)
		os.Setenv("DCOS_DEBUG", "1")
	}
	return logger
}
