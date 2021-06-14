# platform-pipeline

[![CircleCI](https://circleci.com/gh/singnet/platform-pipeline.svg?style=svg)](https://circleci.com/gh/singnet/platform-pipeline)

SingularityNET Platform CI/CD Pipeline Repository


## Run integration tests with CircleCI

* Install [Docker CE](https://docs.docker.com/install/linux/docker-ce/ubuntu)
* Add $USER to the docker group [Post-installation steps](https://docs.docker.com/install/linux/linux-postinstall)
* Install [CircleCI command-line tool](https://circleci.com/docs/2.0/local-cli)
* checkout platform-pipeline build
  * > git checkout https://github.com/singnet/platform-pipeline.git
* Go to the platform-pipeline directory
* Run CircleCI
  * > circleci build

## Add a new test scenario

Integration tests use [GoDog](https://github.com/cucumber/godog) for test writing which is based on
cucumber gherkin3 parser.
 export GO111MODULE=on, running go mod init, and then installing a specific version of godog with a command like go get github.com/cucumber/godog/cmd/godog@v0.8.1
For example, test organization registering:
* Add a new scenario to the [features/publish_service.feature](features/publish_service.feature) file:
```gherkin
	Scenario: Register an organization
		When  organization is registered
			| organization name |
			| TestOrganization  |
		Then  organization is included into the list
			| organization name |
			| TestOrganization  |
```
* Go to the platform-pipeline directory
* Run godog
  * > godog

It will print stub methods which needs to be implemented:
```go
//You can implement step definitions for undefined steps with these snippets:

func organizationIsRegistered(arg1 *godog.Table) error {
	return godog.ErrPending
}

func organizationIsIncludedIntoTheList(arg1 *gherkin.DataTable) error {
	return godog.ErrPending
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^organization is registered$`, organizationIsRegistered)
	s.Step(`^organization is included into the list$`, organizationIsIncludedIntoTheList)
}
```

* Put the FeatureContext steps into the [features_steps_test.go](features_steps_test.go) file:
```go
func FeatureContext(s *godog.Suite) {

	// background
	// ...

	// example-services sample
	// ...

	// Check organization
	s.Step(`^organization is registered$`, organizationIsRegistered)
	s.Step(`^organization is included into the list$`, organizationIsIncludedIntoTheList)
}
```

* Create a test file **register_organization_test.go** and add stub functions into it:
```go
package main

import (
	"github.com/cucumber/godog"
)

func organizationIsRegistered(table  *godog.Table) error {
	return godog.ErrPending
}

func organizationIsIncludedIntoTheList(table *godog.Table) error {
	return godog.ErrPending
}
```

* Use Command DSL to run commands from the shell and check that the output
contains necessary words:

```go
func organizationIsRegistered(table *gherkin.DataTable) error {

	organizationName := getTableValue(table, "organization name")

	return NewCommand().
		Run("snet organization create %s --registry-at %s -y", organizationName, registryAddress).
		Err()
}

func organizationIsIncludedIntoTheList(table *gherkin.DataTable) error {

	organizationName := getTableValue(table, "organization name")
	output := "organizations.out"

	return NewCommand().
		Output(output).
		Run("snet organization list").
		CheckOutput(organizationName).
		Err()
}
```

`Note`: Background in publish_service.feature file is executed before each scenario. It means that in case of
more than one scenario the background tests should be updated to be executed only once.


## Troubleshoot failed platform-pipeline project

The following steps can help to reproduce and find root cause of the failed platform pipeline build.

`Note`: platform-pipeline always builds all singnet projects from master branch.

* Find the step where platform-pipeline failed in [CircleCI Web UI](https://circleci.com/gh/singnet/platform-pipeline)
* Open the .circleci/config.yaml file in the platform-pipeline project
* Insert *bash* command before the failed step.

For example, the *bash* command was inserted before running GoDog integration tests.

```yaml
    - checkout
    - run:
        name: Run integration tests
        command: |
          export PATH=$PATH:$GOPATH/bin
          mkdir $GOPATH/log
          go get github.com/DATA-DOG/godog/cmd/godog
          bash # <-- inserted command
          godog
```

* Run CircleCI in the platform-pipeline project
  * > cd platform-pipeline
  * > circleci build
* Wait until CircleCI build reaches the *bash* point
* Copy the docker image
  * List docker running containers
    * > docker ps
  * Select the container id for image "ubuntu:latest" from the output
  * Copy the docker image
    * > docker commit container-id circleci-platform-pipeline
* Run the copied CircleCI image by docker
  * > docker run -it circleci-platform-pipeline bash
  * > cd /root/singnet/src/github.com/singnet

* Connect to the running docker image
  * List docker running containers
    * > docker ps
  * Select the container id for required docker
  * Run
    * > docker exec -it container-id bash

`Note`: necessary test environment variables are not set in this *bash* session and should be set manually.
```
export GOPATH=/root/singnet
export SINGNET_REPOS=/root/singnet/src/github.com/singnet
export PATH=$PATH:$GOPATH/bin
export IPFS_PATH=$GOPATH/ipfs
```
