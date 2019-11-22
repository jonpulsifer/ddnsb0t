package cloudfunction

import (
	dns "google.golang.org/api/dns/v1"
	"reflect"
	"testing"
)

func Test_cleanDNSName(t *testing.T) {
	var (
		testDomain   = "pulsifer.dev"
		testHostname = "test"
		want         = testHostname + "." + testDomain + "."
	)
	tests := []struct {
		name    string
		DNSName string
		domain  string
		want    string
		wantErr bool
	}{
		{name: "domain with trailing dot", DNSName: testHostname, domain: testDomain + ".", want: want},
		{name: "hostname only", DNSName: "test", domain: testDomain, want: want},
		{name: "hostname with trailing dot", DNSName: testHostname + ".", domain: testDomain + ".", want: want},
		{name: "no hostname", DNSName: "", domain: testDomain, want: "", wantErr: true},
		{name: "fqdn", DNSName: "test.pulsifer.dev", domain: testDomain, want: want},
		{name: "fqdn with trailing dot", DNSName: "test.pulsifer.dev.", domain: testDomain, want: want},
		{name: "short domain name", DNSName: "test.local", domain: testDomain, want: want},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanDNSName(tt.DNSName, tt.domain)
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
