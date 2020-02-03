package opentelekomcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"net"
)

const (
	driverName           = "otc"
	defaultSecurityGroup = "docker-machine-grp"
	defaultAZ            = "eu-de-03"
	defaultFlavor        = "s2.large.2"
	defaultImage         = "Standard_Debian_10_latest"
	defaultCloudName     = "cloud"
	defaultSSHUser       = "linux"
	defaultSSHPort       = 22
	defaultRegion        = "eu-de"
	defaultAuthURL       = "https://iam.eu-de.otc.t-systems.com/v3"
)

type Driver struct {
	*drivers.BaseDriver
	Cloud            string
	AuthUrl          string
	CaCert           string
	ValidateCert     bool
	DomainID         string
	DomainName       string
	Username         string
	Password         string
	TenantName       string
	TenantId         string
	ProjectName      string
	ProjectId        string
	Region           string
	AvailabilityZone string
	EndpointType     string
	MachineId        string
	FlavorName       string
	FlavorId         string
	ImageName        string
	ImageId          string
	KeyPairName      string
	VPCName          string
	VPCId            string
	SubnetName       string
	SubnetId         string
	PrivateKeyFile   string
	SecurityGroup    string
	FloatingIpPool   string
	FloatingIpPoolId string
	Token            string
	IpVersion        int
	ConfigDrive      bool
	client           *Client
}

func (d *Driver) Create() error {
	panic("implement me")
}

func (d *Driver) checkArgs() error {
	panic("implement me")
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "otc-cloud",
			EnvVar: "OS_CLOUD",
			Usage:  "Name of cloud in `clouds.yaml` file",
			Value:  "",
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
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-id",
			EnvVar: "OS_DOMAIN_ID",
			Usage:  "OpenTelekomCloud domain ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-domain-name",
			EnvVar: "OS_DOMAIN_NAME",
			Usage:  "OpenTelekomCloud domain name",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-username",
			EnvVar: "OS_USERNAME",
			Usage:  "OpenTelekomCloud username",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-password",
			EnvVar: "OS_PASSWORD",
			Usage:  "OpenTelekomCloud password",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-tenant-name",
			EnvVar: "OS_TENANT_NAME",
			Usage:  "OpenTelekomCloud project name",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-tenant-id",
			EnvVar: "OS_TENANT_ID",
			Usage:  "OpenTelekomCloud project ID",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:   "otc-region",
			EnvVar: "OS_REGION_NAME",
			Usage:  "OpenTelekomCloud region name",
			Value:  defaultRegion,
		},
		mcnflag.StringFlag{
			Name:   "otc-availability-zone",
			EnvVar: "OS_AVAILABILITY_ZONE",
			Usage:  "OpenTelekomCloud availability zone",
			Value:  defaultAZ,
		},
		mcnflag.StringFlag{
			Name:   "otc-flavor-id",
			EnvVar: "OS_FLAVOR_ID",
			Usage:  "OpenTelekomCloud flavor id to use for the instance",
			Value:  "",
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
			Value:  "",
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
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:  "otc-vpc-id",
			Usage: "OpenTelekomCloud VPC id the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "otc-vpc-name",
			Usage: "OpenTelekomCloud VPC name the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "otc-subnet-id",
			Usage: "OpenTelekomCloud subnet id the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "otc-subnet-name",
			Usage: "OpenTelekomCloud subnet name the machine will be connected on",
			Value: "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_PRIVATE_KEY_FILE",
			Name:   "otc-private-key-file",
			Usage:  "Private keyfile to use for SSH (absolute path)",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_USER_DATA_FILE",
			Name:   "otc-user-data-file",
			Usage:  "File containing an otc userdata script",
		},
		mcnflag.StringFlag{
			Name:   "otc-token",
			EnvVar: "OS_TOKEN",
			Usage:  "OpenTelekomCloud authorization token",
		},
		mcnflag.StringFlag{
			EnvVar: "OS_SECURITY_GROUPS",
			Name:   "otc-sec-groups",
			Usage:  "OpenTelekomCloud comma separated security groups for the machine",
			Value:  defaultSecurityGroup,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_FLOATINGIP_POOL",
			Name:   "otc-floatingip-pool",
			Usage:  "OpenTelekomCloud floating IP pool to get an IP from to assign to the instance",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "OS_IP_VERSION",
			Name:   "otc-ip-version",
			Usage:  "OpenTelekomCloud version of IP address assigned for the machine",
			Value:  4,
		},
		mcnflag.StringFlag{
			EnvVar: "OS_SSH_USER",
			Name:   "otc-ssh-user",
			Usage:  "otc SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "OS_SSH_PORT",
			Name:   "otc-ssh-port",
			Usage:  "otc SSH port",
			Value:  defaultSSHPort,
		},
	}
}

func (d *Driver) GetMachineName() string {
	panic("implement me")
}

const dockerPort = "2376"

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

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil || ip == "" {
		return "", err
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, dockerPort)), nil
}

func (d *Driver) GetState() (state.State, error) {
	panic("implement me")
}

func (d *Driver) Kill() error {
	panic("implement me")
}

func (d *Driver) PreCreateCheck() error {
	panic("implement me")
}

func (d *Driver) Remove() error {
	panic("implement me")
}

func (d *Driver) Restart() error {
	panic("implement me")
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	panic("implement me")
}

func (d *Driver) Start() error {
	panic("implement me")
}

func (d *Driver) Stop() error {
	panic("implement me")
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			SSHUser:     defaultSSHUser,
			SSHPort:     defaultSSHPort,
			StorePath:   storePath,
		},
		client: &Client{},
	}
}
