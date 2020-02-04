package opentelekomcloud

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/docker/machine/libmachine/version"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/utils/openstack/clientconfig"
	huaweisdk "github.com/huaweicloud/golangsdk"
	"io/ioutil"
	"net/http"
)

// Client contains service clients
type Client struct {
	OSProvider *gophercloud.ProviderClient
	ComputeV2  *gophercloud.ServiceClient

	OTCProvider *huaweisdk.ProviderClient
	VPC         *huaweisdk.ServiceClient
}

// Authenticate authenticate client in the cloud
func (c *Client) Authenticate(d *Driver) error {
	if c.OSProvider != nil {
		return nil
	}

	clientOpts := &clientconfig.ClientOpts{
		Cloud:      d.Cloud,
		RegionName: d.Region,
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:           d.AuthUrl,
			Username:          d.Username,
			Password:          d.Password,
			ProjectName:       d.ProjectName,
			ProjectID:         d.ProjectID,
			ProjectDomainName: d.DomainName,
			ProjectDomainID:   d.DomainID,
			DefaultDomain:     d.DomainName,
			Token:             d.Token,
		},
	}

	cloud, err := clientconfig.GetCloudFromYAML(clientOpts)
	if err != nil {
		return err
	}

	opts, err := clientconfig.AuthOptions(clientOpts)
	if err != nil {
		return err
	}

	// mimic behaviour of OTC terraform provider

	if opts.DomainID == "" {
		opts.DomainID = cloud.AuthInfo.ProjectDomainID
	}

	if opts.DomainName == "" {
		opts.DomainName = cloud.AuthInfo.ProjectDomainName
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return err
	}

	c.OSProvider = provider

	c.OSProvider.UserAgent.Prepend(fmt.Sprintf("docker-machine/v%d", version.APIVersion))

	err = openstack.Authenticate(c.OSProvider, *opts)
	if err != nil {
		return err
	}

	// Duplicate to HuaweiSDK auth options
	//hwOpts := huaweisdk.AuthOptions{
	//	IdentityEndpoint: opts.IdentityEndpoint,
	//	Username:         opts.Username,
	//	UserID:           opts.UserID,
	//	Password:         opts.Password,
	//	DomainID:         opts.DomainID,
	//	DomainName:       opts.DomainName,
	//	TenantID:         opts.TenantID,
	//	TenantName:       opts.TenantName,
	//	TokenID:          opts.TokenID,
	//}

	return nil
}

func getEndpointType(endpointType string) string {
	if endpointType == "internal" || endpointType == "internalURL" {
		return "internal"
	}
	if endpointType == "admin" || endpointType == "adminURL" {
		return "admin"
	}
	return "public"
}

// SetTLSConfig change default HTTPClient.Transport with TLS CA configuration using CACert from config
func (c *Client) SetTLSConfig(d *Driver) error {
	config := &tls.Config{}
	if d.CACert != "" {
		caCert, err := ioutil.ReadFile(d.CACert)

		if err != nil {
			return fmt.Errorf("error reading CA Cert: %s", err)
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return fmt.Errorf("can't use CA cert: %s", d.CACert)
		}
		config.RootCAs = caCertPool
	}

	if !d.ValidateCert {
		config.InsecureSkipVerify = true
	}

	c.OSProvider.HTTPClient.Transport = &http.Transport{TLSClientConfig: config}
	return nil
}
