package jfrog

import (
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// Uploader is responsible for uploading artifacts to an artifactory repository.
type Uploader struct {
	rtManager *artifactory.ArtifactoryServicesManager
}

// New creates an Uploader.
func New(rtDetails auth.ServiceDetails) (*Uploader, error) {
	log.SetLogger(log.NewLogger(log.INFO, nil))
	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(rtDetails).
		SetDryRun(false).
		Build()
	if err != nil {
		return nil, err
	}
	rtManager, err := artifactory.New(&rtDetails, serviceConfig)
	if err != nil {
		return nil, err
	}
	return &Uploader{rtManager: rtManager}, nil
}

// Upload uploads all sources matching srcPattern to target.
func (u Uploader) Upload(srcPattern, target string) error {
	params := services.UploadParams{
		ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{},
		Deb:                     "",
		Symlink:                 false,
		ExplodeArchive:          false,
		Flat:                    false,
		Retries:                 1,
	}
	params.Pattern = srcPattern
	params.Target = target
	_, _, _, err := u.rtManager.UploadFiles(params)
	if err != nil {
		return err
	}
	return nil
}
