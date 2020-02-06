package services

import (
	"fmt"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/subnets"
	"github.com/huaweicloud/golangsdk/openstack/networking/v1/vpcs"
)

const (
	vpcCIDR        = "192.168.0.0/24"
	subnetCIDR     = "192.168.0.0/26"
	primaryDNS     = "100.125.4.25"
	secondaryDNS   = "8.8.8.8"
	defaultGateway = "192.168.0.1"
)

var defaultDNS = []string{primaryDNS, secondaryDNS}

// InitNetwork initializes VPC v1 service
func (c *Client) InitNetwork() error {
	if c.VPC != nil {
		return nil
	}
	vpc, err := openstack.NewNetworkV1(c.Provider, golangsdk.EndpointOpts{
		Region:       c.region,
		Availability: c.endpointType,
	})
	if err != nil {
		return err
	}
	c.VPC = vpc
	return nil
}

// CreateVPC creates new VPC by d.VpcName
func (c *Client) CreateVPC(vpcName string) (*vpcs.Vpc, error) {
	return vpcs.Create(c.VPC, vpcs.CreateOpts{
		Name: vpcName,
		CIDR: vpcCIDR,
	}).Extract()
}

// GetVPCDetails returns details of VPC
func (c *Client) GetVPCDetails(vpcID string) (*vpcs.Vpc, error) {
	return vpcs.Get(c.VPC, vpcID).Extract()
}

// FindVPC find VPC in list by its name and return VPC ID
func (c *Client) FindVPC(vpcName string) (string, error) {
	opts := vpcs.ListOpts{
		Name: vpcName,
	}
	vpcList, err := vpcs.List(c.VPC, opts)
	if err != nil {
		return "", err
	}
	if len(vpcList) == 0 {
		return "", nil
	}
	if len(vpcList) > 1 {
		return "", fmt.Errorf("multiple VPC found by name %s. Please provide VPC ID instead", vpcName)
	}
	return vpcList[0].ID, nil
}

// WaitForVPCStatus waits until VPC is in given status
func (c *Client) WaitForVPCStatus(vpcID, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (b bool, err error) {
		cur, err := c.GetVPCDetails(vpcID)
		if err != nil {
			return true, err
		}
		if cur.Status == "ERROR" {
			return true, fmt.Errorf("VPC creation failed. Instance `%s` is in ERROR state", vpcID)
		}
		if cur.Status == status {
			return false, nil
		}
		return false, nil
	}, maxAttempts, waitInterval)
}

// DeleteVPC removes existing VPC
func (c *Client) DeleteVPC(vpcID string) error {
	return vpcs.Delete(c.VPC, vpcID).Err
}

// CreateSubnet creates new Subnet and set Driver.SubnetID
func (c *Client) CreateSubnet(vpcID string, subnetName string) (*subnets.Subnet, error) {
	return subnets.Create(c.VPC, subnets.CreateOpts{
		VPC_ID:    vpcID,
		Name:      subnetName,
		CIDR:      subnetCIDR,
		DnsList:   defaultDNS,
		GatewayIP: defaultGateway,
	},
	).Extract()
}

// FindSubnet find subnet by name in given VPC and return ID
func (c *Client) FindSubnet(vpcID string, subnetName string) (string, error) {
	subnetList, err := subnets.List(c.VPC, subnets.ListOpts{
		Name:   subnetName,
		VPC_ID: vpcID,
	})
	if err != nil {
		return "", err
	}
	if len(subnetList) == 0 {
		return "", nil
	}
	if len(subnetList) > 1 {
		return "", fmt.Errorf("multiple Subnets found by name %s in VPC %s. "+
			"Please provide Subnet ID instead", subnetName, vpcID)
	}
	return subnetList[0].ID, nil
}

func (c *Client) GetSubnetStatus(subnetID string) (*subnets.Subnet, error) {
	return subnets.Get(c.VPC, subnetID).Extract()
}

// WaitForSubnetStatus waits for subnet to be in given status
func (c *Client) WaitForSubnetStatus(subnetID string, status string) error {
	return mcnutils.WaitForSpecificOrError(func() (b bool, err error) {
		curStatus, err := c.GetSubnetStatus(subnetID)
		if err != nil {
			return true, err
		}
		if curStatus.Status == "ERROR" {
			return true, fmt.Errorf("subnet `%s` is in error status", subnetID)
		}
		if curStatus.Status == status {
			return true, nil
		}
		return false, nil
	}, maxAttempts, waitInterval)
}

// DeleteSubnet removes subnet from VPC
func (c *Client) DeleteSubnet(vpcID string, subnetID string) error {
	return subnets.Delete(c.VPC, vpcID, subnetID).Err
}
