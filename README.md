# tugbot-collect
collects test results from test containers and send the results to results service

## Docker run command

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect
```
