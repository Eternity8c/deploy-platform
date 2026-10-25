package features_application

import (
	"regexp"
	"strings"
)

const (
	apiVersionV1    = "platform.deploy/v1"
	kindApplication = "Application"
	defaultTenant   = "shop"
)

var dnsName = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func (a *Application) Validate() error {
	if a == nil {
		return ValidationErrors{{Message: "application is nil"}}
	}

	var errs ValidationErrors
	add := func(field, msg string) {
		errs = append(errs, ValidationError{Field: field, Message: msg})
	}

	if a.APIVersion != apiVersionV1 {
		add("apiVersion", "must be "+apiVersionV1)
	}

	if a.Kind != kindApplication {
		add("kind", "must be "+kindApplication)
	}

	if len(a.Metadata.Name) == 0 || len(a.Metadata.Name) > 63 || !dnsName.MatchString(a.Metadata.Name) {
		add("metadata.name", "must be a lowercase DNS label (a-z, 0-9, '-'), 1-63 chars")
	}

	switch {
	case a.Spec.Tag == "":
		add("spec.tag", "must not be empty")
	case a.Spec.Tag == "latest":
		add("spec.tag", `must be a concrete tag: "latest" is not reproducible`)
	}

	switch {
	case a.Spec.Image == "":
		add("spec.image", "must not be empty")
	case strings.Contains(a.Spec.Image, "@"):
		add("spec.image", "must not contain a digest; put the tag into spec.tag")
	case strings.LastIndex(a.Spec.Image, ":") > strings.LastIndex(a.Spec.Image, "/"):
		add("spec.image", "must not contain a tag; the tag goes into spec.tag")
	}

	if a.Spec.Replicas < 0 {
		add("spec.replicas", "must not be negative")
	}

	if a.Spec.Port < 1 || a.Spec.Port > 65535 {
		add("spec.port", "must be between 1 and 65535")
	}

	if len(errs) > 0 {
		return errs
	}

	if a.Spec.Tenant == "" {
		a.Spec.Tenant = defaultTenant
	}

	return nil
}
