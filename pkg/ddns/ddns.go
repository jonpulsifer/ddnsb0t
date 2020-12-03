package ddns

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

// Request is the expected struct we want as a request
type Request struct {
	IP     string `json:"ip"`
	FQDN   string `json:"fqdn"`
	Domain string `json:"domain,omitempty"`
	Token  string `json:"token"`
}

// Response is the response we expect to return
type Response struct {
	FQDN      string `json:"fqdn"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions,omitempty"`
}

const (
	// CloudEventSource is the repo where this project lives
	CloudEventSource = "https://github.com/jonpulsifer/ddnsb0t"
	// CloudEventSchema is the location of the Request and Response types
	CloudEventSchema = "https://github.com/jonpulsifer/ddnsb0t/pkg/ddns/ddns.go"
	// CloudEventRequestType is the CloudEvent Request
	CloudEventRequestType = "dev.pulsifer.ddns.request"
	// CloudEventResponseType is the CloudEvent Response
	CloudEventResponseType = "dev.pulsifer.ddns.response"
)

// GenerateCloudEventRequest builds a cloudevents.Event from a ddns.Request
func GenerateCloudEventRequest(request Request) cloudevents.Event {
	ce := buildCloudEvent()
	ce.SetType(CloudEventRequestType)
	ce.SetData(cloudevents.ApplicationJSON, request)
	return ce
}

// GenerateCloudEventResponse builds a cloudevents.Event from a ddns.Response
func GenerateCloudEventResponse(response Response) cloudevents.Event {
	ce := buildCloudEvent()
	ce.SetType(CloudEventResponseType)
	ce.SetData(cloudevents.ApplicationJSON, response)
	return ce
}

func buildCloudEvent() cloudevents.Event {
	ce := cloudevents.NewEvent()
	ce.SetID(uuid.New().String())
	ce.SetSource(CloudEventSource)
	ce.SetDataSchema(CloudEventSchema)
	return ce
}
