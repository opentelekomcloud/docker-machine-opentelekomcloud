#### Supported options for v0.3.x

Flag | Env variable | Default value | Description
--- | --- | --- | ---
`--otc-access-key`        | `OS_ACCESS_KEY`        |                                     | Access key for AK/SK auth
`--otc-secret-key`        | `OS_SECRET_KEY`        |                                     | Secret key for AK/SK auth
`--otc-auth-url`          | `OS_AUTH_URL`          | https://iam.eu-de.otc.t-systems.com | Authentication URL
`--otc-availability-zone` | `OS_AVAILABILITY_ZONE` | eu-de-03                            | Availability zone
`--otc-cloud`             | `OS_CLOUD`             |                                     | Name of cloud in `clouds.yaml` file
`--otc-cacert`            | `OS_CACERT`            |                                     | CA certificate bundle to verify against
`--otc-domain-id`         | `OS_DOMAIN_ID`         |                                     | OpenTelekomCloud Domain ID
`--otc-domain-name`       | `OS_DOMAIN_NAME`       |                                     | OpenTelekomCloud Domain name
`--otc-eip`               | `OS_EIP`               |                                     | Elastic IP to use
`--otc-eip-type`          | `OS_EIP_TYPE`          | 5_bgp                               | Bandwidth type (either `5_bgp` or `5_mailbgp`)
`--otc-endpoint-type`     | `OS_INTERFACE`         | public                              | Endpoint type
`--otc-flavor-id`         | `OS_FLAVOR_ID`         |                                     | Flavor id to use for the instance
`--otc-flavor-name`       | `OS_FLAVOR_NAME`       | s2.large.2                          | Flavor name to use for the instance
`--otc-bandwidth-size`    | `OS_BANDWIDTH_SIZE`    | 100 (MBit/s)                        | Bandwidth size
`--otc-bandwidth-type`    | `OS_BANDWIDTH_TYPE`    | PER (exclusive bandwidth)           | Bandwidth share type
`--otc-image-id`          | `OS_IMAGE_ID`          |                                     | Image ID to use for the instance
`--otc-image-name`        | `OS_IMAGE_NAME`        | Standard_Ubuntu_20.04_latest        | Image name to use for the instance
`--otc-ip-version`        | `OS_IP_VERSION`        | 4                                   | Version of IP address assigned for the machine (only 4 is supported by OTC for now)
`--otc-keypair-name`      | `OS_KEYPAIR_NAME`      |                                     | Key pair to use to SSH to the instance
`--otc-password`          | `OS_PASSWORD`          |                                     | OpenTelekomCloud Password
`--otc-private-key-file`  | `OS_PRIVATE_KEY_FILE`  |                                     | Private key file to use for SSH (absolute path)
`--otc-project-id`        | `OS_PROJECT_ID`        |                                     | OpenTelekomCloud Project ID
`--otc-project-name`      | `OS_PROJECT_NAME`      |                                     | OpenTelekomCloud Project name
`--otc-region`            | `OS_REGION`            | eu-de                               | Region name
`--otc-root-volume-size`  | `OS_ROOT_VOLUME_SIZE`  | 40                                  | Set volume size of root partition (in GB)
`--otc-root-volume-type`  | `OS_ROOT_VOLUME_TYPE`  | SSD                                 | Set volume type of root partition (one of `SATA`, `SAS`, `SSD`)
`--otc-sec-groups`        | `OS_SECURITY_GROUP`    |                                     | Existing security groups to use, separated by comma
`--otc-server-group`      | `OS_SERVER_GROUP`      |                                     | Define server group where server will be created
`--otc-server-group-id`   | `OS_SERVER_GROUP_ID`   |                                     | Define server group where server will be created by ID
`--otc-skip-default-sg`   |                        |                                     | Don't create default security group
`--otc-skip-eip`          |                        |                                     | If set, elastic IP won't be created, machine IP will be set to instance local IP
`--otc-ssh-port`          | `OS_SSH_PORT`          | 22                                  | Machine SSH port
`--otc-ssh-user`          | `OS_SSH_USER`          | ubuntu                              | SSH user
`--otc-subnet-id`         | `OS_SUBNET_ID`         |                                     | Subnet ID the machine will be connected on
`--otc-subnet-name`       | `OS_SUBNET_NAME`       | subnet-docker-machine               | Subnet name the machine will be connected on
`--otc-token`             | `OS_TOKEN`             |                                     | Authorization token
`--otc-tags`              | `OS_TAGS`              |                                     | Comma-separated list of instance tags
`--otc-user-data-file`    | `OS_USER_DATA_FILE`    |                                     | File containing an userdata script
`--otc-user-data-raw`     |                        |                                     | Contents of user data file as a string
`--otc-username`          | `OS_USERNAME`          |                                     | OpenTelekomCloud username
`--otc-vpc-id`            | `OS_VPC_ID`            |                                     | VPC ID the machine will be connected on
`--otc-vpc-name`          | `OS_VPC_NAME`          | vpc-docker-machine                  | VPC name the machine will be connected on
