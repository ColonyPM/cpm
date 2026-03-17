# Setup & Example
This document is a guide to setup the ColonyPM system and aids in performing a use case, from initializing a package to deploying it.

## Setup
### Prerequisites
- A running [colony server](https://github.com/colonyos/colonies) with minio
- Docker and Docker Compose
- A configured [GitHub OAuth app](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app)

### Repository
See the repository [README](https://github.com/ColonyPM)

### cpm CLI
See the cpm [README](https://github.com/ColonyPM/cpm)
To see all available commands and their use cases visit our [CLI docs](docs.colonypm.xyz) or build it yourself, alternatively read the source markdown files at the [docs repository](https://github.com/ColonyPM/docs).


### A cpm-integrated  colony node (host or separate)
See the cpm-executor [README](https://github.com/ColonyPM/cpm-executor)

## Example
### Create & upload a package
```
cpm pkg init pingpong
```
modify ``pingpong/package.yaml`` to contain:
```yaml
name: pingpong
version: 1.5.0
description: A demo of CPM
author: "Your Name"
deprecated: false
deploy:
    functionSpecs: []
    workflows: []
    executors:
        - name: ping-executor
          img: "confusedswede/ping-executor"
        - name: pong-executor
          img: "confusedswede/pong-executor"
    setup: []
    teardown: []

```
create ``ping.json``in ``pingponpg/templates/`` with the following contents:
```json
{
    "conditions": {
        "executortype": "ping-executor",
	"colonyname": "dev"
    },
    "funcname": "ping"
}

```
Visit your local instance of the repository, login, navigate to your profile and generate an upload token. Run:
```bash
cpm pkg upload pingpong --token <your token>
```
### Download & Deploy Pingpong
Download with:
```bash
cpm pkg download pingpong@latest
```
To deploy:
```bash
cpm pkg download pingpong@latest
```
Done! If you wish to see the executors in action run:
```bash
docker logs -f <container id>
```
where ``container id`` is either the ping or pong executor's container id.
