package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	opentelekomcloud "github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver"
)

func main() {
	plugin.RegisterDriver(opentelekomcloud.NewDriver("default", ""))
}
