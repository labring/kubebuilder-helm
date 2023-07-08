/*
Copyright 2022 The Kubernetes Authors.

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

package api

import (
	"fmt"
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Types{}

// Types scaffolds the file that defines the schema for a CRD
// nolint:maligned
type Types struct {
	machinery.TemplateMixin
	machinery.MultiGroupMixin
	machinery.BoilerplateMixin
	machinery.ResourceMixin

	Force bool
}

// SetTemplateDefaults implements file.Template
func (f *Types) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.MultiGroup && f.Resource.Group != "" {
			f.Path = filepath.Join("api", "%[group]", "%[version]", "%[kind]_types.go")
		} else {
			f.Path = filepath.Join("api", "%[version]", "%[kind]_types.go")
		}
	}

	f.Path = f.Resource.Replacer().Replace(f.Path)
	fmt.Println(f.Path)

	f.TemplateBody = typesTemplate

	if f.Force {
		f.IfExistsAction = machinery.OverwriteFile
	} else {
		f.IfExistsAction = machinery.Error
	}

	return nil
}

const typesTemplate = `{{ .Boilerplate }}

package {{ .Resource.Version }}

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// {{ .Resource.Kind }}Spec defines the desired state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of {{ .Resource.Kind }}. Edit {{ lower .Resource.Kind }}_types.go to remove/update
	Foo string ` + "`" + `json:"foo,omitempty"` + "`" + `
}

type {{ .Resource.Kind }}Phase string

// These are the valid phases of node.
const (
	{{ .Resource.Kind }}Pending {{ .Resource.Kind }}Phase = "Pending"
	{{ .Resource.Kind }}Unknown {{ .Resource.Kind }}Phase = "Unknown"
	{{ .Resource.Kind }}Active  {{ .Resource.Kind }}Phase = "Active"
)

// {{ .Resource.Kind }}Status defines the observed state of {{ .Resource.Kind }}
type {{ .Resource.Kind }}Status struct {
	// Phase represents the current phase of {{ .Resource.Kind }}.
	//+kubebuilder:default:=Unknown
	Phase {{ .Resource.Kind }}Phase ` + "`" + `json:"phase,omitempty"` + "`" + `
	// Represents the observations of a {{ .Resource.Kind }}'s current state.
	// {{ .Resource.Kind }}.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// {{ .Resource.Kind }}.status.conditions.status are one of True, False, Unknown.
	// {{ .Resource.Kind }}.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// {{ .Resource.Kind }}.status.conditions.Message is a human readable message indicating details about the transition.
	// For further information see: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	Conditions []metav1.Condition ` + "`" + `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"` + "`" + `
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
{{- if and (not .Resource.API.Namespaced) (not .Resource.IsRegularPlural) }}
//+kubebuilder:resource:path={{ .Resource.Plural }},scope=Cluster
{{- else if not .Resource.API.Namespaced }}
//+kubebuilder:resource:scope=Cluster
{{- else if not .Resource.IsRegularPlural }}
//+kubebuilder:resource:path={{ .Resource.Plural }}
{{- end }}

// {{ .Resource.Kind }} is the Schema for the {{ .Resource.Plural }} API
type {{ .Resource.Kind }} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	Spec   {{ .Resource.Kind }}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{ .Resource.Kind }}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

//+kubebuilder:object:root=true

// {{ .Resource.Kind }}List contains a list of {{ .Resource.Kind }}
type {{ .Resource.Kind }}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{ .Resource.Kind }} ` + "`" + `json:"items"` + "`" + `
}

func init() {
	SchemeBuilder.Register(&{{ .Resource.Kind }}{}, &{{ .Resource.Kind }}List{})
}
`
