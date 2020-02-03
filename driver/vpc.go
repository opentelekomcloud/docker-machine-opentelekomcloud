package opentelekomcloud

import (
	"github.com/huaweicloud/golangsdk"
	huaweisdk "github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/vpcs"
)

func (c *Client) initializeVPCService(d *Driver) error {
	if c.VPC != nil {
		return nil
	}
	vpc, err := huaweisdk.NewNetworkV1(c.OTCProvider, golangsdk.EndpointOpts{
		Region:       d.Region,
		Availability: golangsdk.Availability(getEndpointType(d.EndpointType)),
	})
	if err != nil {
		return err
	}
	c.VPC = vpc
	return nil
}

func (c *Client) CreateVPC(d *Driver) (string, error) {
	if err := c.initializeVPCService(d); err != nil {
		return "", err
	}

	result, err := vpcs.Create(c.VPC, vpcs.CreateOpts{
		Name: d.SubnetName,
		CIDR: "192.168.0.0/24",
	}).Extract()

	vpcID := ""
	if result != nil {
		vpcID = result.ID
	}
	return vpcID, err

}
