package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	assert := assert.New(t)
	authConfig := &AuthConfig{}

	authConfig.RemoveAuthFile()

	assert.Equal(false, authConfig.AuthFileExists())
}

func TestWriteAuth(t *testing.T) {
	assert := assert.New(t)
	authJson := &AuthJson{Token: "abcd1234"}
	authConfig := &AuthConfig{AuthJson: authJson}
	authConfig.RemoveAuthFile()

	authConfig.WriteAuth()

	assert.Equal(true, authConfig.AuthFileExists())
}

func TestRemoveAuthFile(t *testing.T) {
	assert := assert.New(t)
	authJson := &AuthJson{Token: "abcd1234"}
	authConfig := &AuthConfig{AuthJson: authJson}
	authConfig.RemoveAuthFile()

	authConfig.WriteAuth()
	assert.Equal(true, authConfig.AuthFileExists())

	authConfig.RemoveAuthFile()
	assert.Equal(false, authConfig.AuthFileExists())
}
