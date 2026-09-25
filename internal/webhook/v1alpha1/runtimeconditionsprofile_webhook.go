/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	resolver "github.com/runtimeconditions/rc-extension-resolver"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var runtimeconditionsprofilelog = logf.Log.WithName("runtimeconditionsprofile-resource")

// +kubebuilder:webhook:path=/validate-runtimeconditions-io-v1alpha1-runtimeconditionsprofile,mutating=false,failurePolicy=fail,sideEffects=None,groups=runtimeconditions.io,resources=runtimeconditionsprofiles,verbs=create;update,versions=v1alpha1,name=vruntimeconditionsprofile-v1alpha1.kb.io,admissionReviewVersions=v1

// SetupRuntimeConditionsProfileWebhookWithManager registers the webhook for RuntimeConditionsProfile in the manager.
func SetupRuntimeConditionsProfileWebhookWithManager(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().Register(
		"/validate-runtimeconditions-io-v1alpha1-runtimeconditionsprofile",
		&admission.Webhook{Handler: &RuntimeConditionsProfileValidator{Loader: resolver.NewHTTPLoader(10 * time.Second)}},
	)
	return nil
}

// RuntimeConditionsProfileValidator validates RuntimeConditionsProfile
// against the extensions it declares.
//
// This is a raw admission.Handler, not a CustomValidator: Condition.Interface
// carries extension-defined fields the CRD's Go type doesn't know about
// (it's typed as PreserveUnknownFields for exactly that reason). Decoding
// into that typed struct would silently drop them before validation ever
// saw them, so we decode the admission request's raw JSON into a
// map[string]any instead.
type RuntimeConditionsProfileValidator struct {
	Loader resolver.LoaderFunc
}

type profileForValidation struct {
	Extensions []string         `json:"extensions"`
	Conditions []map[string]any `json:"conditions"`
}

func (v *RuntimeConditionsProfileValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	var profile profileForValidation
	if err := json.Unmarshal(req.Object.Raw, &profile); err != nil {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("decoding RuntimeConditionsProfile: %w", err))
	}

	runtimeconditionsprofilelog.Info("Validating RuntimeConditionsProfile", "name", req.Name, "namespace", req.Namespace)

	graph, err := resolveAll(v.Loader, profile.Extensions)
	if err != nil {
		return admission.Denied(fmt.Sprintf("resolving extensions: %v", err))
	}
	catalog, err := resolver.Merge(graph)
	if err != nil {
		return admission.Denied(fmt.Sprintf("merging extensions: %v", err))
	}

	var problems []string
	for _, condition := range profile.Conditions {
		if msg := validateCondition(catalog, condition); msg != "" {
			problems = append(problems, msg)
		}
	}

	if len(problems) > 0 {
		return admission.Denied(strings.Join(problems, "; "))
	}
	return admission.Allowed("")
}

func validateCondition(catalog *resolver.Catalog, condition map[string]any) string {
	kind, _ := condition["kind"].(string)
	label := kind
	if name, _ := condition["name"].(string); name != "" {
		label = fmt.Sprintf("%s (kind %s)", name, kind)
	}

	if !catalog.IsValidKind(kind) {
		return fmt.Sprintf("condition %s: unknown kind %q", label, kind)
	}

	iface, _ := condition["interface"].(map[string]any)
	interfaceType, _ := iface["type"].(string)
	if !catalog.IsValidInterfaceType(kind, interfaceType) {
		return fmt.Sprintf("condition %s: unknown interface type %q for kind %q", label, interfaceType, kind)
	}

	result, err := catalog.ValidateCondition(condition)
	if err != nil {
		return fmt.Sprintf("condition %s: %v", label, err)
	}
	if !result.Valid {
		return fmt.Sprintf("condition %s: %s", label, strings.Join(result.Errors, "; "))
	}
	return ""
}

// resolveAll resolves every root in rootURIs and combines them into one
// graph, deduplicated by extension identifier. A profile can declare
// several independent extensions, not just one dependency tree, so this
// fans Resolver.Resolve out across all of them before a single Merge.
func resolveAll(loader resolver.LoaderFunc, rootURIs []string) (*resolver.ResolvedGraph, error) {
	r := resolver.NewResolver(loader)
	byID := make(map[string]*resolver.ExtensionDefinition)
	var order []*resolver.ExtensionDefinition

	for _, root := range rootURIs {
		graph, err := r.Resolve(root)
		if err != nil {
			return nil, err
		}
		for _, ext := range graph.Extensions {
			if _, ok := byID[ext.Metadata.ID]; ok {
				continue
			}
			byID[ext.Metadata.ID] = ext
			order = append(order, ext)
		}
	}

	return &resolver.ResolvedGraph{Extensions: order, ByID: byID}, nil
}
