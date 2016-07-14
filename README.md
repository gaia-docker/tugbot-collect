# tugbot-collect
[![CircleCI](https://circleci.com/gh/gaia-docker/tugbot-collect.svg?style=shield)](https://circleci.com/gh/gaia-docker/tugbot-collect)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaia-docker/tugbot-collect)](https://goreportcard.com/report/github.com/gaia-docker/tugbot-collect)
[![Docker badge](https://img.shields.io/docker/pulls/gaiadocker/tugbot-collect.svg)](https://hub.docker.com/r/gaiadocker/tugbot-collect/)
[![Docker Image Layers](https://imagelayers.io/badge/gaiadocker/tugbot-collect:latest.svg)](https://imagelayers.io/?images=gaiadocker/tugbot-collect:latest 'Get your own badge on imagelayers.io')

collects test results from test containers and save the results to disk

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
GLOBAL OPTIONS:
   --resultserviceurl URL, -u URL             write results to URL, if you want to post results to the results service - set this flag with 'null' (default: "http://localhost:8080/results")
   --outputdir DIR_LOCATION, -o DIR_LOCATION  write results to DIR_LOCATION, if you want not to output results - set this flag with the directory '/dev/null' (default: "/tmp/tugbot-collect")
   --resultsdirlabel KEY, -r KEY              tugbot-collect will use this label KEY to fetch the label value, to find out the results dir of the test container (default: "tugbot.results.dir")
   --matchlabel KEY, -m KEY                   tugbot-collect will collect results from test containers matching this label KEY (default: "tugbot.created.from")
   --scanonstartup, -e                        scan for existed containers on startup and extract their results (default is false)
   --dockerrm, -d                             remove the container after extracting results (default is false)
   --skipevents, -s                           do not register to docker 'die' events (default is false - hence by default we do register to events)
   --help, -h                                 show help
   --version, -v                              print the version

```

# Addtional notes
- If you want to use tubgot-collect default settings, this should be enough:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect`
- If you want to write the results to a disk on the host, you should add volume mapping, for example:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock -v /tmp/tugbot-collect:/tmp/tugbot-collect gaiadocker/tugbot-collect`
- To change log level (which is debug by default), use this for exmaple: 
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock -e TS_LOG_LEVEL=warn gaiadocker/tugbot-collect`
- To pass flag to tugbot-collect, use this for example:
`docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect tugbot-collect -e -d`

## Missions to complete
- Integrate with result service
- Extract logs and docker inspect info
- Writing tests (including code cov)
