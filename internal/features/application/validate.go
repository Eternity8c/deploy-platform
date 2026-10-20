package features_application

import (
	"regexp"
	"strings"
)

var dnsName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func (a *Application) Validate() error {
	if a == nil {
		return invalid("", "application is nil")
	}
	if a.APIVersion != "platform.deploy/v1" {
		return invalid("apiVersion", "must be platform.deploy/v1")
	}
	if a.Kind != "Application" {
		return invalid("kind", "must be Application")
	}
	if !dnsName.MatchString(a.Metadata.Name) || len(a.Metadata.Name) > 63 {
		return invalid("metadata.name", "incorrect name")
	}
	if a.Spec.Tag == "" {
		return invalid("spec.tag", "invalid name")
	}
	if a.Spec.Tag == "latest" {
		return invalid("spec.tag", "tag is latest")
	}
	if a.Spec.Image == "" {
		return invalid("spec.image", "invalid name")
	}
	if strings.Contains(a.Spec.Image, "@") {
		return invalid("spec.image", "must not contain a digest")
	}
	if strings.LastIndex(a.Spec.Image, ":") > strings.LastIndex(a.Spec.Image, "/") {
		return invalid("spec.image", "must not contain a tag; the tag goes into spec.tag")
	}
	if a.Spec.Replicas < 0 {
		return invalid("spec.replicas", "cannot be less than zero")
	}
	if a.Spec.Port < 1 || a.Spec.Port > 65535 {
		return invalid("spec.port", "invalid port")
	}
	if a.Spec.Tenant == "" {
		a.Spec.Tenant = "shop"
	}
	return nil
}
