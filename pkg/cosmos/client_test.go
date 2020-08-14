package cosmos

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcos/client-go/dcos"
	"github.com/dcos/dcos-cli/pkg/config"
	"github.com/dcos/dcos-cli/pkg/httpclient"
	"github.com/dcos/dcos-cli/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_client_PackageList(t *testing.T) {
	content, err := ioutil.ReadFile("testdata/vnd.dcos.package.list-response.json")
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := "application/vnd.dcos.package.list-response+json;charset=utf-8;version=v1"
		if !(r.URL.String() == "/package/list" &&
			r.Method == http.MethodPost &&
			r.Header.Get("Accept") == contentType+",application/vnd.dcos.package.error+json;charset=utf-8;version=v1") {
			t.Error("Not expected", r)
		}
		w.Header().Set("Content-Type", contentType)
		_, err := w.Write(content)
		require.NoError(t, err)

	}))
	defer ts.Close()
	ctx := newContext(ts)

	client, err := NewClient(ctx, httpclient.New(ts.URL))
	require.NoError(t, err)

	packages, err := client.PackageList()
	require.NoError(t, err)

	jenkinsCmd := &dcos.CosmosPackageCommand{Name: "jenkins"}
	mysqlCmd := dcos.CosmosPackageCommand{Name: "mysql"}
	releaseVersion := int64(0)

	expected := []Package{
		{
			Apps:        []string{"/jenkins-test"},
			Command:     jenkinsCmd,
			Description: "Jenkins is an award-winning, cross-platform, continuous integration and continuous delivery application…",
			Framework:   true,
			Licenses: []dcos.CosmosPackageLicense{
				{
					Name: "Apache License Version 2.0",
					Url:  "https://github.com/mesosphere/dcos-jenkins-service/blob/master/LICENSE",
				},
				{
					Name: "Apache License Version 2.0",
					Url:  "https://github.com/jenkinsci/mesos-plugin/blob/master/LICENSE",
				},
				{
					Name: "MIT License",
					Url:  "https://github.com/jenkinsci/jenkins/blob/master/LICENSE.txt",
				},
			},
			Maintainer:         "support@mesosphere.io",
			Name:               "jenkins",
			PackagingVersion:   "3.0",
			PostInstallNotes:   "Jenkins has been installed.",
			PostUninstallNotes: "Jenkins has been uninstalled…",
			PreInstallNotes:    "WARNING: If you didn't provide a value for `storage.host-volume`…",
			ReleaseVersion:     &releaseVersion,
			Scm:                "https://github.com/mesosphere/dcos-jenkins-service.git",
			Selected:           true,
			Tags:               []string{"continuous-integration", "ci", "jenkins"},
			Version:            "3.6.1-2.190.1",
			Website:            "https://jenkins.io",
		},
		{
			Apps:        []string{"/jenkins/jenkins0"},
			Command:     jenkinsCmd,
			Description: "Jenkins is an award-winning, cross-platform, continuous integration and continuous delivery application…",
			Framework:   true,
			Licenses: []dcos.CosmosPackageLicense{
				{
					Name: "Apache License Version 2.0",
					Url:  "https://github.com/mesosphere/dcos-jenkins-service/blob/master/LICENSE",
				},
				{
					Name: "Apache License Version 2.0",
					Url:  "https://github.com/jenkinsci/mesos-plugin/blob/master/LICENSE",
				},
				{
					Name: "MIT License",
					Url:  "https://github.com/jenkinsci/jenkins/blob/master/LICENSE.txt",
				},
			},
			Maintainer:         "support@mesosphere.io",
			Name:               "jenkins",
			PackagingVersion:   "3.0",
			PostInstallNotes:   "Jenkins has been installed.",
			PostUninstallNotes: "Jenkins has been uninstalled…",
			PreInstallNotes:    "WARNING: If you didn't provide a value for `storage.host-volume`…",
			ReleaseVersion:     &releaseVersion,
			Scm:                "https://github.com/mesosphere/dcos-jenkins-service.git",
			Selected:           true,
			Tags:               []string{"continuous-integration", "ci", "jenkins"},
			Version:            "4.0.0-2.204.6-beta9",
			Website:            "https://jenkins.io",
		},
		{
			Apps:        []string{"/data-services/kafka", "/data-services/kafka-kerberos"},
			Description: "Apache Kafka is used for building real-time data pipelines and streaming apps…",
			Framework:   true,
			Licenses: []dcos.CosmosPackageLicense{
				{
					Name: "Apache License Version 2.0",
					Url:  "https://raw.githubusercontent.com/apache/kafka/trunk/LICENSE",
				},
			},
			Maintainer:         "support@mesosphere.io",
			Name:               "kafka",
			PackagingVersion:   "4.0",
			PostInstallNotes:   "The DC/OS Apache Kafka service is being installed…",
			PostUninstallNotes: "The DC/OS Apache Kafka service is being uninstalled…",
			PreInstallNotes:    "Default configuration requires 3 agent nodes…",
			Scm:                "https://github.com/mesosphere/dcos-kafka-service",
			Selected:           true,
			Tags:               []string{"message", "broker", "pubsub", "kafka"},
			Version:            "2.10.0-2.4.0",
			Website:            "https://docs.mesosphere.com/services/kafka/2.10.0-2.4.0",
		},
		{
			Apps: []string{
				"/data-services/kafka-zookeeper",
			},
			Description: "Apache ZooKeeper is a centralized service for maintaining configuration information…",
			Framework:   true,
			Licenses: []dcos.CosmosPackageLicense{
				{
					Name: "Apache License Version 2.0",
					Url:  "https://raw.githubusercontent.com/apache/zookeeper/master/LICENSE.txt",
				},
			},
			Maintainer:         "support@mesosphere.com",
			Name:               "kafka-zookeeper",
			PackagingVersion:   "4.0",
			PostInstallNotes:   "The DC/OS Apache ZooKeeper service is being installed…",
			PostUninstallNotes: "The DC/OS Apache ZooKeeper service is being uninstalled…",
			PreInstallNotes:    "Default configuration requires 3 agent nodes…",
			Scm:                "https://github.com/mesosphere/dcos-zookeeper",
			Selected:           true,
			Tags:               []string{"zookeeper", "kafka"},
			Version:            "2.7.0-3.4.14",
			Website:            "https://docs.mesosphere.com/services/kafka-zookeeper/2.7.0-3.4.14",
		},
		{
			Apps: []string{
				"/data-services/mysql",
			},
			Command:     &mysqlCmd,
			Description: "MySQL is the world's most popular open source database…",
			Framework:   false,
			Licenses: []dcos.CosmosPackageLicense{
				{
					Name: "GNU GENERAL PUBLIC LICENSE",
					Url:  "https://github.com/mysql/mysql-server/blob/5.7/COPYING",
				},
			},
			Maintainer:         "https://dcos.io/community/",
			Name:               "mysql",
			PackagingVersion:   "3.0",
			PostInstallNotes:   "Service installed…",
			PostUninstallNotes: "Service uninstalled…",
			PreInstallNotes:    "This DC/OS Service is currently in preview…",
			ReleaseVersion:     &releaseVersion,
			Scm:                "https://github.com/mysql/mysql-server.git",
			Selected:           false,
			Tags:               []string{"database", "mysql", "sql"},
			Version:            "5.7.12-0.3",
			Website:            "https://mysql-ci.org",
		},
	}

	assert.Equal(t, expected, packages, packages)
}

func newContext(ts *httptest.Server) *mock.Context {
	env := mock.NewEnvironment()
	ctx := mock.NewContext(env)
	cluster := config.NewCluster(nil)
	cluster.SetURL(ts.URL)
	cluster.Config().SetPath("testDir/")
	ctx.SetCluster(cluster)
	return ctx
}
