package opentelekomcloud

import (
	"os"
	"strings"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver/services"
)

// GetCreateFlags - DMD flags
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-cloud",
			EnvVar: "OS_CLOUD",
			Usage:  "Name of cloud in `clouds.yaml` file",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-auth-url",
			EnvVar: "OS_AUTH_URL",
			Usage:  "OpenTelekomCloud authentication URL",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-cacert",
			EnvVar: "OS_CACERT",
			Usage:  "CA certificate bundle to verify against",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-domain-id",
			EnvVar: "OS_DOMAIN_ID",
			Usage:  "OpenTelekomCloud domain ID",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-domain-name",
			EnvVar: "OS_DOMAIN_NAME",
			Usage:  "OpenTelekomCloud domain name",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-username",
			EnvVar: "OS_USERNAME",
			Usage:  "OpenTelekomCloud username",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-password",
			EnvVar: "OS_PASSWORD",
			Usage:  "OpenTelekomCloud password",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-project-name",
			EnvVar: "OS_PROJECT_NAME",
			Usage:  "OpenTelekomCloud project name",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-project-id",
			EnvVar: "OS_PROJECT_ID",
			Usage:  "OpenTelekomCloud project ID",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-region",
			EnvVar: "OS_REGION",
			Usage:  "OpenTelekomCloud region name",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-access-key",
			Usage:  "OpenTelekomCloud access key ID for AK/SK auth",
			EnvVar: "OS_ACCESS_KEY",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-secret-key",
			Usage:  "OpenTelekomCloud secret access key for AK/SK auth",
			EnvVar: "OS_SECRET_KEY",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-availability-zone",
			EnvVar: "OS_AVAILABILITY_ZONE",
			Usage:  "OpenTelekomCloud availability zone",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-flavor-id",
			EnvVar: "OS_FLAVOR_ID",
			Usage:  "OpenTelekomCloud flavor id to use for the instance",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-flavor-name",
			EnvVar: "OS_FLAVOR_NAME",
			Usage:  "OpenTelekomCloud flavor name to use for the instance",
			Value:  defaultFlavor,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-image-id",
			EnvVar: "OS_IMAGE_ID",
			Usage:  "OpenTelekomCloud image id to use for the instance",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-image-name",
			EnvVar: "OS_IMAGE_NAME",
			Usage:  "OpenTelekomCloud image name to use for the instance",
			Value:  defaultImage,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-keypair-name",
			EnvVar: "OS_KEYPAIR_NAME",
			Usage:  "OpenTelekomCloud keypair to use to SSH to the instance",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-vpc-id",
			EnvVar: "OS_VPC_ID",
			Usage:  "OpenTelekomCloud VPC id the machine will be connected on",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-vpc-name",
			EnvVar: "OS_VPC_NAME",
			Usage:  "OpenTelekomCloud VPC name the machine will be connected on",
			Value:  defaultVpcName,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-subnet-id",
			EnvVar: "OS_NETWORK_ID",
			Usage:  "OpenTelekomCloud subnet id the machine will be connected on",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-subnet-name",
			EnvVar: "OS_NETWORK_NAME",
			Usage:  "OpenTelekomCloud subnet name the machine will be connected on",
			Value:  defaultSubnetName,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-private-key-file",
			EnvVar: "OS_PRIVATE_KEY_FILE",
			Usage:  "Private key file to use for SSH (absolute path)",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-user-data-file",
			EnvVar: "OS_USER_DATA_FILE",
			Usage:  "File containing an user data script",
		},
		mcnflag.StringFlag{
			Name:  "opentelekomcloud-user-data-raw",
			Usage: "Contents of user data file as a string",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-token",
			EnvVar: "OS_TOKEN",
			Usage:  "OpenTelekomCloud authorization token",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-sec-groups",
			EnvVar: "OS_SECURITY_GROUP",
			Usage:  "Existing security groups to use, separated by comma",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-eip",
			EnvVar: "OS_EIP",
			Usage:  "OpenTelekomCloud floating IP to use",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-eip-type",
			EnvVar: "OS_EIP_TYPE",
			Usage:  "OpenTelekomCloud bandwidth type",
			Value:  "5_bgp",
		},
		mcnflag.IntFlag{
			Name:   "opentelekomcloud-bandwidth-size",
			EnvVar: "OS_BANDWIDTH_SIZE",
			Usage:  "OpenTelekomCloud bandwidth size",
			Value:  100,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-bandwidth-type",
			EnvVar: "OS_BANDWIDTH_TYPE",
			Usage:  "OpenTelekomCloud bandwidth share type",
			Value:  "PER",
		},
		mcnflag.BoolFlag{
			Name:  "opentelekomcloud-skip-eip",
			Usage: "If set, elastic IP won't be created",
		},
		mcnflag.IntFlag{
			Name:   "opentelekomcloud-ip-version",
			EnvVar: "OS_IP_VERSION",
			Usage:  "OpenTelekomCloud version of IP address assigned for the machine",
			Value:  4,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-ssh-user",
			EnvVar: "OS_SSH_USER",
			Usage:  "Machine SSH username",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			Name:   "opentelekomcloud-ssh-port",
			EnvVar: "OS_SSH_PORT",
			Usage:  "Machine SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-endpoint-type",
			EnvVar: "OS_INTERFACE",
			Usage:  "OpenTelekomCloud interface (endpoint) type",
			Value:  "public",
		},
		mcnflag.BoolFlag{
			Name:  "opentelekomcloud-skip-default-sg",
			Usage: "Don't create default security group",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-server-group",
			EnvVar: "OS_SERVER_GROUP",
			Usage:  "Define server group where server will be created",
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-server-group-id",
			EnvVar: "OS_SERVER_GROUP_ID",
			Usage:  "Define server group where server will be created by ID",
		},
		mcnflag.IntFlag{
			Name:   "opentelekomcloud-root-volume-size",
			EnvVar: "OS_ROOT_VOLUME_SIZE",
			Usage:  "Set volume size of root partition",
			Value:  defaultVolumeSize,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-root-volume-type",
			EnvVar: "OS_ROOT_VOLUME_TYPE",
			Usage:  "Set volume type of root partition (one of SATA, SAS, SSD)",
			Value:  defaultVolumeType,
		},
		mcnflag.StringFlag{
			Name:   "opentelekomcloud-tags",
			EnvVar: "OS_TAGS",
			Usage:  "Comma-separated list of instance tags (e.g. key1.value1,key2.value2,key3)",
		},
	}
}

// SetConfigFromFlags loads driver configuration from given flags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AuthURL = flags.String("opentelekomcloud-auth-url")
	d.Cloud = flags.String("opentelekomcloud-cloud")
	d.CACert = flags.String("opentelekomcloud-cacert")
	d.DomainID = flags.String("opentelekomcloud-domain-id")
	d.DomainName = flags.String("opentelekomcloud-domain-name")
	d.Username = flags.String("opentelekomcloud-username")
	d.Password = flags.String("opentelekomcloud-password")
	d.ProjectName = flags.String("opentelekomcloud-project-name")
	d.ProjectID = flags.String("opentelekomcloud-project-id")
	d.Region = flags.String("opentelekomcloud-region")
	d.Token = flags.String("opentelekomcloud-token")
	d.AccessKey = flags.String("opentelekomcloud-access-key")
	d.SecretKey = flags.String("opentelekomcloud-secret-key")

	d.AvailabilityZone = flags.String("opentelekomcloud-availability-zone")
	d.EndpointType = flags.String("opentelekomcloud-endpoint-type")
	d.FlavorID = flags.String("opentelekomcloud-flavor-id")
	d.FlavorName = flags.String("opentelekomcloud-flavor-name")
	d.ImageName = flags.String("opentelekomcloud-image-name")
	d.VpcID = managedSting{Value: flags.String("opentelekomcloud-vpc-id")}
	d.VpcName = flags.String("opentelekomcloud-vpc-name")
	d.SubnetID = managedSting{Value: flags.String("opentelekomcloud-subnet-id")}
	d.SubnetName = flags.String("opentelekomcloud-subnet-name")
	d.ElasticIP = managedSting{Value: flags.String("opentelekomcloud-eip")}
	d.IPVersion = flags.Int("opentelekomcloud-ip-version")
	d.SSHUser = flags.String("opentelekomcloud-ssh-user")
	d.SSHPort = flags.Int("opentelekomcloud-ssh-port")
	d.KeyPairName = managedSting{Value: flags.String("opentelekomcloud-keypair-name")}
	d.PrivateKeyFile = flags.String("opentelekomcloud-private-key-file")

	userDataFile := flags.String("opentelekomcloud-user-data-file")
	if userDataFile != "" {
		log.Debugf("opentelekomcloud: user-data-file flag is set: %s", userDataFile)
		userData, err := os.ReadFile(userDataFile)
		if err == nil {
			d.UserData = userData
			log.Debugf("opentelekomcloud: loaded user-data from file, size=%d bytes", len(userData))
		} else {
			log.Errorf("opentelekomcloud: failed to read user-data-file %s: %v", userDataFile, err)
			return err
		}
	} else {
		log.Debug("opentelekomcloud: opentelekomcloud-user-data-file flag is empty")
	}

	rawUserData := flags.String("opentelekomcloud-user-data-raw")
	if rawUserData != "" {
		log.Debugf("opentelekomcloud: user-data-raw flag is set, size=%d bytes", len(rawUserData))
		d.UserData = []byte(rawUserData)
	} else if d.UserData == nil {
		log.Debug("opentelekomcloud: no user-data provided via flags")
	}

	d.ServerGroup = flags.String("opentelekomcloud-server-group")
	d.ServerGroupID = flags.String("opentelekomcloud-server-group-id")
	tags := flags.String("opentelekomcloud-tags")
	if tags != "" {
		d.Tags = strings.Split(tags, ",")
	}

	d.RootVolumeOpts = &services.DiskOpts{
		SourceID: flags.String("opentelekomcloud-image-id"),
		Size:     flags.Int("opentelekomcloud-root-volume-size"),
		Type:     flags.String("opentelekomcloud-root-volume-type"),
	}

	d.eipConfig = &services.ElasticIPOpts{
		IPType:        flags.String("opentelekomcloud-eip-type"),
		BandwidthSize: flags.Int("opentelekomcloud-bandwidth-size"),
		BandwidthType: flags.String("opentelekomcloud-bandwidth-type"),
	}
	d.skipEIPCreation = flags.Bool("opentelekomcloud-skip-eip")

	if sg := flags.String("opentelekomcloud-sec-groups"); sg != "" {
		d.SecurityGroups = strings.Split(sg, ",")
	}

	if !flags.Bool("opentelekomcloud-skip-default-sg") {
		d.ManagedSecurityGroup = defaultSecurityGroup
	}

	// Fill region-derived defaults (AuthURL, AvailabilityZone) for any values
	// the user did not set explicitly. Keeps backwards-compat for eu-de and
	// makes eu-ch2 (Swiss OTC, `iam-pub.` prefix) work with just --region.
	d.applyRegionDefaults()

	d.SetSwarmConfigFromFlags(flags)
	return d.checkConfig()
}
