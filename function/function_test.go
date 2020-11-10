package cloudfunction

import (
	"reflect"
	"testing"

	dns "google.golang.org/api/dns/v1"
)

func Test_cleanDNSName(t *testing.T) {
	var (
		testDomain = "pulsifer.dev"
		want       = testDomain + "."
	)
	tests := []struct {
		name    string
		DNSName string
		want    string
		wantErr bool
	}{
		{name: "hostname only", DNSName: testDomain, want: want},
		{name: "hostname with trailing dot", DNSName: testDomain + ".", want: want},
		{name: "only hostname", DNSName: "test", want: "", wantErr: true},
		{name: "no hostname", DNSName: "", want: "", wantErr: true},
		{name: "fqdn", DNSName: "test." + testDomain, want: "test." + want},
		{name: "fqdn with trailing dot", DNSName: "test." + testDomain + ".", want: "test." + want},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanDNSName(tt.DNSName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getManagedZoneFromDNSName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanDNSName() = %v, want %v", got, tt.want)
			}
		})
	}

}

func Test_getManagedZoneFromDNSName(t *testing.T) {
	type args struct {
		c       *dns.Service
		project string
		DNSName string
	}
	tests := []struct {
		name    string
		args    args
		want    *dns.ManagedZone
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getManagedZoneFromDNSName(tt.args.c, tt.args.project, tt.args.DNSName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getManagedZoneFromDNSName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getManagedZoneFromDNSName() = %v, want %v", got, tt.want)
			}
		})
	}
}
