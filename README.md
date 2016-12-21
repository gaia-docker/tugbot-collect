# tugbot-collect
[![CircleCI](https://circleci.com/gh/gaia-docker/tugbot-collect.svg?style=shield)](https://circleci.com/gh/gaia-docker/tugbot-collect)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaia-docker/tugbot-collect)](https://goreportcard.com/report/github.com/gaia-docker/tugbot-collect)
[![Coverage Status](https://coveralls.io/repos/github/gaia-docker/tugbot-collect/badge.svg?branch=master)](https://coveralls.io/github/gaia-docker/tugbot-collect?branch=master)
[![Docker badge](https://img.shields.io/docker/pulls/gaiadocker/tugbot-collect.svg)](https://hub.docker.com/r/gaiadocker/tugbot-collect/)
[![Docker Image Layers](https://imagelayers.io/badge/gaiadocker/tugbot-collect:latest.svg)](https://imagelayers.io/?images=gaiadocker/tugbot-collect:latest 'Get your own badge on imagelayers.io')

collects test results from test containers, digest and send to result services or to disk.

## Usage
Run `docker run -it -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect tugbot-collect -h`

To get the usage:
```
NAME:
   tugbot-collect - Collects result from test containers (use TC_LOG_LEVEL env var to change the default which is debug

USAGE:
   tugbot-collect [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --publishTarGzTo URL, -g URL               send http POST to URL with tar.gz payload contains all of the extracted results. To disable - set this flag to 'null' (default: "http://result-service:8080/results")
   --publishTestsTo URL, -c URL               send http POST to URL in json format for any junit test extracted from junit XMLs within the results dir. To disable - set this flag to 'null' (default: "http://result-service-es:8081/results")
   --outputDir DIR_LOCATION, -o DIR_LOCATION  write results to DIR_LOCATION, if you want not to output results - set this flag with the directory '/dev/null' (default: "/tmp/tugbot-collect")
   --resultsDirLabel KEY, -r KEY              tugbot-collect will use this docker label KEY and fetch the label value from the test container to resolve the results dir location. If the label could not be found on the test container, the default results dir location: '/var/tests/results' will be in use (default: "tugbot-results-dir")
   --matchLabel KEY, -m KEY                   tugbot-collect will collect results from test containers matching this label KEY (default: "tugbot-test")
   --scanOnStartup, -e                        scan for existed containers on startup and extract their results (default is false)
   --dockerRM, -d                             remove the container after extracting results (default is false)
   --skipEvents, -s                           do not register to docker 'die' events (default is false - by default we register to events and collect results for any stopped or killed test container)
   --help, -h                                 show help
   --version, -v                              print the version
```

# Additional notes
- If you want to use tubgot-collect default settings, this should be enough:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect`
- If you want to write the results to a disk on the host, you should add volume mapping, for example:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/tugbot-collect:/tmp/tugbot-collect gaiadocker/tugbot-collect`
- To change log level (which is debug by default), use this for example: 
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock -e TS_LOG_LEVEL=warn gaiadocker/tugbot-collect`
- To pass flag to tugbot-collect, use this for example:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect tugbot-collect -e -d`

