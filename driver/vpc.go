package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/huaweicloud/golangsdk"
	huaweisdk "github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/vpcs"
	"time"
)

// InitVPCService initializes VPC v1 service
func (c *Client) InitVPCService(d *Driver) error {
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

// CreateVPC creates new VPC by given VPC name
func (c *Client) CreateVPC(d *Driver) (string, error) {
	result, err := vpcs.Create(c.VPC, vpcs.CreateOpts{
		Name: d.VpcName,
		CIDR: "192.168.0.0/24",
	}).Extract()

	vpcID := ""
	if result != nil {
		vpcID = result.ID
	}
	return vpcID, err

}

// GetVPCDetails returns details of VPC
func (c *Client) GetVPCDetails(d *Driver) (*vpcs.Vpc, error) {
	result, err := vpcs.Get(c.VPC, d.VpcID).Extract()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveVpcID resolves `Driver.VpcId` for given `Driver.VpcName`
func (c *Client) ResolveVpcID(d *Driver) error {
	if d.VpcID != "" {
		return nil
	}
	vpcList, err := vpcs.List(c.VPC, vpcs.ListOpts{
		Name: d.VpcName,
	})
	if err != nil {
		return err
	}
	if len(vpcList) > 1 {
		return fmt.Errorf("multiple VPC found by name %s. Please provide VPC ID instead", d.VpcName)
	}

	d.VpcID = vpcList[0].ID
	return nil
}

// WaitForVPCStatus waits until VPC is in given status
func (c *Client) WaitForVPCStatus(d *Driver, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (b bool, err error) {
		cur, err := c.GetVPCDetails(d)
		if err != nil {
			return true, err
		}
		if cur.Status == "ERROR" {
			return true, fmt.Errorf("VPC creation failed. Instance `%s` is in ERROR state", d.VpcID)
		}
		if cur.Status == status {
			return false, nil
		}
		return false, nil
	}, 50, 5*time.Second)
}

// DeleteVPC removes existing VPC
func (c *Client) DeleteVPC(d *Driver) error {
	return vpcs.Delete(c.VPC, d.VpcID).Err
}
