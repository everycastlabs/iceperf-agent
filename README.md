# ICEPerf
![ICEPerf logo](assets/ICEPerf_fulllogo_nobuffer.png)

ICEPerf is an open source project that helps you test and compare the performance of TURN networks. See more info and the latest results on [ICEPerf.com](https://iceperf.com).


# ICEPerf CLI APP
Run ICE servers performance tests and report the results.

## Installation
You can download the pre-built binary, or download the code and build it locally.

### Binary
Download the binary to run it natively (currently only on Linux x86).

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

### Commands
None yet.

### Flags
- `--config` or `-c` to specify the path for the config `.yaml` file (required)
- `-h` or `--help` for the help menu
- `-v` or `--version` for the app version

### Config file
A `.yaml` file to provide ICE server providers credentials and other settings.

Example:

```yaml
logging:
  level: debug
   loki:
    enabled: true
    use_headers_auth: false
    use_basic_auth: true
    username: username
    password: password
    url: url
ice_servers:
  metered:
    enabled: false
    username: aaaaa
    password: bbbbb
    api_key: ccccc
    request_url: https://accountname.metered.live/api/v1/turn/credentials
  cloudflare:
    enabled: true
    username: aaaaa
    password: bbbbb
    stun_host: stun.cloudflare.com
    turn_host: turn.cloudflare.com
    turn_ports:
      udp:
        - 3478
        - 53
      tcp:
        - 3478
        - 80
      tls:
        - 5349
        - 443
   elixir:
    enabled: true
    http_username: username
    request_url: url
    stun_enabled: false
    turn_enabled: true
  twilio:
    enabled: false
    http_username: aaaaa
    http_password: bbbbb
    request_url: https://api.twilio.com/2010-04-01/Accounts/$TWILIO_ACCOUNT_SID/Tokens.json
    request_method: POST
  xirsys:
    enabled: false
    http_username: aaaaa
    http_password: bbbbb
    request_url: https://global.xirsys.net/_turn/IcePerf
  expressturn:
    enabled: false
    username: aaaaa
    password: bbbbb
    stun_host: relay1.expressturn.com
    turn_host: relay1.expressturn.com
    turn_ports:
      udp:
        - 3478
        - 80
      tcp:
        - 3478
        - 443
      tls:
        - 5349
        - 443
```
