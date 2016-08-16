#!/bin/bash

set -x

# running the test container - the container will exit immediately 
# we map a volume that contains junit xml example file to be used as the results dir
docker run -v "$PWD":/tmp --label tugbot.test=true --label tugbot.results.dir=/tmp alpine /bin/sh

# running mock http server to simulate tar.gz result service api and json test result service api
docker run -d -p 8085:8085 --name mock-result-service msoap/shell2http -port=8085 -cgi /tar-results 'cat /dev/stdin > result.tar' /get-tar-results 'cat result.tar > /dev/stdout' /json-results 'cat /dev/stdin > result.json' /get-json-results 'cat result.json > /dev/stdout'

# running tugbot-collect to report the test results from the test container into the mock result service
docker run -d --name tugbot-collect --link mock-result-service:mock-result-service -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect tugbot-collect --scanOnStartup -skipEvents --publishTestsTo http://mock-result-service:8085/json-results --publishTarGzTo http://mock-result-service:8085/tar-results

sleep 1

# check that the mock result service recieved the data
curl http://localhost:8085/get-tar-results | tar tvz | grep results.tar
curl http://localhost:8085/get-json-results | grep TugbotData | grep NumericStatus
