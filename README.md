# OpenTelekomCloud Docker Machine driver
[![Build Status](https://travis-ci.org/opentelekomcloud/docker-machine-opentelekomcloud.svg)](https://travis-ci.org/opentelekomcloud/docker-machine-opentelekomcloud)
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/docker-machine-opentelekomcloud)](https://goreportcard.com/report/github.com/opentelekomcloud/docker-machine-opentelekomcloud)
[![codecov](https://codecov.io/gh/opentelekomcloud/docker-machine-opentelekomcloud/branch/devel/graph/badge.svg)](https://codecov.io/gh/opentelekomcloud/docker-machine-opentelekomcloud/branch/devel)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/docker-machine-opentelekomcloud)

OpenTelekomCloud driver for docker-machine

### Installation

Driver can be installed several ways

#### From source code
_(Requires Go 1.11+, gcc and make installed)_
1. Clone [this](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud) git repository to any location
2. Run `make build && sudo make install`, driver for linux will be built and copied to `/usr/local/bin`

#### Using built binary
Already built driver for both Linux and Windows distributions can be found in
[releases section](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud/releases)

You will have to copy driver to directory in `PATH` so `docker-machine` would be able to find it

### Usage

`OpenTelekomCloud` driver processes existing `clouds.yaml` files to authenticate in OTC

Having `otc` cloud in your `clouds.yaml`, creating of docker-machine is as easy as running

```bash
docker-machine create -d otc --otc-cloud otc default
```

**Following will be created if not provided:**

- **Security Group**: `docker-machine-grp`
- **VPC** `vpc-docker-machine`
- **Subnet** `subnet-docker-machine`
- **Floating IP** with bandwidth limited to 1000

**Machine with following setup will be started:**
- **Flavor** `s2.large.2`
- **Image** `Standard_Debian_10_latest`

*Removing machine will remove all resources created on machine creation*

#### Supported options
Flag | Env variable | Default value | Description
--- | --- | --- | ---
`--otc-access-key-id`                           | `ACCESS_KEY_ID`          |                                        | Access key ID for AK/SK auth
`--otc-access-key-key`                          | `ACCESS_KEY_SECRET`      |                                        | Secret access key for AK/SK auth
`--otc-auth-url`                                | `OS_AUTH_URL`            | https://iam.eu-de.otc.t-systems.com    | Authentication URL
`--otc-availability-zone`                       | `OS_AVAILABILITY_ZONE`   | eu-de-03                               | Availability zone
`--otc-available-zone`                          |                          |                                        | Availability zone. **DEPRECATED**: use `-otc-availability-zone` instead
`--otc-bandwidth-size`                          | `BANDWIDTH_SIZE`         | 100 (MBit/s)                           | Bandwidth size
`--otc-bandwidth-type`                          | `BANDWIDTH_TYPE`         | PER (exclusive bandwidth)              | Bandwidth share type
`--otc-cloud`                                   | `OS_CLOUD`               |                                        | Name of cloud in `clouds.yaml` file
`--otc-domain-id`                               | `OS_DOMAIN_ID`           |                                        | OpenTelekomCloud Domain ID
`--otc-domain-name`                             | `OS_DOMAIN_NAME`         |                                        | OpenTelekomCloud Domain name
`--otc-elastic-ip`                              | `ELASTIC_IP`             | 1                                      | If set to 0, elastic IP won't be created. **DEPRECATED**: use `-otc-skip-ip` instead
`--otc-elastic-ip-type`                         | `ELASTICIP_TYPE`         |                                        | Bandwidth type. **DEPRECATED!** Use `-otc-floating-ip-type` instead
`--otc-endpoint-type`                           |                          | public                                 | Endpoint type
`--otc-flavor-id`                               | `OS_FLAVOR_ID`           |                                        | Flavor id to use for the instance
`--otc-flavor-name`                             | `OS_FLAVOR_NAME`         | s2.large.2                             | Flavor name to use for the instance
`--otc-floating-ip`                             | `OS_FLOATINGIP`          |                                        | Floating IP to use
`--otc-floating-ip-type`                        |                          | 5_bgp                                  | Bandwidth type (either `5_bgp` or `5_mailbgp`)
`--otc-image-id`                                | `OS_IMAGE_ID`            |                                        | Image id to use for the instance
`--otc-image-name`                              | `OS_IMAGE_NAME`          | Standard_Debian_10_latest              | Image name to use for the instance
`--otc-ip-version    `                          | `OS_IP_VERSION`          | 4                                      | Version of IP address assigned for the machine
`--otc-k8s-group`                               |                          |                                        | Create security group with k8s ports allowed
`--otc-keypair-name`                            | `OS_KEYPAIR_NAME`        |                                        | Key pair to use to SSH to the instance
`--otc-password`                                | `OS_PASSWORD`            |                                        | OpenTelekomCloud Password
`--otc-private-key-file`                        | `OS_PRIVATE_KEY_FILE`    |                                        | Private key file to use for SSH (absolute path)
`--otc-project-id`                              | `OS_TENANT_ID`           |                                        | OpenTelekomCloud Project ID
`--otc-project-name`                            | `OS_TENANT_NAME`         |                                        | OpenTelekomCloud Project name
`--otc-region`                                  | `OS_REGION_NAME`         | eu-de                                  | Region name
`--otc-root-volume-size`                        |                          | 40                                     | Set volume size of root partition (in GB)
`--otc-root-volume-type`                        |                          | SATA                                   | Set volume type of root partition (one of SATA, SAS, SSD)
`--otc-sec-groups`                              | `OS_SECURITY_GROUP`      |                                        | Existing security groups to use, separated by comma
`--otc-skip-default-sg`                         |                          |                                        | Don't create default security group
`--otc-skip-ip`                                 |                          |                                        | If set, elastic IP won't be created
`--otc-ssh-port`                                | `OS_SSH_PORT`            | 22                                     | Machine SSH port
`--otc-ssh-user`                                | `OS_SSH_USER`            | linux                                  | SSH user
`--otc-subnet-id`                               |                          |                                        | Subnet id the machine will be connected on
`--otc-subnet-name`                             |                          | subnet-docker-machine                  | Subnet name the machine will be connected on
`--otc-token`                                   | `OS_TOKEN`               |                                        | Authorization token
`--otc-user-data-file`                          | `OS_USER_DATA_FILE`      |                                        | File containing an userdata script
`--otc-username`                                | `OS_USERNAME`            |                                        | OpenTelekomCloud username
`--otc-vpc-id`                                  |                          |                                        | VPC id the machine will be connected on
`--otc-vpc-name`                                |                          | vpc-docker-machine                     | VPC name the machine will be connected on
