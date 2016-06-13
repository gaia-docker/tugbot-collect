# tugbot-collect
collects test results from test containers and send the results to results service

## Docker run command

```
docker run -d -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot-collect
```

## Description
By default, tugbot-collect will:

1. One-time scan, on startup all of the `Exited` containers that has the label `tugbot.created.from` (means tugbot runner was the one that managed them).

2. Register to docker container's `die` event (container that ends the run for any reason cause the `die` event to be published, more info about docker container events you can find [here](https://docs.docker.com/engine/reference/api/docker_remote_api/)). Look for contianers with the label `tugbot.created.from`

3. For any container discovered by the `one time scan` or as a result of the `die` event: 
  - Look for `tugbot.results.dir` label and extract the result directory as a tar file. If the label does not exist, or this operation failed for any reason - skip this step.
  
  - Extract the container info json (similar to what you get from `docker inspect`)
  
  - Try to extract the container's stdout.
  
  - Save all to the `output directory` (by default `/var/logs/tugbot-collect/`) under a unique folder (folder name will be composed of `image name`-`time`-`short container id`).
  
  - `docker rm` the container.

