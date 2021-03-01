#### Supported options for v0.2.x

Flag | Env variable | Default value | Description
--- | --- | --- | ---
`--otc-access-key-id`     | `ACCESS_KEY_ID`        |                                     | Access key ID for AK/SK auth
`--otc-access-key-key`    | `ACCESS_KEY_SECRET`    |                                     | Secret access key for AK/SK auth
`--otc-auth-url`          | `OS_AUTH_URL`          | https://iam.eu-de.otc.t-systems.com | Authentication URL
`--otc-availability-zone` | `OS_AVAILABILITY_ZONE` | eu-de-03                            | Availability zone
`--otc-available-zone`    | `AVAILABLE_ZONE`       |                                     | Availability zone. **DEPRECATED**: use `-otc-availability-zone` instead
`--otc-bandwidth-size`    | `BANDWIDTH_SIZE`       | 100 (MBit/s)                        | Bandwidth size
`--otc-bandwidth-type`    | `BANDWIDTH_TYPE`       | PER (exclusive bandwidth)           | Bandwidth share type
`--otc-cloud`             | `OS_CLOUD`             |                                     | Name of cloud in `clouds.yaml` file
`--otc-domain-id`         | `OS_DOMAIN_ID`         |                                     | OpenTelekomCloud Domain ID
`--otc-domain-name`       | `OS_DOMAIN_NAME`       |                                     | OpenTelekomCloud Domain name
`--otc-elastic-ip`        | `ELASTIC_IP`           | 1                                   | If set to 0, elastic IP won't be created. **DEPRECATED**: use `-otc-skip-ip` instead
`--otc-elastic-ip-type`   | `ELASTICIP_TYPE`       |                                     | Bandwidth type. **DEPRECATED!** Use `-otc-floating-ip-type` instead
`--otc-endpoint-type`     | `OS_INTERFACE`         | public                              | Endpoint type
`--otc-flavor-id`         | `FLAVOR_ID`            |                                     | Flavor id to use for the instance
`--otc-flavor-name`       | `OS_FLAVOR_NAME`       | s2.large.2                          | Flavor name to use for the instance
`--otc-floating-ip`       | `OS_FLOATING_IP`       |                                     | Floating IP to use
`--otc-floating-ip-type`  | `OS_FLOATING_IP_TYPE`  | 5_bgp                               | Bandwidth type (either `5_bgp` or `5_mailbgp`)
`--otc-image-id`          | `IMAGE_ID`             |                                     | Image id to use for the instance
`--otc-image-name`        | `OS_IMAGE_NAME`        | Standard_Ubuntu_18.04_latest        | Image name to use for the instance
`--otc-ip-version`        | `OS_IP_VERSION`        | 4                                   | Version of IP address assigned for the machine (only 4 is supported by OTC for now)
`--otc-k8s-group`         |                        |                                     | Create security group with k8s ports allowed
`--otc-keypair-name`      | `OS_KEYPAIR_NAME`      |                                     | Key pair to use to SSH to the instance
`--otc-password`          | `OS_PASSWORD`          |                                     | OpenTelekomCloud Password
`--otc-private-key-file`  | `OS_PRIVATE_KEY_FILE`  |                                     | Private key file to use for SSH (absolute path)
`--otc-project-id`        | `OS_PROJECT_ID`        |                                     | OpenTelekomCloud Project ID
`--otc-project-name`      | `OS_PROJECT_NAME`      |                                     | OpenTelekomCloud Project name
`--otc-region`            | `REGION`               | eu-de                               | Region name
`--otc-root-volume-size`  | `ROOT_VOLUME_SIZE`     | 40                                  | Set volume size of root partition (in GB)
`--otc-root-volume-type`  | `ROOT_VOLUME_TYPE`     | SATA                                | Set volume type of root partition (one of `SATA`, `SAS`, `SSD`)
`--otc-sec-groups`        | `OS_SECURITY_GROUP`    |                                     | Existing security groups to use, separated by comma
`--otc-skip-default-sg`   |                        |                                     | Don't create default security group
`--otc-skip-ip`           |                        |                                     | If set, elastic IP won't be created, machine IP will be set to instance local IP
`--otc-ssh-port`          | `OS_SSH_PORT`          | 22                                  | Machine SSH port
`--otc-ssh-user`          | `SSH_USER`             | ubuntu                              | SSH user
`--otc-subnet-id`         | `SUBNET_ID`            |                                     | Subnet id the machine will be connected on
`--otc-subnet-name`       | `SUBNET_NAME`          | subnet-docker-machine               | Subnet name the machine will be connected on
`--otc-token`             | `OS_TOKEN`             |                                     | Authorization token
`--otc-tenant-id`         | `TENANT_ID`            |                                     | Project ID. DEPRECATED: use `-otc-project-id` instead
`--otc-user-data-file`    | `OS_USER_DATA_FILE`    |                                     | File containing an userdata script
`--otc-username`          | `OS_USERNAME`          |                                     | OpenTelekomCloud username
`--otc-vpc-id`            | `VPC_ID`               |                                     | VPC id the machine will be connected on
`--otc-vpc-name`          | `OS_VPC_NAME`          | vpc-docker-machine                  | VPC name the machine will be connected on