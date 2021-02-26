package opentelekomcloud

import (
	"encoding/json"
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/opentelekomcloud-infra/crutch-house/services"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/utils"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
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

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "otc-cloud",
			EnvVar: "OS_CLOUD",
			Usage:  "Name of cloud in `clouds.yaml` file",
		},
		mcnflag.StringFlag{
			Name:   "otc-auth-url",
			EnvVar: "OS_AUTH_URL",
			Usage:  "OpenTelekomCloud authentication URL",
			Value:  defaultAuthURL,
		},
		mcnflag.StringFlag{
			Name:   "otc-cacert",
			EnvVar: "OS_CACERT",
			Usage:  "CA certificate bundle to verify against",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-id",
			EnvVar: "OS_DOMAIN_ID",
			Usage:  "OpenTelekomCloud domain ID",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-name",
			EnvVar: "OS_DOMAIN_NAME",
			Usage:  "OpenTelekomCloud domain name",
		},
		mcnflag.StringFlag{
			Name:   "otc-username",
			EnvVar: "OS_USERNAME",
			Usage:  "OpenTelekomCloud username",
		},
		mcnflag.StringFlag{
			Name:   "otc-password",
			EnvVar: "OS_PASSWORD",
			Usage:  "OpenTelekomCloud password",
		},
		mcnflag.StringFlag{
			Name:   "otc-project-name",
			EnvVar: "OS_PROJECT_NAME",
			Usage:  "OpenTelekomCloud project name",
		},
		mcnflag.StringFlag{
			Name:   "otc-project-id",
			EnvVar: "OS_PROJECT_ID",
			Usage:  "OpenTelekomCloud project ID",
		},
		mcnflag.StringFlag{
			Name:   "otc-region",
			EnvVar: "REGION",
			Usage:  "OpenTelekomCloud region name",
			Value:  defaultRegion,
		},
		mcnflag.StringFlag{
			Name:   "otc-access-key-id",
			Usage:  "OpenTelekomCloud access key ID for AK/SK auth",
			EnvVar: "ACCESS_KEY_ID",
		},
		mcnflag.StringFlag{
			Name:   "otc-access-key-key",
			Usage:  "OpenTelekomCloud secret access key for AK/SK auth",
			EnvVar: "ACCESS_KEY_SECRET",
		},
		mcnflag.StringFlag{
			Name:   "otc-availability-zone",
			EnvVar: "OS_AVAILABILITY_ZONE",
			Usage:  "OpenTelekomCloud availability zone",
			Value:  defaultAZ,
		},
		mcnflag.StringFlag{
			Name:   "otc-flavor-id",
			EnvVar: "FLAVOR_ID",
			Usage:  "OpenTelekomCloud flavor id to use for the instance",
		},
		mcnflag.StringFlag{
			Name:   "otc-flavor-name",
			EnvVar: "OS_FLAVOR_NAME",
			Usage:  "OpenTelekomCloud flavor name to use for the instance",
			Value:  defaultFlavor,
		},
		mcnflag.StringFlag{
			Name:   "otc-image-id",
			EnvVar: "OS_IMAGE_ID",
			Usage:  "OpenTelekomCloud image id to use for the instance",
		},
		mcnflag.StringFlag{
			Name:   "otc-image-name",
			EnvVar: "OS_IMAGE_NAME",
			Usage:  "OpenTelekomCloud image name to use for the instance",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			Name:   "otc-keypair-name",
			EnvVar: "OS_KEYPAIR_NAME",
			Usage:  "OpenTelekomCloud keypair to use to SSH to the instance",
		},
		mcnflag.StringFlag{
			Name:   "otc-vpc-id",
			EnvVar: "OS_VPC_ID",
			Usage:  "OpenTelekomCloud VPC id the machine will be connected on",
		},
		mcnflag.StringFlag{
			Name:   "otc-vpc-name",
			EnvVar: "OS_VPC_NAME",
			Usage:  "OpenTelekomCloud VPC name the machine will be connected on",
			Value:  defaultVpcName,
		},
		mcnflag.StringFlag{
			Name:   "otc-subnet-id",
			EnvVar: "OS_NETWORK_ID",
			Usage:  "OpenTelekomCloud subnet id the machine will be connected on",
		},
		mcnflag.StringFlag{
			Name:   "otc-subnet-name",
			EnvVar: "OS_NETWORK_NAME",
			Usage:  "OpenTelekomCloud subnet name the machine will be connected on",
			Value:  defaultSubnetName,
		},
		mcnflag.StringFlag{
			Name:   "otc-private-key-file",
			EnvVar: "OS_PRIVATE_KEY_FILE",
			Usage:  "Private key file to use for SSH (absolute path)",
		},
		mcnflag.StringFlag{
			Name:   "otc-user-data-file",
			EnvVar: "OS_USER_DATA_FILE",
			Usage:  "File containing an user data script",
		},
		mcnflag.StringFlag{
			Name:  "otc-user-data-raw",
			Usage: "Contents of user data file as a string",
		},
		mcnflag.StringFlag{
			Name:   "otc-token",
			EnvVar: "OS_TOKEN",
			Usage:  "OpenTelekomCloud authorization token",
		},
		mcnflag.StringFlag{
			Name:   "otc-sec-groups",
			EnvVar: "OS_SECURITY_GROUP",
			Usage:  "Existing security groups to use, separated by comma",
		},
		mcnflag.StringFlag{
			Name:   "otc-floating-ip",
			EnvVar: "OS_FLOATING_IP",
			Usage:  "OpenTelekomCloud floating IP to use",
		},
		mcnflag.StringFlag{
			Name:   "otc-floating-ip-type",
			EnvVar: "OS_FLOATING_IP_TYPE",
			Usage:  "OpenTelekomCloud bandwidth type",
			Value:  "5_bgp",
		},
		mcnflag.IntFlag{
			Name:   "otc-bandwidth-size",
			EnvVar: "OS_BANDWIDTH_SIZE",
			Usage:  "OpenTelekomCloud bandwidth size",
			Value:  100,
		},
		mcnflag.StringFlag{
			Name:   "otc-bandwidth-type",
			EnvVar: "OS_BANDWIDTH_TYPE",
			Usage:  "OpenTelekomCloud bandwidth share type",
			Value:  "PER",
		},
		mcnflag.BoolFlag{
			Name:  "otc-skip-ip",
			Usage: "If set, elastic IP won't be created",
		},
		mcnflag.IntFlag{
			Name:   "otc-ip-version",
			EnvVar: "OS_IP_VERSION",
			Usage:  "OpenTelekomCloud version of IP address assigned for the machine",
			Value:  4,
		},
		mcnflag.StringFlag{
			Name:   "otc-ssh-user",
			EnvVar: "OS_SSH_USER",
			Usage:  "Machine SSH username",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			Name:   "otc-ssh-port",
			EnvVar: "OS_SSH_PORT",
			Usage:  "Machine SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			Name:   "otc-endpoint-type",
			EnvVar: "OS_INTERFACE",
			Usage:  "OpenTelekomCloud interface (endpoint) type",
			Value:  "public",
		},
		mcnflag.BoolFlag{
			Name:  "otc-skip-default-sg",
			Usage: "Don't create default security group",
		},
		mcnflag.StringFlag{
			Name:   "otc-server-group",
			EnvVar: "OS_SERVER_GROUP",
			Usage:  "Define server group where server will be created",
		},
		mcnflag.StringFlag{
			Name:   "otc-server-group-id",
			EnvVar: "OS_SERVER_GROUP_ID",
			Usage:  "Define server group where server will be created by ID",
		},
		mcnflag.IntFlag{
			Name:   "otc-root-volume-size",
			EnvVar: "OS_ROOT_VOLUME_SIZE",
			Usage:  "Set volume size of root partition",
			Value:  defaultVolumeSize,
		},
		mcnflag.StringFlag{
			Name:   "otc-root-volume-type",
			EnvVar: "OS_ROOT_VOLUME_TYPE",
			Usage:  "Set volume type of root partition (one of SATA, SAS, SSD)",
			Value:  defaultVolumeType,
		},
		mcnflag.StringFlag{
			Name:   "otc-tags",
			EnvVar: "OS_TAGS",
			Usage:  "Comma-separated list of instance tags",
		},
	}
}

// SetConfigFromFlags loads driver configuration from given flags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthURL = flags.String("otc-auth-url")
	d.Cloud = flags.String("otc-cloud")
	d.CACert = flags.String("otc-cacert")
	d.DomainID = flags.String("otc-domain-id")
	d.DomainName = flags.String("otc-domain-name")
	d.Username = flags.String("otc-username")
	d.Password = flags.String("otc-password")
	d.ProjectName = flags.String("otc-project-name")
	projectID := flags.String("otc-tenant-id")
	if projectID == "" {
		projectID = flags.String("otc-project-id")
	}
	d.ProjectID = projectID
	d.Region = flags.String("otc-region")
	d.EndpointType = flags.String("otc-endpoint-type")
	d.FlavorID = flags.String("otc-flavor-id")
	d.FlavorName = flags.String("otc-flavor-name")
	d.ImageName = flags.String("otc-image-name")
	d.VpcID = managedSting{Value: flags.String("otc-vpc-id")}
	d.VpcName = flags.String("otc-vpc-name")
	d.SubnetID = managedSting{Value: flags.String("otc-subnet-id")}
	d.SubnetName = flags.String("otc-subnet-name")
	d.FloatingIP = managedSting{Value: flags.String("otc-floating-ip")}
	d.IPVersion = flags.Int("otc-ip-version")
	d.SSHUser = flags.String("otc-ssh-user")
	d.SSHPort = flags.Int("otc-ssh-port")
	d.KeyPairName = managedSting{Value: flags.String("otc-keypair-name")}
	d.PrivateKeyFile = flags.String("otc-private-key-file")
	d.Token = flags.String("otc-token")
	d.UserDataFile = flags.String("otc-user-data-file")
	d.UserData = []byte(flags.String("otc-user-data-raw"))
	d.ServerGroup = flags.String("otc-server-group")
	d.ServerGroupID = flags.String("otc-server-group-id")
	tags := flags.String("otc-tags")
	if tags != "" {
		d.Tags = strings.Split(tags, ",")
	}
	d.AccessKey = flags.String("otc-access-key-id")
	d.SecretKey = flags.String("otc-access-key-key")

	d.RootVolumeOpts = &services.DiskOpts{
		SourceID: flags.String("otc-image-id"),
		Size:     flags.Int("otc-root-volume-size"),
		Type:     flags.String("otc-root-volume-type"),
	}
	ipType := flags.String("otc-elastic-ip-type")
	if ipType == "" {
		ipType = flags.String("otc-floating-ip-type")
	}

	d.eipConfig = &services.ElasticIPOpts{
		IPType:        ipType,
		BandwidthSize: flags.Int("otc-bandwidth-size"),
		BandwidthType: flags.String("otc-bandwidth-type"),
	}
	d.skipEIPCreation = flags.Int("otc-elastic-ip") == 0 || flags.Bool("otc-skip-ip")

	az := flags.String("otc-available-zone")
	if az == "" {
		az = flags.String("otc-availability-zone")
	}
	d.AvailabilityZone = az

	if sg := flags.String("otc-sec-groups"); sg != "" {
		d.SecurityGroups = strings.Split(sg, ",")
	}

	if !flags.Bool("otc-skip-default-sg") {
		d.ManagedSecurityGroup = defaultSecurityGroup
	}

	d.SetSwarmConfigFromFlags(flags)
	return d.checkConfig()
}
