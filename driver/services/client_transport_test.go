package services

import (
	"net/http"
	"testing"
	"time"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResetHTTPTransportReplacesTransportWithClone(t *testing.T) {
	original := &http.Transport{IdleConnTimeout: 17 * time.Second}
	client := &Client{
		Provider: &golangsdk.ProviderClient{
			HTTPClient: http.Client{Transport: original},
		},
	}

	client.ResetHTTPTransport()

	replacement, ok := client.Provider.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotSame(t, original, replacement)
	assert.Equal(t, original.IdleConnTimeout, replacement.IdleConnTimeout)
}

func TestResetHTTPTransportReplacesDefaultTransport(t *testing.T) {
	client := &Client{Provider: &golangsdk.ProviderClient{}}

	client.ResetHTTPTransport()

	transport, ok := client.Provider.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.NotSame(t, http.DefaultTransport, transport)
}

func TestResetHTTPTransportHandlesNilClient(t *testing.T) {
	var client *Client
	assert.NotPanics(t, client.ResetHTTPTransport)

	client = &Client{}
	assert.NotPanics(t, client.ResetHTTPTransport)
}
