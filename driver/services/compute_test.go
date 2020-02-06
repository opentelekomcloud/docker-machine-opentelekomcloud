package services

import "testing"

func TestClient_CreateSecurityGroup(t *testing.T) {
	client, err := authClient()
	if err != nil {
		t.Error(err)
		return
	}
	if err := client.InitCompute(); err != nil {
		t.Error(err)
		return
	}
	sg, err := client.CreateSecurityGroup(sgName)
	if err != nil {
		t.Error(err)
		return
	}

	sgID, err := client.FindSecurityGroup(sgName)
	if err != nil {
		t.Error(err)
	}
	if sgID != sg.ID {
		t.Errorf(invalidFind, "sec grp")
	}

	err = client.DeleteSecurityGroup(sg.ID)
	if err != nil {
		t.Error(err)
	}
}
