package services

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestCreateSecurityGroupUsesPerPortCIDR(t *testing.T) {
	const restrictedCIDR = "203.0.113.42/32"
	var ruleCIDRs []string

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var responseBody string
		switch r.URL.Path {
		case "/os-security-groups":
			require.Equal(t, http.MethodPost, r.Method)
			responseBody = `{"security_group":{"id":"sg-id","name":"test-sg","description":"test","rules":[],"tenant_id":"tenant-id"}}`
		case "/os-security-group-rules":
			require.Equal(t, http.MethodPost, r.Method)
			var request struct {
				Rule struct {
					CIDR     string `json:"cidr"`
					FromPort int    `json:"from_port"`
					ToPort   int    `json:"to_port"`
				} `json:"security_group_rule"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
			ruleCIDRs = append(ruleCIDRs, request.Rule.CIDR)
			require.Equal(t, request.Rule.FromPort, request.Rule.ToPort)
			responseBody = `{"security_group_rule":{"id":"rule-id"}}`
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Request:    r,
		}, nil
	})

	provider := &golangsdk.ProviderClient{HTTPClient: http.Client{Transport: transport}}
	client := &Client{ComputeV2: &golangsdk.ServiceClient{
		ProviderClient: provider,
		Endpoint:       "https://compute.example.com/",
	}}

	_, err := client.CreateSecurityGroup("test-sg",
		PortRange{From: 22, CIDR: restrictedCIDR},
		PortRange{From: 2376},
	)
	require.NoError(t, err)
	require.Equal(t, []string{restrictedCIDR, cidrAll}, ruleCIDRs)
}
