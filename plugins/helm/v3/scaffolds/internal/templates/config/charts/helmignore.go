/*
Copyright 2020 The Kubernetes Authors.

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

package charts

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &HelmIgnore{}

// HelmIgnore scaffolds a file that defines the kustomization scheme for the webhook folder
type HelmIgnore struct {
	machinery.TemplateMixin
	machinery.ProjectNameMixin

	Force bool
}

// SetTemplateDefaults implements file.Template
func (f *HelmIgnore) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join("config", "charts", f.ProjectName, ".helmignore")
	}

	f.TemplateBody = helmIgnoreTemplate

	if f.Force {
		f.IfExistsAction = machinery.OverwriteFile
	} else {
		// If file exists (ex. because a webhook was already created), skip creation.
		f.IfExistsAction = machinery.SkipFile
	}

	return nil
}

const helmIgnoreTemplate = `# Patterns to ignore when building packages.
# This supports shell glob matching, relative path matching, and
# negation (prefixed with !). Only one pattern per line.
.DS_Store
# Common VCS dirs
.git/
.gitignore
.bzr/
.bzrignore
.hg/
.hgignore
.svn/
# Common backup files
*.swp
*.bak
*.tmp
*.orig
*~
# Various IDEs
.project
.idea/
*.tmproj
.vscode/

`
