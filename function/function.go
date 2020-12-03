package function

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/jonpulsifer/ddnsb0t/pkg/ddns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	dns "google.golang.org/api/dns/v1"
)

var (
	client      *dns.Service
	debug       string = os.Getenv("DDNS_DEBUG")
	environment string = os.Getenv("DDNS_ENVIRONMENT")
	project     string = os.Getenv("DDNS_GCP_PROJECT")
	token       string = os.Getenv("DDNS_API_TOKEN")
)

func init() {
	log.SetLevel(log.InfoLevel)
	if debug != "" {
		log.SetLevel(log.DebugLevel)
	}
	if environment == "production" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	var err error
	client, err = dns.NewService(context.Background())
	if err != nil {
		log.Fatalf("could not build cloud dns client: %v", err.Error())
	}

	if project == "" {
		projectFromMetadata, err := metadata.ProjectID()
		if err == nil {
			project = projectFromMetadata
		} else {
			log.Fatalf("could not determine GCP project: %v", err)
		}
	}

	if token == "" {
		log.Warnf("token is not set, unauthenticated requests enabled")
	}
}

// DDNSCloudEventReceiver is an HTTP Cloud Event receiver expecting events of both
// ddns.CloudEventRequestType and ddns.CloudEventResponseType
func DDNSCloudEventReceiver(w http.ResponseWriter, r *http.Request) {
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	event := cloudevents.NewEvent()
	if err := event.UnmarshalJSON(respBody); err != nil {
		log.Errorf("Could not convert to cloudevent: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if event.Type() != ddns.CloudEventRequestType {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}

	log.Debugf("raw cloudevent request:\n%v", event)

	var request ddns.Request
	if err := event.DataAs(&request); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if token != "" && token != request.Token {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	change, err := updateDNS(&request)
	if err != nil {
		log.Errorf("could not update dns record: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ddnsResponse := ddns.Response{
		FQDN:      request.FQDN,
		Status:    change.Status,
		Additions: len(change.Additions),
		Deletions: len(change.Deletions),
	}

	response := ddns.GenerateCloudEventResponse(ddnsResponse)
	responseBytes, err := response.MarshalJSON()
	if err != nil {
		log.Errorf("could not encode response: %v", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	w.Write(responseBytes)
	log.WithFields(log.Fields{
		"ce-id":     response.ID(),
		"ce-source": response.Source(),
		"ce-type":   response.Type(),
		"ddns-fqdn": request.FQDN,
		"ddns-ip":   request.IP,
	}).Infof("processed cloudevent")
	log.Debugf("raw cloudevent response:\n%v", event)

	// https://github.com/GoogleCloudPlatform/functions-framework-go/issues/58

	// response := cloudevents.NewEvent(cloudevents.VersionV1)
	// response.SetType("dev.pulsifer.ddns.response")
	// response.SetSource("https://github.com/jonpulsifer/ddnsb0t/function")
	// if err := response.SetData(cloudevents.ApplicationJSON, ddnsResponse); err != nil {
	// 	return cloudevents.NewHTTPResult(500, "failed to set response data: %s", err)
	// }

	// return &response, nil
}

func updateDNS(request *ddns.Request) (*dns.Change, error) {
	var (
		additions  []*dns.ResourceRecordSet
		deletions  []*dns.ResourceRecordSet
		change     dns.Change
		dnsRequest = &dns.ResourceRecordSet{
			Name:    request.FQDN,
			Type:    "A",
			Ttl:     60,
			Rrdatas: []string{request.IP},
		}
	)

	managedZone, err := getManagedZoneFromDNSName(request.FQDN)
	if err != nil {
		return nil, err
	}

	// get the current records for the requested endpoint
	resp, err := client.ResourceRecordSets.List(project, managedZone.Name).Do()
	if err != nil {
		return nil, err
	}

	for _, recordset := range resp.Rrsets {
		if recordset.Name == request.FQDN && recordset.Type == "A" {
			deletions = append(deletions, recordset)
		}
	}
	additions = append(additions, dnsRequest)

	log.WithFields(log.Fields{
		"additions": len(additions),
		"deletions": len(deletions),
		"endpoint":  request.FQDN,
		"ip":        request.IP,
	}).Debugf("Preparing DNS change")

	if len(deletions) == 0 {
		change = dns.Change{
			Additions: additions,
		}
	} else {
		change = dns.Change{
			Additions: additions,
			Deletions: deletions,
		}
	}

	return client.Changes.Create(project, managedZone.Name, &change).Context(context.Background()).Do()
}

func getManagedZoneFromDNSName(DNSName string) (*dns.ManagedZone, error) {
	managedZoneList, err := client.ManagedZones.List(project).Do()
	if err != nil {
		return nil, err
	}

	for _, managedZone := range managedZoneList.ManagedZones {
		if strings.HasSuffix(DNSName, managedZone.DnsName) {
			return managedZone, nil
		}
	}
	return nil, errors.New("No managed zone found")
}
