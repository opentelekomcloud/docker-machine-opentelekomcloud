package services

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClient_CreateSecurityGroup(t *testing.T) {
	client := authClient(t)
	require.NoError(t, client.InitCompute())
	sg, err := client.CreateSecurityGroup(sgName)
	require.NoError(t, err)
	if err != nil {
		t.Error(err)
		return
	}

	sgID, err := client.FindSecurityGroup(sgName)
	assert.NoError(t, err)
	assert.EqualValuesf(t, sg.ID, sgID, invalidFind, "subnet")

	assert.NoError(t, client.DeleteSecurityGroup(sg.ID))
}
