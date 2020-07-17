package jfrog

import (
	"errors"
	"strings"

	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

var (
	ErrWrongDebianFormat = errors.New("target has to be of the form '<destination>:<distribution>/<component>/<architecture>' for debian packages")
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
//
// If package ends with '.deb' the target has to be of the following
// form: '<destination>:<distribution>/<component>/<architecture>'
//
// The following is a valid debian target: 'local-deb-repo/pool:buster/main/amd64'.
func (u Uploader) Upload(srcPattern, target string) error {
	params, err := createParams(srcPattern, target)
	if err != nil {
		return err
	}

	_, _, _, err = u.rtManager.UploadFiles(params)
	if err != nil {
		return err
	}
	return nil
}

func createParams(srcPattern, target string) (services.UploadParams, error) {
	params := services.UploadParams{
		ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{
			Pattern: srcPattern,
			Target:  target,
		},
		Deb:            "",
		Symlink:        false,
		ExplodeArchive: false,
		Flat:           false,
		Retries:        1,
	}

	if strings.HasSuffix(srcPattern, ".deb") {
		splitted := strings.Split(target, ":")

		if len(splitted) != 2 {
			return params, ErrWrongDebianFormat
		}

		count := 0

		for _, item := range strings.Split(splitted[1], "/") {
			count++
			if item == "" {
				return params, ErrWrongDebianFormat
			}
		}

		if count != 3 {
			return params, ErrWrongDebianFormat
		}

		params.Target = splitted[0]
		params.Deb = splitted[1]
	}

	return params, nil
}
