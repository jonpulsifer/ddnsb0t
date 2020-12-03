package main

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/genuinetools/pkg/cli"
	"github.com/jonpulsifer/ddnsb0t/pkg/ddns"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

var (
	domain      string
	endpoint    string
	environment string
	external    bool
	hostname    string
	interval    time.Duration
	once        bool
	request     ddns.Request
	response    ddns.Response
	token       string
	verbose     bool
)

func main() {
	p := cli.NewProgram()
	p.Name = "ddnsb0t"
	p.Description = "A bot that fires a CloudEvent down range to a cloud function to update my DNS records in Google Cloud"

	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.BoolVar(&external, "external", false, "use the network's external IP address")
	p.FlagSet.BoolVar(&once, "once", false, "run the thing once")
	p.FlagSet.BoolVar(&verbose, "verbose", false, "set the log level to debug")
	p.FlagSet.DurationVar(&interval, "interval", 5*time.Minute, "how long between each update (eg. 30s, 5m, 1h)")
	p.FlagSet.StringVar(&domain, "domain", os.Getenv("DDNS_DOMAIN"), "the default domain")
	p.FlagSet.StringVar(&endpoint, "endpoint", os.Getenv("DDNS_ENDPOINT"), "the remote URL for the cloud function")
	p.FlagSet.StringVar(&environment, "environment", os.Getenv("DDNS_ENVIRONMENT"), "set to 'production' for JSON logging")
	p.FlagSet.StringVar(&hostname, "hostname", os.Getenv("DDNS_HOSTNAME"), "the hostname to update")
	p.FlagSet.StringVar(&token, "token", os.Getenv("DDNS_API_TOKEN"), "an api token for the cloud function to prevent abuse")

	p.Before = func(ctx context.Context) error {
		if environment == "production" {
			log.SetFormatter(&log.JSONFormatter{})
		}

		if verbose {
			log.SetLevel(log.DebugLevel)
		}

		if endpoint == "" {
			log.Fatalf("endpoint not found")
		}

		if token == "" {
			log.Warnf("API token not configured")
		}

		if hostname == "" {
			OSHostname, err := os.Hostname()
			if err != nil {
				log.Fatalf("could not get hostname from OS and hostname not set: %q", err)
			}
			log.WithFields(log.Fields{
				"hostname": hostname,
				"os":       OSHostname,
			}).Debugf("Got hostname from OS")
			hostname = OSHostname
		}

		if dns.CountLabel(hostname) < 2 && domain != "" {
			log.WithFields(log.Fields{
				"hostname": hostname,
				"domain":   domain,
			}).Debugf("Domain name too short, appending domain")
			hostname = strings.Join([]string{hostname, domain}, ".")
		}

		if !strings.HasSuffix(hostname, ".") {
			hostname = hostname + string(".")
		}

		hostname = strings.ToLower(hostname)

		if !dns.IsFqdn(hostname) || dns.CountLabel(hostname) < 2 {
			log.Fatalf("Could not determine fully qualified domain name, got: %s", hostname)
		}

		if external {
			externalIP, err := GetExternalIP()
			if err != nil {
				return err
			}
			request.IP = externalIP
		} else {
			internalIP, err := GetInternalIP()
			if err != nil {
				return err
			}
			request.IP = internalIP
		}

		if net.ParseIP(request.IP) == nil {
			log.Fatalf("could not determine IP address, got: %s", request.IP)
		}
		return nil
	}

	p.Action = func(ctx context.Context, args []string) error {

		ticker := time.NewTicker(interval)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			for sig := range c {
				ticker.Stop()
				log.Warnf("got %s, exiting", sig.String())
				os.Exit(0)
			}
		}()

		if once {
			err := update()
			if err != nil {
				return err
			}
			os.Exit(0)
		}

		for range ticker.C {
			err := update()
			if err != nil {
				return err
			}
		}
		return nil
	}
	p.Run()
}

func update() error {
	request.Token = token
	request.FQDN = hostname

	log.WithFields(log.Fields{
		"ip":   request.IP,
		"name": request.FQDN,
	}).Debugf("Built Request")

	return sendRequest(request)
}

// sendRequest fires a ddns.Request in the form of a cloud event to a receiver
func sendRequest(request ddns.Request) error {
	event := ddns.GenerateCloudEventRequest(request)
	eventBytes, err := event.MarshalJSON()
	if err != nil {
		return err
	}
	log.Debugf("raw cloudevent:\n%v", event)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(eventBytes))
	req.Header.Set("Content-Type", cloudevents.ApplicationJSON)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseEvent := cloudevents.NewEvent()
	if err := responseEvent.UnmarshalJSON(respBody); err != nil {
		return err
	}

	if responseEvent.Type() != ddns.CloudEventResponseType {
		log.Fatalf("Response was not of the expected type, expected %s, got: %s", ddns.CloudEventResponseType, responseEvent.Type())
	}

	if err := responseEvent.DataAs(&response); err != nil {
		log.Fatalf("Could not decode response: %v", err.Error())
	}
	log.WithFields(log.Fields{
		"fqdn":   response.FQDN,
		"ip":     request.IP,
		"status": response.Status,
	}).Infof("dns update requested")
	log.Debugf("raw cloudevent:\n%v", responseEvent)
	return nil

	// Expand the CloudEvent function signatures to support reply #58
	// https://github.com/GoogleCloudPlatform/functions-framework-go/issues/58
	// c, err := cloudevents.NewDefaultClient()
	// if err != nil {
	// 	log.Fatalf("failed to create client, %v", err)
	// }
	// ctx := cloudevents.ContextWithTarget(context.Background(), endpoint)

	// resp, result := c.Request(ctx, event)
	// log.WithFields(log.Fields{
	// 	"endpoint": endpoint,
	// }).Debugf("Sent request")
	// if cloudevents.IsUndelivered(result) {
	// 	log.Fatalf("failed to send, %v", result)
	// } else {
	// 	log.Infof("event delivered at %s, ack: %t ", time.Now(), cloudevents.IsACK(result))
	// 	var httpResult *cehttp.Result
	// 	if cloudevents.ResultAs(result, &httpResult) {
	// 		log.Infof("status code %d", httpResult.StatusCode)
	// 	}
	// 	if resp != nil {
	// 		log.Infof("response:\n%s\n", resp)
	// 	}
	// }
}

// GetInternalIP mocks a connection using net.Dial but does not actually make a request
func GetInternalIP() (string, error) {
	// mock a connection, this does not make a request
	conn, err := net.Dial("udp", "1.1.1.1:53")
	if err != nil {
		return "", err
	}
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}

// GetExternalIP uses Cloudflare's whoami.cloudflare TXT record cheatcode to get the external IP address
// dig -4 ch txt whoami.cloudflare @1.1.1.1
func GetExternalIP() (string, error) {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: "whoami.cloudflare.", Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}

	c := new(dns.Client)
	in, _, err := c.Exchange(m, "1.1.1.1:53")
	if err != nil {
		return "", err
	}

	if t, ok := in.Answer[0].(*dns.TXT); ok {
		return t.Txt[0], nil
	}
	return "", err
}

// cloudflareDialer returns a custom dialer that uses 1.1.1.1 instead of the default resolver
// https://github.com/golang/go/issues/19268
func cloudflareDialer(ctx context.Context, network, address string) (net.Conn, error) {
	d := net.Dialer{}
	return d.DialContext(ctx, "udp", "1.1.1.1:53")
}
