// Package common holds reference configuration shared by every group package
// in both API trees.
package common

import (
	"fmt"
	"strings"

	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Organization references sentry_organization. The reference is optional: the
// organization usually pre-exists, so users may keep passing the slug as a
// plain string.
func Organization() config.Reference {
	return config.Reference{TerraformName: "sentry_organization"}
}

// Project references sentry_project; the Terraform id of a project is its slug.
func Project() config.Reference {
	return config.Reference{TerraformName: "sentry_project"}
}

// Team references sentry_team; the Terraform id of a team is its slug.
func Team() config.Reference {
	return config.Reference{TerraformName: "sentry_team"}
}

// TeamInternalID references sentry_team and extracts the numeric internal_id,
// which is what owner.team_id expects.
func TeamInternalID() config.Reference {
	return config.Reference{
		TerraformName: "sentry_team",
		Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("internal_id",true)`,
	}
}

// OrganizationMember references sentry_organization_member; its external name
// is the member id.
func OrganizationMember() config.Reference {
	return config.Reference{TerraformName: "sentry_organization_member"}
}

// Omit removes deprecated or redundant attributes from the generated API. Paths
// are dot separated Terraform attribute names ("action_filters.actions.plugin").
// It panics when a path does not exist so that a provider version bump that
// renames or drops an attribute fails at generation time instead of silently
// re-exposing it.
func Omit(r *config.Resource, paths ...string) {
	for _, p := range paths {
		parts := strings.Split(p, ".")
		sch := r.TerraformResource.Schema
		for _, part := range parts[:len(parts)-1] {
			s, ok := sch[part]
			if !ok {
				panic(fmt.Sprintf("%s: cannot omit %q: %q not found", r.Name, p, part))
			}
			el, ok := s.Elem.(*schema.Resource)
			if !ok {
				panic(fmt.Sprintf("%s: cannot omit %q: %q is not an object", r.Name, p, part))
			}
			sch = el.Schema
		}
		last := parts[len(parts)-1]
		if _, ok := sch[last]; !ok {
			panic(fmt.Sprintf("%s: cannot omit %q: attribute not found", r.Name, p))
		}
		delete(sch, last)
	}
}
