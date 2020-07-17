package jfrog

import (
	"testing"

	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateParams(t *testing.T) {
	var tt = []struct {
		sourcePattern  string
		targetString   string
		expectsError   bool
		expectedParams services.UploadParams
	}{
		{
			"*.deb",
			"repo/dir:/buster/main/amd64",
			true,
			services.UploadParams{},
		},
		{
			"*.deb",
			"repo/dir:/buster/main/",
			true,
			services.UploadParams{},
		},
		{
			"*.deb",
			"repo/dir:buster/main",
			true,
			services.UploadParams{},
		},
		{
			"*.deb",
			"repo/dir:buster/main/amd64",
			false,
			services.UploadParams{
				ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{
					Pattern: "*.deb",
					Target:  "repo/dir",
				},
				Deb:     "buster/main/amd64",
				Retries: 1,
			},
		},
		{
			"*.rpm",
			"repo/dir",
			false,
			services.UploadParams{
				ArtifactoryCommonParams: &utils.ArtifactoryCommonParams{
					Pattern: "*.rpm",
					Target:  "repo/dir",
				},
				Retries: 1,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.sourcePattern, func(t *testing.T) {
			p, err := createParams(tc.sourcePattern, tc.targetString)
			if tc.expectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, tc.expectedParams, p)
		})
	}
}
