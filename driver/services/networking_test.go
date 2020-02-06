package services

import (
	"github.com/huaweicloud/golangsdk"
	"testing"
)

func TestClient_CreateVPC(t *testing.T) {
	client, err := authClient()
	if err != nil {
		t.Error(authFailedMessage)
		return
	}
	if err := client.InitNetwork(); err != nil {
		t.Error(err)
		return
	}
	vpc, err := client.CreateVPC(vpcName)
	if err != nil {
		t.Error(err)
		return
	}
	if err := client.DeleteVPC(vpc.ID); err != nil {
		t.Error(err)
		return
	}
}

func TestClient_CreateSubnet(t *testing.T) {
	client, err := authClient()
	if err != nil {
		t.Error(authFailedMessage)
		return
	}
	if err := client.InitNetwork(); err != nil {
		t.Error(err)
		return
	}
	vpc, err := client.CreateVPC(vpcName)
	if err != nil {
		t.Error(err)
		return
	}

	subnet, err := client.CreateSubnet(vpc.ID, subnetName)
	if err != nil {
		t.Error(err)
	}

	err = client.WaitForSubnetStatus(subnet.ID, "ACTIVE")

	found, err := client.FindSubnet(vpc.ID, subnetName)
	if err != nil {
		t.Error(err)
	}
	if found != subnet.ID {
		t.Errorf(invalidFind, "subnet")
	}

	defer func() {

		if err == nil {
			err := client.DeleteSubnet(vpc.ID, found)
			if err != nil {
				t.Fatal(err)
			}
		}

		err = client.WaitForSubnetStatus(subnet.ID, "")
		switch err.(type) {
		case golangsdk.ErrDefault404:
		default:
			t.Error(err)
		}

		if err := client.DeleteVPC(vpc.ID); err != nil {
			t.Error(err)
			return
		}
	}()
}
