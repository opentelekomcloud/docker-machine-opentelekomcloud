package services

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/identity/v2/tenants"
	"github.com/huaweicloud/golangsdk/pagination"
)

func (c *Client) InitIdentity() error {
	if c.Identity != nil {
		return nil
	}

	identity, err := openstack.NewIdentityV2(c.Provider, golangsdk.EndpointOpts{
		Region:       c.region,
		Availability: c.endpointType,
	})
	if err != nil {
		return err
	}
	c.Identity = identity
	return nil
}

func (c *Client) GetTenantID(tenantName string) (string, error) {
	pager := tenants.List(c.Identity, nil)

	tenantID := ""
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		tenantList, err := tenants.ExtractTenants(page)
		if err != nil {
			return false, err
		}

		for _, i := range tenantList {
			if i.Name == tenantName {
				tenantID = i.ID
				return false, nil
			}
		}

		return true, nil
	})
	return tenantID, err
}
