package app

import (
	"fmt"

	"github.com/okteto/app/api/k8s/client"
	"github.com/okteto/app/api/k8s/deployments"
	"github.com/okteto/app/api/k8s/ingresses"
	"github.com/okteto/app/api/k8s/secrets"
	"github.com/okteto/app/api/k8s/services"
	"github.com/okteto/app/api/k8s/volumes"
	"github.com/okteto/app/api/model"
)

//DevModeOn activates a development environment
func DevModeOn(u *model.User, dev *model.Dev) error {
	if len(dev.Volumes) > 2 {
		return fmt.Errorf("the maximum number of volumes is 2")
	}
	s := &model.Space{
		ID:   u.ID,
		Name: u.GithubID,
	}
	c, err := client.Get()
	if err != nil {
		return fmt.Errorf("error getting k8s client: %s", err)
	}

	if err := secrets.Create(dev, s, c); err != nil {
		return err
	}

	if err := volumes.Create(dev.GetVolumeName(), s, c); err != nil {
		return err
	}

	for i := range dev.Volumes {
		if err := volumes.Create(dev.GetVolumeDataName(i), s, c); err != nil {
			return err
		}
	}

	if err := deployments.DevOn(dev, s, c); err != nil {
		return err
	}

	new := services.Translate(dev, s)
	if err := services.Deploy(new, s, c); err != nil {
		return err
	}

	if err := ingresses.Deploy(dev, s, c); err != nil {
		return err
	}

	return nil
}

//RunImage runs a docker image
func RunImage(u *model.User, dev *model.Dev) error {
	s := &model.Space{
		ID:   u.ID,
		Name: u.GithubID,
	}
	c, err := client.Get()
	if err != nil {
		return fmt.Errorf("error getting k8s client: %s", err)
	}

	if err := deployments.Run(dev, s, c); err != nil {
		return err
	}

	new := services.Translate(dev, s)
	if err := services.Deploy(new, s, c); err != nil {
		return err
	}

	if err := ingresses.Deploy(dev, s, c); err != nil {
		return err
	}

	return nil
}

//DevModeOff deactivates a development environment
func DevModeOff(u *model.User, dev *model.Dev, removeVolumes bool) error {
	s := &model.Space{
		ID:   u.ID,
		Name: u.GithubID,
	}
	c, err := client.Get()
	if err != nil {
		return fmt.Errorf("error getting k8s client: %s", err)
	}

	dev = deployments.GetDev(dev, s, c)

	if err := ingresses.Destroy(dev, s, c); err != nil {
		return err
	}

	if err := services.Destroy(dev.Name, s, c); err != nil {
		return err
	}

	if err := deployments.Destroy(dev, s, c); err != nil {
		return err
	}

	if err := volumes.Destroy(dev.GetVolumeName(), s, c); err != nil {
		return err
	}

	for i := range dev.Volumes {
		if err := volumes.Destroy(dev.GetVolumeDataName(i), s, c); err != nil {
			return err
		}
	}

	if err := secrets.Destroy(dev, s, c); err != nil {
		return err
	}

	return nil
}
