# OpenTelekomCloud Docker Machine driver

[![Zuul Gate](https://zuul.eco.tsi-dev.otc-service.com/api/tenant/eco/badge?project=opentelekomcloud/docker-machine-opentelekomcloud&pipeline=check&branch=devel)](https://zuul.eco.tsi-dev.otc-service.com/t/eco/builds?project=opentelekomcloud%2Fdocker-machine-opentelekomcloud)
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/docker-machine-opentelekomcloud)](https://goreportcard.com/report/github.com/opentelekomcloud/docker-machine-opentelekomcloud)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/docker-machine-opentelekomcloud)

OpenTelekomCloud driver for docker-machine

### Comparing with other drivers

There are two more options of docker-machine driver suitable for usage with OpenTelekomCloud:

* [docker-machine-openstack](https://opendev.org/x/docker-machine-openstack) ― docker-machine built-in
* [DockerMachine4OTC](https://github.com/Huawei/DockerMachineDriver4OTC) ― older OTC driver implementation by Huawei

This driver is inspired by `docker-machine-openstack`.

In versions `v0.3.+` duplicating options were removed and all environment variables are prefixed with `OS_`.

Feature                                        | OTC (new)   | OTC (old) | Openstack
---                                            | ---         | ---       | ---
Automated creation of required infrastructure  | **Yes**     | No        | No
Support of `clouds.yaml` and `OS_CLOUD`        | **Yes**     | No        | No
Support using of resource names instead of IDs | Yes         | No        | Yes
User data injection                            | Yes         | No        | Yes
Elastic IP pool selection                      | No          | No        | Yes
Custom CA usage                                | Yes         | No        | Yes
Insecure mode (without TLS certificate check)  | No          | No        | Yes
Bandwidth configuration                        | Yes         | Yes       | No
Root volume configuration                      | Yes         | Yes       | No
Optional usage of elastic IP                   | Yes         | Yes       | No
AK/SK auth                                     | Yes         | Yes       | No
Server group                                   | **Yes**     | No        | No
Security group(s)                              | Multiple    | Single    | Multiple
Rancher integration                            | Needs setup | Built-in  | Built-in

### Installation

Driver can be installed several ways

#### From source code

_(Requires Go 1.13+, gcc and make installed)_

1. Clone [this](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud) git repository to any location
2. Run `make build && sudo make install`, driver for linux will be built and copied to `/usr/local/bin`

#### Using built binary

An already built driver for both Linux and Windows distributions can be found in
[releases section](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud/releases)

You will have to copy driver to directory in `$PATH` so `docker-machine` would be able to find it

### Usage

`docker-machine-opentelekomcloud` can be used either as Rancher node driver or as stand-alone Docker Machine driver

#### Stand-alone

`OpenTelekomCloud` driver processes existing `clouds.yaml` files to authenticate in OTC

Having `otc` cloud in your `clouds.yaml`, creating of docker-machine is as easy as running

```shell
$ docker-machine create -d otc --otc-cloud otc default
```

**Following will be created if not provided:**

- **Security Group:** `docker-machine-grp`
- **VPC:** `vpc-docker-machine`
- **Subnet:** `subnet-docker-machine`
- **Elastic IP:** with bandwidth limited to `100` MBit/s

**Machine with following setup will be started:**

- **Flavor:** `s2.large.2`
- **Image:** `Standard_Ubuntu_20.04_latest`
- **Volume Size:** `40` GB
- **Volume Type:** `SSD`

*Removing machine will remove all resources created on machine creation*

#### Supported options

For versions `v0.3.x` see [supported-options](docs/supported-options-v0.3.x.md)

For versions `v0.2.x` see [supported-options](docs/supported-options-v0.2.x.md).

Please **note** that only `v0.2.x` support old flags and targets to provide full backward compatibility
with `DockerMachineDriver4OTC`.

#### With rancher

See [usage with rancher](docs/usage-with-rancher.md)
