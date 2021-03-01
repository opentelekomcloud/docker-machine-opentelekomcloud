## Usage With Rancher

Currently, `docker-machine-opentelekomcloud` is not a part of the Rancher, so additional steps required to use it as a node driver

### Remove old node driver:

 * Open Rancher UI page and go to `Tools` -> `Drivers` -> `Node Drivers`
 * Check current preinstalled `Open Telekom Cloud` driver and remove it, because it produces conflicts with current implementation.

### Usage of new node driver:

 * Open OpenTelekomDriver [releases](https://github.com/opentelekomcloud/docker-machine-opentelekomcloud/releases) page and copy link of binary for 64-bit Linux
 * Click `Add New Driver` button, insert copied link and click `Create`
 * Wait for a while. Driver should be downloaded and be in `Active` state.
 * Create new OTC driver template
