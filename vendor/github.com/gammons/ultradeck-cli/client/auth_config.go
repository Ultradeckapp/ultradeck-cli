package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type AuthConfig struct {
	AuthJson *AuthJson
}

type AuthJson struct {
	Token            string `json:"token"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	ImageUrl         string `json:"image_url"`
	Email            string `json:"email"`
	SubscriptionName string `json:"subscription_name"`
}

func NewAuthConfig(response map[string]interface{}) *AuthConfig {
	authJson := &AuthJson{
		Token:            response["access_token"].(string),
		Username:         response["username"].(string),
		Name:             response["name"].(string),
		ImageUrl:         response["image_url"].(string),
		Email:            response["email"].(string),
		SubscriptionName: response["subscription_name"].(string),
	}

	return &AuthConfig{AuthJson: authJson}
}

func (c *AuthConfig) AuthFileExists() bool {
	if _, err := os.Stat(c.configFileLocation()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *AuthConfig) WriteAuth() {
	data, _ := json.Marshal(c.AuthJson)

	if c.AuthFileExists() {
		c.RemoveAuthFile()
	}

	if err := os.MkdirAll(c.configFilePath(), os.ModePerm); err != nil {
		os.Exit(1)
	}

	if err := ioutil.WriteFile(c.configFileLocation(), []byte(data), 0644); err != nil {
		log.Println("Error writing json file", err)
	}
}

func (c *AuthConfig) ReadConfig() *AuthJson {
	if !c.AuthFileExists() {
		return nil
	}

	data, err := ioutil.ReadFile(c.configFileLocation())
	if err != nil {
		log.Println("error reading auth config file: ", err)
	}

	var authJson *AuthJson
	err = json.Unmarshal(data, &authJson)
	if err != nil {
		log.Println("error reading auth config json: ", err)
	}
	return authJson
}

func (c *AuthConfig) GetToken() string {
	authJson := c.ReadConfig()
	return authJson.Token
}

func (c *AuthConfig) RemoveAuthFile() {
	if !c.AuthFileExists() {
		return
	}

	if err := os.RemoveAll(c.configFilePath()); err != nil {
		log.Println("Error removing config file", err)
	}
}

func (c *AuthConfig) configFilePath() string {
	usr, _ := user.Current()
	return fmt.Sprintf("%s/.config/ultradeck/", usr.HomeDir)
}

func (c *AuthConfig) configFileLocation() string {
	return c.configFilePath() + "auth.json"
}
