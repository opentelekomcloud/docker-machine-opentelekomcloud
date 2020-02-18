# OpenTelekomCloud Docker Machine driver
[![Build Status](https://travis-ci.org/opentelekomcloud/docker-machine-opentelekomcloud.svg)](https://travis-ci.org/opentelekomcloud/docker-machine-opentelekomcloud)
[![Go Report Card](https://goreportcard.com/badge/github.com/opentelekomcloud/docker-machine-opentelekomcloud)](https://goreportcard.com/report/github.com/opentelekomcloud/docker-machine-opentelekomcloud)
[![codecov](https://codecov.io/gh/opentelekomcloud/docker-machine-opentelekomcloud/branch/master/graph/badge.svg)](https://codecov.io/gh/opentelekomcloud/docker-machine-opentelekomcloud)
![GitHub](https://img.shields.io/github/license/opentelekomcloud/docker-machine-opentelekomcloud)

OpenTelekomCloud driver for docker-machine

---
NB! Driver is currently in active development phase
---

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
docker-machine create -d opentelekomcloud --otc-cloud otc default
```

**Following will be created:**

- **Security Group**: `docker-machine-grp`
- **VPC** `vpc-docker-machine`
- **Subnet** `subnet-docker-machine`
- **Floating IP** with bandwidth limited to 1000

**Machine with following setup will be started:**
- **Flavor** `s2.large.2`
- **Image** `Standard_Debian_10_latest`

*Removing machine will remove all resources created on machine creation*
