package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/twinj/uuid"
)

type AssetManager struct{}

type AwsCreds struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
}

func (a *AssetManager) PushLocalAssets(token string, deckConfig *DeckConfig) *DeckConfig {
	localFiles := a.readFiles()
	uploader := a.setupUploader(token)
	for _, fileName := range localFiles {
		var found bool
		for _, asset := range deckConfig.Assets {
			if asset.Filename == fileName {
				// asset is both local and remote
				// TODO handle if asset has changed?
				found = true
			}
		}

		if !found {
			fmt.Println("Uploading ", fileName)
			asset := a.uploadFile(fileName, uploader)
			deckConfig.Assets = append(deckConfig.Assets, asset)
		}
	}

	// handle the case where there is a remote asset that is not local
	for i, asset := range deckConfig.Assets {
		var found bool
		for _, fileName := range localFiles {
			if asset.Filename == fileName {
				found = true
			}
		}
		if !found {
			fmt.Printf("The file %s exists on app.ultradeck.co, but not locally.  Do you want to delete it from your deck? (y/n) ", asset.Filename)
			reader := bufio.NewReader(os.Stdin)
			name, _ := reader.ReadString('\n')
			if name == "y\n" {
				deckConfig.Assets = append(deckConfig.Assets[:i], deckConfig.Assets[i+1:]...)
			}

		}
	}
	return deckConfig
}

func (a *AssetManager) PullRemoteAssets(deckConfig *DeckConfig) {
	localFiles := a.readFiles()
	for _, asset := range deckConfig.Assets {
		var found bool
		for _, fileName := range localFiles {
			if asset.Filename == fileName {
				// asset is both local and remote
				// TODO handle if asset has changed?
				found = true
			}
		}

		if !found {
			fmt.Println("Downloading ", asset.Filename)
			a.downloadFile(asset)
		}
	}
}

func (a *AssetManager) setupUploader(token string) *s3manager.Uploader {
	httpClient := NewHttpClient(token)
	jsonData := httpClient.GetRequest("/api/v1/auth/aws_creds")
	awsCreds := &AwsCreds{}
	json.Unmarshal(jsonData, awsCreds)

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(endpoints.UsEast1RegionID),
		Credentials: credentials.NewStaticCredentials(awsCreds.AccessKeyID, awsCreds.SecretAccessKey, awsCreds.SessionToken),
	}))
	return s3manager.NewUploader(sess)
}

func (a *AssetManager) downloadFile(asset *Asset) {
	resp, err := http.Get(asset.URL)
	if err != nil {
		fmt.Println("Error downloading asset: ", err)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err := ioutil.WriteFile(asset.Filename, body, 0644); err != nil {
		fmt.Println("Couldn't write file "+asset.Filename, err)
	}
}

func (a *AssetManager) uploadFile(fileName string, uploader *s3manager.Uploader) *Asset {
	keyName := fmt.Sprintf("/uploads/%s/%s", uuid.NewV4(), fileName)

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)

	bucketName := "ultradeck-assets-dev"
	acl := "public-read"
	mimeType := a.mimeType(fileName)

	upParams := &s3manager.UploadInput{
		Bucket:      &bucketName,
		Key:         &keyName,
		Body:        fileBytes,
		ACL:         &acl,
		ContentType: &mimeType,
	}

	result, err := uploader.Upload(upParams)
	if err != nil {
		fmt.Println("error uploading file: ", err)
	}

	asset := &Asset{Filename: fileName, URL: result.Location}
	return asset
}

func (a *AssetManager) readFiles() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal("Error reading directory: ", err)
	}

	var ret []string

	for _, file := range files {
		// TODO: support more extension types?
		if strings.HasPrefix(a.mimeType(file.Name()), "image") {
			ret = append(ret, file.Name())
		}
	}

	return ret
}

func (a *AssetManager) mimeType(fileName string) string {
	splitted := strings.Split(fileName, ".")
	ext := splitted[len(splitted)-1]
	ext = "." + ext
	mimeType := mime.TypeByExtension(ext)
	return mimeType
}
