package opentelekomcloud

import (
	"encoding/json"
	"fmt"
	"os"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/utils"
)

const (
	errorBothOptions     = "both %s and %s must be specified"
	notFound             = "%s not found by name `%s`"
	driverName           = "otc"
	dockerPort           = 2376
	defaultSecurityGroup = "docker-machine-grp"
	defaultAZ            = "eu-de-01"
	defaultFlavor        = "s3.xlarge.2"
	defaultImage         = "Standard_Ubuntu_20.04_latest"
	defaultSSHUser       = "ubuntu"
	defaultSSHPort       = 22
	defaultRegion        = "eu-de"
	defaultAuthURL       = "https://iam.eu-de.otc.t-systems.com/v3"
	defaultVpcName       = "vpc-docker-machine"
	defaultSubnetName    = "subnet-docker-machine"
	defaultVolumeSize    = 40
	defaultVolumeType    = "SSD"
)

// logHTTP500 appends error message with response 500 body
func logHTTP500(err error) error {
	if e, ok := err.(golangsdk.ErrDefault500); ok {
		return fmt.Errorf("%s: %s", e, string(e.Body))
	}
	return err
}

// resolveIDs resolves name to IDs where possible
func (d *Driver) resolveIDs() error {
	if d.VpcID.Value == "" && d.VpcName != "" {
		vpcID, err := d.client.FindVPC(d.VpcName)
		if err != nil {
			return fmt.Errorf("failed to find VPC by name: %s", logHTTP500(err))
		}
		d.VpcID = managedSting{Value: vpcID}
	}

	if d.SubnetID.Value == "" && d.SubnetName != "" {
		subnetID, err := d.client.FindSubnet(d.VpcID.Value, d.SubnetName)
		if err != nil {
			return fmt.Errorf("failed to find subnet by name: %s", logHTTP500(err))
		}
		d.SubnetID = managedSting{Value: subnetID}
	}

	if d.FlavorID == "" && d.FlavorName != "" {
		flavorID, err := d.client.FindFlavor(d.FlavorName)
		if err != nil {
			return fmt.Errorf("fail when searching flavor by name: %s", logHTTP500(err))
		}
		if flavorID == "" {
			return fmt.Errorf(notFound, "flavor", d.FlavorName)
		}
		d.FlavorID = flavorID
	}
	if d.RootVolumeOpts.SourceID == "" && d.ImageName != "" {
		imageID, err := d.client.FindImage(d.ImageName)
		if err != nil {
			return fmt.Errorf("failed to find image by name: %s", logHTTP500(err))
		}
		if imageID == "" {
			return fmt.Errorf(notFound, "image", d.ImageName)
		}
		d.RootVolumeOpts.SourceID = imageID
	}
	sgIDs, err := d.client.FindSecurityGroups(d.SecurityGroups)
	if err != nil {
		return fmt.Errorf("failed to resolve security group IDs: %s", logHTTP500(err))
	}
	d.SecurityGroupIDs = sgIDs

	if d.ServerGroupID == "" && d.ServerGroup != "" {
		serverGroupID, err := d.client.FindServerGroup(d.ServerGroup)
		if err != nil {
			return fmt.Errorf("failed to resolve server group: %s", logHTTP500(err))
		}
		d.ServerGroupID = serverGroupID
	}

	return nil
}

func (d *Driver) checkConfig() error {
	if (d.KeyPairName.Value != "" && d.PrivateKeyFile == "") || (d.KeyPairName.Value == "" && d.PrivateKeyFile != "") {
		return fmt.Errorf(errorBothOptions, "KeyPairName", "PrivateKeyFile")
	}
	if d.Cloud == "" &&
		(d.Username == "" || d.Password == "") &&
		d.Token == "" &&
		(d.AccessKey == "" || d.SecretKey == "") {
		return fmt.Errorf("at least one authorization method must be provided")
	}
	if len(d.UserData) > 0 && d.UserDataFile != "" {
		return fmt.Errorf("both `-otc-user-data` and `-otc-user-data-file` is defined")
	}
	return nil
}

// mergeClouds merges two Config recursively (the AuthInfo also gets merged).
// In case both Config define a value, the value in the 'cloud' cloud takes precedence
func mergeClouds(cloud, fallback interface{}) (*openstack.Cloud, error) {
	overrideJSON, err := json.Marshal(fallback)
	if err != nil {
		return nil, err
	}
	cloudJSON, err := json.Marshal(cloud)
	if err != nil {
		return nil, err
	}
	var fallbackInterface interface{}
	err = json.Unmarshal(overrideJSON, &fallbackInterface)
	if err != nil {
		return nil, err
	}
	var cloudInterface interface{}
	err = json.Unmarshal(cloudJSON, &cloudInterface)
	if err != nil {
		return nil, err
	}
	mergedCloud := new(openstack.Cloud)
	mergedInterface := utils.MergeInterfaces(cloudInterface, fallbackInterface)
	mergedJSON, err := json.Marshal(mergedInterface)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(mergedJSON, mergedCloud); err != nil {
		return nil, err
	}
	return mergedCloud, nil
}

func (d *Driver) getUserData() error {
	if d.UserDataFile == "" || len(d.UserData) != 0 {
		return nil
	}
	userData, err := os.ReadFile(d.UserDataFile)
	if err != nil {
		return fmt.Errorf("failed to load user data file: %s", err)
	}
	d.UserData = userData
	return nil
}
