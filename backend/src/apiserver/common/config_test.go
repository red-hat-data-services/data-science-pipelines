package common

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateServiceAccountAllowList_AllowedSA(t *testing.T) {
	viper.Reset()
	viper.Set(AllowedServiceAccountsFlag, "custom-sa")
	err := ValidateServiceAccountAllowList("custom-sa")
	assert.Nil(t, err)
}

func TestValidateServiceAccountAllowList_DisallowedSA(t *testing.T) {
	viper.Reset()
	viper.Set(AllowedServiceAccountsFlag, "other-sa")
	err := ValidateServiceAccountAllowList("evil-sa")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateServiceAccountAllowList_DefaultAlwaysAllowed(t *testing.T) {
	viper.Reset()
	err := ValidateServiceAccountAllowList(DefaultPipelineRunnerServiceAccount)
	assert.Nil(t, err)
}

func TestValidateServiceAccountAllowList_EmptyListRejectsCustom(t *testing.T) {
	viper.Reset()
	err := ValidateServiceAccountAllowList("custom-sa")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateServiceAccountAllowList_MultipleSAs(t *testing.T) {
	viper.Reset()
	viper.Set(AllowedServiceAccountsFlag, "sa1,sa2,sa3")
	err := ValidateServiceAccountAllowList("sa2")
	assert.Nil(t, err)
}

func TestValidateServiceAccountAllowList_WhitespaceTrimming(t *testing.T) {
	viper.Reset()
	viper.Set(AllowedServiceAccountsFlag, " sa1 , sa2 ")
	err := ValidateServiceAccountAllowList("sa1")
	assert.Nil(t, err)
}

func TestValidateServiceAccountAllowList_EmptyStringAllowed(t *testing.T) {
	viper.Reset()
	err := ValidateServiceAccountAllowList("")
	assert.Nil(t, err)
}

func TestValidateServiceAccountAllowList_ErrorDoesNotLeakAllowList(t *testing.T) {
	viper.Reset()
	viper.Set(AllowedServiceAccountsFlag, "secret-sa-1,secret-sa-2")
	err := ValidateServiceAccountAllowList("evil-sa")
	require.NotNil(t, err)
	assert.NotContains(t, err.Error(), "secret-sa-1")
	assert.NotContains(t, err.Error(), "secret-sa-2")
}

func TestValidateServiceAccountAllowList_ConfiguredDefaultAllowed(t *testing.T) {
	viper.Reset()
	viper.Set(DefaultPipelineRunnerServiceAccountFlag, "my-runner")
	err := ValidateServiceAccountAllowList("my-runner")
	assert.Nil(t, err)
}
