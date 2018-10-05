# platform-pipeline

[![CircleCI](https://circleci.com/gh/singnet/platform-pipeline.svg?style=svg)](https://circleci.com/gh/singnet/platform-pipeline)

SingularityNET Platform CI/CD Pipeline Repository


## Troubleshooting troubleshoot failed platform-pipeline project

The following steps can help to reproduce and find root cause of the failed platform pipeline build.

`Note`: platform-pipeline always builds all singnet projects from master branch.

* Install [Docker](https://docs.docker.com/install)
* Install [CircleCI command-line tool](https://circleci.com/docs/2.0/local-cli)
* checkout platform-pipeline build
  * > git checkout https://github.com/singnet/platform-pipeline.git
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
          # Disable TensorFlow warnings wich pollute example-service log file
          export TF_CPP_MIN_LOG_LEVEL=2
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
 * Use instructions from the [Wiki page](https://github.com/singnet/wiki/wiki/Tutorial:-Build-and-deploy-SingularityNET-locally) to reproduce integration test

`Note`: necessary test environment variables are not set in this *bash* session and should be set manually.
