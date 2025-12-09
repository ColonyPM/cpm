
# Colony Package Manager (cpm)

Colony Package Manager (cpm) is a package manager designed for ColonyOS. It allows you to search, install, manage, and deploy packages within the ColonyOS environment.

## Features

- **Package Management**: Easily search, install, upload, list, and remove packages.
- **Deployment**: Deploy packages with executors, functions or workflows to ColonyOS with ease.
- **Template Support**: Customize packages using YAML templates.


## Repository

You can browse, search, and find uploaded packages on our [ColonyPM Repo](https://colony.xyz). 

## CLI Commands

### Package Management (`cpm pkg`)

- **Initialize a package**  
  `cpm pkg init [DIR]`

- **Search for a package**  
  `cpm pkg search <pkg>`

- **Install a package**  
  `cpm pkg install <pkg>`

- **Upload a package**  
  `cpm pkg upload <token>`

- **Remove a package** (alias: `rm`)  
  `cpm pkg remove <pkg>`

- **List installed packages** (alias: `ls`)  
  `cpm pkg list`

- **Get package details**  
  `cpm pkg get <pkg>`

### Deployment Commands (`cpm deploy`)

- **Deploy a package**  
  `cpm deploy <pkg>`

- **List deployed packages** (alias: `ls`)  
  `cpm deploy list`

- **Remove a deployed package** (alias: `rm`)  
  `cpm deploy remove <pkg>`

- **Get details of a deployed package**  
  `cpm deploy get <pkg>`

- **Deploy a function from a package** (alias: `func`)  
  `cpm deploy function <pkg> <fn-spec>`

- **Deploy a workflow from a package** (alias: `wf`)  
  `cpm deploy workflow <pkg@version> <workflow>`

## Package File Structure

Here's an example of a package's file structure:
```json
my-package/
└─── templates/
│   │   executor.json
│   │   function.json
│   │   workflow.json
│   │   ...
│  package.yaml
│  readme.md
│  values.yaml 

```

## Package Manifest

Here's an example of the `package.yaml` manifest for a package:

```yaml
name: "my-package"
version: "0.1.0"
description: "This is an example package"
author: "John Snow"
deprecated: false
deploy:
  setup: 
	  - "colonypm/a-setup-img"
  funcSpecs:
  workflows:
  executors:
	  - name: "an-executor"
	    img: "colonypm/an-executor"
