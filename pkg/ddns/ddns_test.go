package ddns

import (
	"testing"
)

// TestBuildCloudEvent validates that GenerateCloudEventRequest and GenerateCloudEventResponse
// produce valid cloudevents
func TestBuildCloudEvent(t *testing.T)  {
	requests := []Request{
		{FQDN: "wat", Token: "", IP: "10.0.0.2"},
		{FQDN: "laptop.example.com", Token: "", IP: "10.0.0.2"},
	}

	for _, request := range requests {
		ce := GenerateCloudEventRequest(request)
		if err := ce.Validate(); err != nil {
			t.Errorf("GenerateCloudEventRequest produced an invalid cloudevent: %v", err)
		}
	}

	responses := []Response{
		{Additions: 0, Status: "pending", FQDN: "test.example.com", Deletions: 0 },
		{Additions: 0, Status: "lol", FQDN: "example.com", Deletions: 0},
	}

	for _, response := range responses {
		ce := GenerateCloudEventResponse(response)
		if err := ce.Validate(); err != nil {
			t.Errorf("GenerateCloudEventResponse produced an invalid cloudevent: %v", err)
		}
	}
}
