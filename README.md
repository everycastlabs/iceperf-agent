# ICEPerf
![ICEPerf logo](assets/ICEPerf_fulllogo_nobuffer.png)

ICEPerf is an open source project that helps you test and compare the performance of TURN networks. See more info and the latest results on [ICEPerf.com](https://iceperf.com).


# ICEPerf CLI APP
Run ICE servers performance tests and report the results.

## Installation
You can download the pre-built binary, the docker image or download the code and build it locally.

### Binary
Download the binary to run it natively - You can find these in the [github releases](https://github.com/nimbleape/iceperf-agent/releases).

### Docker

Run the docker image as a container.

`docker run -d nimbleape/iceperf-agent --timer --api-key=your-api-key`

The default CLI flags passed into the binary is `-h` outputting the help information.

If you want to pass in a config file you would need to pass in the config file as a file mount and the cli flags; like so:

`docker run -v $PWD/config.yaml:/config.yaml -d nimbleape/iceperf-agent --config config.yaml`

### Build locally
To install the local project, clone the repo and run the following command from the root folder:

```zsh
go install ./cmd/iceperf
```

## Running the app
To run the app from the terminal, do:

```zsh
iceperf --config path/to/config.yaml
```

```zsh
iceperf --api-key="foo"
```

### Commands
None yet.

### Flags
- `--config` or `-c` to specify the path for the config `.yaml` file (required)
- `-h` or `--help` for the help menu
- `-v` or `--version` for the app version
- `--api-uri` or `-a` to specify the API URI
- `--api-key` or `-k` to specify the API Key
- `--timer` or `-t` to enable Timer Mode (default: false)

### Config file
A `.yaml` file to provide ICE server providers credentials and other settings. Examlpes of two config files can be found in the repo. Rename `config-api.yaml.exmaple` and `config.yaml.example` to remove the `.example` extension.

`config-api.yaml` is a minimal config when talking to the ICEPerf api.
`config.yaml` is a full example if not talking to the ICEPerf api.
