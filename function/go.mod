module github.com/jonpulsifer/ddnsb0t/function

go 1.16

// replace github.com/jonpulsifer/ddnsb0t => ../

require (
	cloud.google.com/go/compute v1.2.0
	github.com/GoogleCloudPlatform/functions-framework-go v1.5.2
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/jonpulsifer/ddnsb0t v0.0.0-20210206181125-b46f55886693
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20220127074510-2fabfed7e28f // indirect
	google.golang.org/api v0.66.0
)
