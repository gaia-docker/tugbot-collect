#!/bin/bash

set -x

# running the test container - the container will exit immediately 
docker run --label tugbot.test=true --label tugbot.results.dir=/tmp alpine /bin/sh

# running mock http server to simulate tar.gz result service
docker run -d -p 8085:8085 --name mock-tar-gz-service msoap/shell2http -port=8085 -cgi /results 'cat /dev/stdin > file.tar' /get-results 'cat file.tar > /dev/stdout'

# running tugbot-collect to report the test results from the test container into the mock result service
docker run -d --name tugbot-collect --link mock-tar-gz-service:mock-tar-gz-service -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect tugbot-collect --scanOnStartup -skipEvents --publishTestsTo null --publishTarGzTo http://mock-tar-gz-service:8085/results

sleep 1

# check that the mock tar.gz service recieved the data
curl http://localhost:8085/get-results | tar tvz | grep results.tar
