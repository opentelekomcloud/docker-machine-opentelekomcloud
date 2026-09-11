package main

import (
	opentelekomcloud "github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver"
	"github.com/rancher/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(opentelekomcloud.NewDriver("default", ""))
}
