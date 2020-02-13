package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/opentelekomcloud/docker-machine-opentelekomcloud/driver"
)

func main() {
	plugin.RegisterDriver(opentelekomcloud.NewDriver("default", ""))
}
