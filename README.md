# ddnsb0t

![ddnsb0t](https://github.com/jonpulsifer/ddnsb0t/workflows/ddnsb0t/badge.svg)
![function](https://github.com/jonpulsifer/ddnsb0t/workflows/function/badge.svg)
![docker](https://github.com/jonpulsifer/ddnsb0t/workflows/docker/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jonpulsifer/ddnsb0t)](https://goreportcard.com/report/github.com/jonpulsifer/ddnsb0t)

**ddnsb0t** is a program that uses [CloudEvents](https://cloudevents.io/) to communicate to a [Google Cloud Function](https://cloud.google.com/functions) and update my DNS entries using [Google Cloud DNS](https://cloud.google.com/dns).

```raw
ddnsb0t -  A bot that fires a CloudEvent down range to a cloud function to update my DNS records in Google Cloud.

Usage: ddnsb0t <command>

Flags:

  --domain    the default domain (default: <none>)
  --endpoint  the remote URL for the cloud function (default: <none>)
  --external  use the network's external IP address (default: false)
  --hostname  the hostname to update (default: <none>)
  --interval  how long between each update (eg. 30s, 5m, 1h) (default: 5m0s)
  --once      run the thing once (default: false)
  --token     an api token for the cloud function to prevent abuse (default: <none>)
  --verbose   set the log level to debug (default: false)

Commands:

  version  Show the version information.
```

## Installation

1. Install ddnsb0t using `go get https://github.com/jonpulsifer/ddnsb0t`
2. Run ddnsb0t `$GOPATH/bin/ddnsb0t -endpoint=https://fn.example.com/ddns -token=sometoken -once`
3. Optionally configure a cron job (below)

```sh
#!/bin/sh
# */5 * * * * /path/to/this/script.sh
export DDNS_ENDPOINT="https://your.url.example.com/ddns"
export DDNS_DOMAIN="example.com"
export DDNS_API_TOKEN="sometoken"
ddnsb0t --once "$@"
```

Running `ddnsb0t` should produce the response from the cloud function

```sh
INFO[0000] dns update requested   fqdn=somename.example.com. ip=10.13.37.1 status=pending
```

## Development

1. Run a local cloud events receiver by running `go run ./cmd/main.go` from the `function` directory. This will start an HTTP receiver at `http://localhost:8080`
2. Build ddnsb0t: `go build -o ddnsb0t`
3. Use ddnsb0t: `./ddnsb0t -endpoint=http://localhost:8080 -token=sometoken -once`

### Cloud Events

#### Request

```raw
Validation: valid
Context Attributes,
  specversion: 1.0
  type: dev.pulsifer.ddns.request
  source: https://github.com/jonpulsifer/ddnsb0t
  id: 521ccda3-b297-43a0-887c-a76b25557806
  dataschema: https://github.com/jonpulsifer/ddnsb0t/pkg/ddns/ddns.go
  datacontenttype: application/json
Data,
  {
    "ip": "10.13.37.1",
    "fqdn": "somename.example.com.",
    "token": "sometoken"
  }
```

#### Response

```raw
Validation: valid
Context Attributes,
  specversion: 1.0
  type: dev.pulsifer.ddns.response
  source: https://github.com/jonpulsifer/ddnsb0t
  id: 5c50e41c-2074-446e-8d4b-0c038e04eb07
  dataschema: https://github.com/jonpulsifer/ddnsb0t/pkg/ddns/ddns.go
  datacontenttype: application/json
Data,
  {
    "fqdn": "somename.example.com.",
    "status": "pending",
    "additions": 1,
    "deletions": 1
  }
```
