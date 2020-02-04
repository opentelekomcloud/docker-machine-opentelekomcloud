package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/huaweicloud/golangsdk"
	huaweisdk "github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/subnets"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/vpcs"
	"time"
)

const (
	vpcCIDR      = "192.168.0.0/24"
	subnetCIDR   = "192.168.0.0/26"
	primaryDNS   = "100.125.4.25"
	secondaryDNS = "8.8.8.8"
)

var defaultDNS = []string{primaryDNS, secondaryDNS}

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

// CreateVPC creates new VPC by d.VpcName
func (c *Client) CreateVPC(d *Driver) error {
	result, err := vpcs.Create(c.VPC, vpcs.CreateOpts{
		Name: d.VpcName,
		CIDR: vpcCIDR,
	}).Extract()

	if err != nil {
		return err
	}
	d.VpcID = result.ID
	d.VpcName = result.Name
	return nil
}

// GetVPCDetails returns details of VPC
func (c *Client) GetVPCDetails(d *Driver) (*vpcs.Vpc, error) {
	result, err := vpcs.Get(c.VPC, d.VpcID).Extract()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResolveVpcID resolves `Driver.VpcID` for given `Driver.VpcName`
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

// CreateSubnet creates new Subnet and set Driver.SubnetID
func (c *Client) CreateSubnet(d *Driver) error {
	result, err := subnets.Create(c.VPC, subnets.CreateOpts{
		Name:    d.VpcName,
		CIDR:    subnetCIDR,
		DnsList: defaultDNS,
	}).Extract()
	if err != nil {
		return err
	}
	d.SubnetID = result.ID
	return nil
}

// ResolveVpcSubnet resolves `Driver.SubnetID` for given `Driver.SubnetName`
func (c *Client) ResolveVpcSubnet(d *Driver) error {
	if d.SubnetID != "" {
		return nil
	}
	err := c.ResolveVpcID(d)
	if err != nil {
		return err
	}
	subnetList, err := subnets.List(c.VPC, subnets.ListOpts{
		Name:   d.SubnetName,
		VPC_ID: d.VpcID,
	})
	if err != nil {
		return err
	}
	if len(subnetList) > 1 {
		return fmt.Errorf("multiple Subnets found by name %s in VPC %s. "+
			"Please provide Subnet ID instead", d.SubnetName, d.VpcID)
	}
	d.SubnetName = subnetList[0].Name
	return nil
}
