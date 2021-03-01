package opentelekomcloud

import (
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/opentelekomcloud-infra/crutch-house/services"
)

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
			EnvVar: "OS_REGION",
			Usage:  "OpenTelekomCloud region name",
			Value:  defaultRegion,
		},
		mcnflag.StringFlag{
			Name:   "otc-access-key-id",
			Usage:  "OpenTelekomCloud access key ID for AK/SK auth",
			EnvVar: "OS_ACCESS_KEY_ID",
		},
		mcnflag.StringFlag{
			Name:   "otc-secret-access-key",
			Usage:  "OpenTelekomCloud secret access key for AK/SK auth",
			EnvVar: "OS_SECRET_ACCESS_KEY",
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
	d.ProjectID = flags.String("otc-project-id")
	d.Region = flags.String("otc-region")
	d.AvailabilityZone = flags.String("otc-availability-zone")
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
	d.SecretKey = flags.String("otc-secret-access-key")

	d.RootVolumeOpts = &services.DiskOpts{
		SourceID: flags.String("otc-image-id"),
		Size:     flags.Int("otc-root-volume-size"),
		Type:     flags.String("otc-root-volume-type"),
	}

	d.eipConfig = &services.ElasticIPOpts{
		IPType:        flags.String("otc-floating-ip-type"),
		BandwidthSize: flags.Int("otc-bandwidth-size"),
		BandwidthType: flags.String("otc-bandwidth-type"),
	}
	d.skipEIPCreation = flags.Bool("otc-skip-ip")

	if sg := flags.String("otc-sec-groups"); sg != "" {
		d.SecurityGroups = strings.Split(sg, ",")
	}

	if !flags.Bool("otc-skip-default-sg") {
		d.ManagedSecurityGroup = defaultSecurityGroup
	}

	d.SetSwarmConfigFromFlags(flags)
	return d.checkConfig()
}
