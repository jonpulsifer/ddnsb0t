module github.com/jonpulsifer/ddnsb0t/function

go 1.13

// replace github.com/jonpulsifer/ddnsb0t => ../

require (
	cloud.google.com/go/compute v0.1.0
	cloud.google.com/go/functions v1.1.0 // indirect
	github.com/GoogleCloudPlatform/functions-framework-go v1.2.0
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/jonpulsifer/ddnsb0t v0.0.0-20201203003312-ffca5926b430
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420
	google.golang.org/api v0.63.0
)
