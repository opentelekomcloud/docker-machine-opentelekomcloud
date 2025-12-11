# OpenTelekomCloud Docker Machine driver

[![Zuul Gate](https://zuul.otc-service.com/api/tenant/eco/badge?project=opentelekomcloud/docker-machine-opentelekomcloud&pipeline=check&branch=devel)](https://zuul.eco.tsi-dev.otc-service.com/t/eco/builds?project=opentelekomcloud%2Fdocker-machine-opentelekomcloud)
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/docker-machine-opentelekomcloud)](https://goreportcard.com/report/github.com/opentelekomcloud/docker-machine-opentelekomcloud)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/docker-machine-opentelekomcloud)

OpenTelekomCloud driver for docker-machine

### Comparing with other drivers

There is another option of docker-machine driver suitable for usage with OpenTelekomCloud:

* [docker-machine-openstack](https://opendev.org/x/docker-machine-openstack) ― docker-machine built-in

| Feature                                        | OTC         | Openstack |
|------------------------------------------------|-------------|-----------|
| Automated creation of required infrastructure  | **Yes**     | No        |
| Support of `clouds.yaml` and `OS_CLOUD`        | **Yes**     | No        |
| Support using of resource names instead of IDs | Yes         | Yes       |
| User data injection                            | Yes         | Yes       |
| Elastic (floating) IP pool selection           | No          | Yes       |
| Custom CA usage                                | Yes         | Yes       |
| Insecure mode (without TLS certificate check)  | No          | Yes       |
| Bandwidth configuration                        | Yes         | No        |
| Root volume configuration                      | Yes         | No        |
| Optional usage of elastic IP                   | Yes         | No        |
| AK/SK auth                                     | Yes         | No        |
| Server group                                   | **Yes**     | No        |
| Security group(s)                              | Multiple    | Multiple  |
| Instance Tags                                  | Multiple    | No        |
| Rancher integration                            | Needs setup | Built-in  |

### Installation

Driver can be installed several ways

#### From source code

_(Requires Go 1.13+, gcc and make installed)_

1. Clone [this](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud) git repository to any location
2. Run `make build && sudo make install`, driver for linux will be built and copied to `/usr/local/bin`

#### Using built binary

An already built driver for both Linux and Windows distributions can be found in
[releases section](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud/releases).

You will have to copy driver to directory in `$PATH` so `docker-machine` would be able to find it.

### Usage

`docker-machine-opentelekomcloud` can be used either as Rancher node driver or as stand-alone Docker Machine driver.

#### Stand-alone

`OpenTelekomCloud` driver processes existing `clouds.yaml` files to authenticate in OTC.

Having `opentelekomcloud` cloud in your `clouds.yaml`, creating of docker-machine is as easy as running.

```shell
$ docker-machine create -d opentelekomcloud --opentelekomcloud-cloud opentelekomcloud default
```

**Following will be created if not provided:**

- **Security Group:** `docker-machine-grp`
- **VPC:** `vpc-docker-machine`
- **Subnet:** `subnet-docker-machine`
- **Elastic IP:** with bandwidth limited to `100` MBit/s

**Machine with following setup will be started:**

- **Flavor:** `s3.xlarge.2`
- **Image:** `Standard_Ubuntu_24.04_amd64_uefi_latest`
- **Volume Size:** `40` GB
- **Volume Type:** `SSD`

*Removing machine will remove all resources created on machine creation*.

#### Supported options

For versions `v0.3.x` see [supported-options](docs/supported-options-v0.3.x.md).

For versions `v0.2.x` see [supported-options](docs/supported-options-v0.2.x.md).

For versions `v2.0.x` see [supported-options](docs/supported-options-v2.0.x.md).

Please **note** that only `v0.2.x` support old flags and targets to provide full backward compatibility
with `DockerMachineDriver4OTC`. In versions `v0.3.+` duplicating options were removed and all environment variables are
prefixed with `OS_`.

Please **note** that only `v2.0.x` support RKE2 cluster installation.
All variables CLI variables now prefixed with `opentelekomcloud`.

#### With Rancher

See [Rancher integration](docs/usage-with-rancher.md).
