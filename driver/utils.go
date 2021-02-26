package opentelekomcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/docker/machine/libmachine/state"
	"github.com/opentelekomcloud-infra/crutch-house/services"
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
	defaultFlavor        = "s2.large.2"
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

// logHttp500 appends error message with response 500 body
func logHttp500(err error) error {
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
			return fmt.Errorf("failed to find VPC by name: %s", logHttp500(err))
		}
		d.VpcID = managedSting{Value: vpcID}
	}

	if d.SubnetID.Value == "" && d.SubnetName != "" {
		subnetID, err := d.client.FindSubnet(d.VpcID.Value, d.SubnetName)
		if err != nil {
			return fmt.Errorf("failed to find subnet by name: %s", logHttp500(err))
		}
		d.SubnetID = managedSting{Value: subnetID}
	}

	if d.FlavorID == "" && d.FlavorName != "" {
		flavID, err := d.client.FindFlavor(d.FlavorName)
		if err != nil {
			return fmt.Errorf("fail when searching flavor by name: %s", logHttp500(err))
		}
		if flavID == "" {
			return fmt.Errorf(notFound, "flavor", d.FlavorName)
		}
		d.FlavorID = flavID
	}
	if d.RootVolumeOpts.SourceID == "" && d.ImageName != "" {
		imageID, err := d.client.FindImage(d.ImageName)
		if err != nil {
			return fmt.Errorf("failed to find image by name: %s", logHttp500(err))
		}
		if imageID == "" {
			return fmt.Errorf(notFound, "image", d.ImageName)
		}
		d.RootVolumeOpts.SourceID = imageID
	}
	sgIDs, err := d.client.FindSecurityGroups(d.SecurityGroups)
	if err != nil {
		return fmt.Errorf("failed to resolve security group IDs: %s", logHttp500(err))
	}
	d.SecurityGroupIDs = sgIDs

	if d.ServerGroupID == "" && d.ServerGroup != "" {
		serverGroupID, err := d.client.FindServerGroup(d.ServerGroup)
		if err != nil {
			return fmt.Errorf("failed to resolve server group: %s", logHttp500(err))
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
	overrideJson, err := json.Marshal(fallback)
	if err != nil {
		return nil, err
	}
	cloudJson, err := json.Marshal(cloud)
	if err != nil {
		return nil, err
	}
	var fallbackInterface interface{}
	err = json.Unmarshal(overrideJson, &fallbackInterface)
	if err != nil {
		return nil, err
	}
	var cloudInterface interface{}
	err = json.Unmarshal(cloudJson, &cloudInterface)
	if err != nil {
		return nil, err
	}
	mergedCloud := new(openstack.Cloud)
	mergedInterface := utils.MergeInterfaces(cloudInterface, fallbackInterface)
	mergedJson, err := json.Marshal(mergedInterface)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(mergedJson, mergedCloud); err != nil {
		return nil, err
	}
	return mergedCloud, nil
}

func (d *Driver) getUserData() error {
	if d.UserDataFile == "" || len(d.UserData) != 0 {
		return nil
	}
	userData, err := ioutil.ReadFile(d.UserDataFile)
	if err != nil {
		return fmt.Errorf("failed to load user data file: %s", err)
	}
	d.UserData = userData
	return nil
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = defaultSSHPort
	}
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}
	return d.SSHUser
}

func (d *Driver) GetIP() (string, error) {
	d.IPAddress = d.FloatingIP.Value
	return d.BaseDriver.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(dockerPort))), nil
}

func (d *Driver) GetState() (state.State, error) {
	if err := d.initComputeV2(); err != nil {
		return state.None, err
	}
	instance, err := d.client.GetInstanceStatus(d.InstanceID)
	if err != nil {
		return state.None, fmt.Errorf("failed to get instance state: %s", logHttp500(err))
	}
	switch instance.Status {
	case services.InstanceStatusRunning:
		return state.Running, nil
	case "PAUSED":
		return state.Paused, nil
	case services.InstanceStatusStopped:
		return state.Stopped, nil
	case "BUILDING":
		return state.Starting, nil
	case "ERROR":
		return state.Error, nil
	default:
		return state.None, nil
	}
}
