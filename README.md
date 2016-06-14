# tugbot-collect
collects test results from test containers and save the results to disk

## Docker run command

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect
```

## Usage

```
./tugbot-collect --help
NAME:
   tugbot-collect - Collects result from test containers (use TC_LOG_LEVEL env var to change the default which is debug

USAGE:
   tugbot-collect [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
GLOBAL OPTIONS:
   --outputdir DIR_LOCATION, -o DIR_LOCATION  write results to DIR_LOCATION, if you want not to output results - set this flag with the directory '/dev/null' (default: "/tmp/tugbot-collect")
   --resultsdirlabel KEY, -r KEY              tugbot-collect will use this label KEY to fetch the label value, to find out the results dir of the test container (default: "tugbot.results.dir")
   --matchlabel KEY, -m KEY                   tugbot-collect will collect results from test containers matching this label KEY (default: "tugbot.created.from")
   --scanonstartup, -e                        scan for existed containers on startup and extract their results (default is false)
   --dockerrm, -d                             remove the container after extracting results (default is false)
   --skipevents, -s                           do not register to docker 'die' events (default is false - hence by default we do register to events)
   --help, -h                                 show help
   --version, -v                              print the version
```