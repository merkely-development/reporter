---
title: "kosli get snapshot"
---

# kosli get snapshot

## Synopsis

Get a specific environment snapshot.
Specify SNAPPISH by:
	environmentName~<N>  N'th behind the latest snapshot
	environmentName#<N>  snapshot number N
	environmentName      the latest snapshot

```shell
kosli get snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshot  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# get the latest snapshot of an environment:
kosli get snapshot yourEnvironmentName
	--api-token yourAPIToken \
	--org yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli get snapshot yourEnvironmentName~1
	--api-token yourAPIToken \
	--org yourOrgName 

# get the snapshot number 23 of an environment:
kosli get snapshot yourEnvironmentName#23
	--api-token yourAPIToken \
	--org yourOrgName 
```
