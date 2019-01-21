// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	"github.com/postfinance/mage/jfrog"
	// mg contains helpful utility functions, like Deps
)

// Upload uploads created rpm to artifactory
func Upload() error {
	fmt.Println("Uploading...")
	rtDetails := auth.NewArtifactoryDetails()
	rtDetails.SetUrl(os.Getenv("URL"))
	rtDetails.SetUser(os.Getenv("USER"))
	rtDetails.SetPassword(os.Getenv("PASSWORD"))
	u, err := jfrog.New(rtDetails)
	if err != nil {
		return err
	}
	return u.Upload("dist/*.rpm", os.Getenv("TARGET"))
}
