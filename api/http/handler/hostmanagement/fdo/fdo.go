package fdo

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/hostmanagement/fdo"
	"github.com/sirupsen/logrus"
)

type fdoConfigurePayload portainer.FDOConfiguration

func validateURL(u string) error {
	p, err := url.Parse(u)
	if err != nil {
		return err
	}

	if p.Scheme != "http" && p.Scheme != "https" {
		return errors.New("invalid scheme provided, must be 'http' or 'https'")
	}

	if p.Host == "" {
		return errors.New("invalid host provided")
	}

	return nil
}

func (payload *fdoConfigurePayload) Validate(r *http.Request) error {
	if payload.Enabled {
		if err := validateURL(payload.OwnerURL); err != nil {
			return fmt.Errorf("owner server URL: %w", err)
		}

		if err := validateURL(payload.ProfilesURL); err != nil {
			return fmt.Errorf("profile list URL: %w", err)
		}
	}

	return nil
}

func (handler *Handler) saveSettings(config portainer.FDOConfiguration) error {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return err
	}
	settings.FDOConfiguration = config

	return handler.DataStore.Settings().UpdateSettings(settings)
}

func (handler *Handler) newFDOClient() (fdo.FDOOwnerClient, error) {
	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return fdo.FDOOwnerClient{}, err
	}

	return fdo.FDOOwnerClient{
		OwnerURL: settings.FDOConfiguration.OwnerURL,
		Username: settings.FDOConfiguration.OwnerUsername,
		Password: settings.FDOConfiguration.OwnerPassword,
		Timeout:  5 * time.Second,
	}, nil
}

// @id fdoConfigure
// @summary Enable Portainer's FDO capabilities
// @description Enable Portainer's FDO capabilities
// @description **Access policy**: administrator
// @tags intel
// @security jwt
// @accept json
// @produce json
// @param body body fdoConfigurePayload true "FDO Settings"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access settings"
// @failure 500 "Server error"
// @router /fdo [post]
func (handler *Handler) fdoConfigure(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload fdoConfigurePayload

	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		logrus.WithError(err).Error("Invalid request payload")
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings := portainer.FDOConfiguration(payload)
	if err = handler.saveSettings(settings); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Error saving FDO settings", Err: err}
	}

	return response.Empty(w)
}
