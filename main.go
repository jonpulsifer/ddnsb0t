package main

import (
	"context"
	"flag"
	"net"
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
	token       string
	verbose     bool
)

func main() {
	p := cli.NewProgram()
	p.Name = "ddnsb0t"
	p.Description = "A bot that fires a CloudEvent down range to a cloud function to update my DNS records in Google Cloud"
	p.Version = "0.0.2"

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
			log.Fatalf("Endpoint not found")
		}

		if token == "" {
			log.Warnf("API token not configured")
		}

		if hostname == "" {
			OSHostname, err := os.Hostname()
			if err != nil {
				log.Fatalf("Could not get hostname from OS and hostname not set: %q", err)
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
	request := ddns.Request{
		Token: token,
		FQDN:  hostname,
	}

	ip, err := GetIP(external)
	if err != nil {
		return err
	}

	request.IP = ip

	return sendRequest(request)
}

// sendRequest fires a ddns.Request in the form of a cloud event to a receiver
func sendRequest(request ddns.Request) error {
	event := ddns.GenerateCloudEventRequest(request)
	log.Debugf("Raw cloudevent:\n%v", event)

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("Failed to create client, %v", err)
	}

	ctx := cloudevents.ContextWithTarget(context.Background(), endpoint)
	response, result := c.Request(ctx, event)
	log.WithFields(log.Fields{
		"endpoint": endpoint,
	}).Debugf("Dispatched cloudevent")

	if cloudevents.IsUndelivered(result) || cloudevents.IsNACK(result) {
		log.Fatalf("failed to deliver cloudevent, %v", result)
	}

	log.Debugf("raw cloudevent:\n%v", response)
	if response.Type() != ddns.CloudEventResponseType {
		log.Fatalf("Response was not of the expected type, expected %s, got: %s", ddns.CloudEventResponseType, response.Type())
	}

	ddnsResponse := ddns.Response{}
	if err := response.DataAs(&ddnsResponse); err != nil {
		log.Fatalf("Could not decode response: %v", err.Error())
	}

	log.WithFields(log.Fields{
		"fqdn":   ddnsResponse.FQDN,
		"ip":     request.IP,
		"status": ddnsResponse.Status,
	}).Infof("dns update requested")

	return nil
}

// GetIP returns the program's IP address, internal or external
// Uses a custom dialer to get the appropriate interface with internet access
// Uses Cloudflare's whoami.cloudflare TXT record cheatcode to get the external IP address
// dig -4 ch txt whoami.cloudflare @1.1.1.1
func GetIP(external bool) (string, error) {
	var ip string
	if external {
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
			ip = t.Txt[0]
		}
		return "", err
	} else {
		// mock a connection, this does not make a request
		conn, err := net.Dial("udp", "1.1.1.1:53")
		if err != nil {
			return "", err
		}
		ip = strings.Split(conn.LocalAddr().String(), ":")[0]
	}

	if net.ParseIP(ip) == nil {
		log.Fatalf("Could not determine IP address, got: %s", ip)
	}
	return ip, nil
}
