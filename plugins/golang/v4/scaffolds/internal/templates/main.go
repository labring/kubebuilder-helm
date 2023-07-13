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

package templates

import (
	"fmt"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

const defaultMainPath = "cmd/main.go"
const defaultLegacyLayoutMainPath = "main.go"

var _ machinery.Template = &Main{}

// Main scaffolds a file that defines the controller manager entry point
type Main struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.DomainMixin
	machinery.RepositoryMixin
	// IsLegacyLayout is added to ensure backwards compatibility and should
	// be removed when we remove the go/v3 plugin
	IsLegacyLayout bool
}

// SetTemplateDefaults implements file.Template
func (f *Main) SetTemplateDefaults() error {
	if f.Path == "" {
		if f.IsLegacyLayout {
			f.Path = filepath.Join(defaultLegacyLayoutMainPath)
		} else {
			f.Path = filepath.Join(defaultMainPath)
		}
	}

	f.TemplateBody = fmt.Sprintf(mainTemplate,
		machinery.NewMarkerFor(f.Path, importMarker),
		machinery.NewMarkerFor(f.Path, addSchemeMarker),
		machinery.NewMarkerFor(f.Path, setupMarker),
	)

	return nil
}

var _ machinery.Inserter = &MainUpdater{}

// MainUpdater updates cmd/main.go to run Controllers
type MainUpdater struct { //nolint:maligned
	machinery.RepositoryMixin
	machinery.MultiGroupMixin
	machinery.ResourceMixin

	// Flags to indicate which parts need to be included when updating the file
	WireResource, WireController, WireWebhook bool
	// IsLegacyLayout is added to ensure backwards compatibility and should
	// be removed when we remove the go/v3 plugin
	IsLegacyLayout bool
}

// GetPath implements file.Builder
func (f *MainUpdater) GetPath() string {
	if f.IsLegacyLayout {
		return defaultLegacyLayoutMainPath
	} else {
		return defaultMainPath
	}
}

// GetIfExistsAction implements file.Builder
func (*MainUpdater) GetIfExistsAction() machinery.IfExistsAction {
	return machinery.OverwriteFile
}

const (
	importMarker    = "imports"
	addSchemeMarker = "scheme"
	setupMarker     = "builder"
)

// GetMarkers implements file.Inserter
func (f *MainUpdater) GetMarkers() []machinery.Marker {
	return []machinery.Marker{
		machinery.NewMarkerFor(f.GetPath(), importMarker),
		machinery.NewMarkerFor(f.GetPath(), addSchemeMarker),
		machinery.NewMarkerFor(f.GetPath(), setupMarker),
	}
}

const (
	apiImportCodeFragment = `%s "%s"
`
	controllerImportCodeFragment = `"%s/internal/controller"
`
	multiGroupControllerImportCodeFragment = `%scontroller "%s/internal/controller/%s"
`
	addschemeCodeFragment = `utilruntime.Must(%s.AddToScheme(scheme))
`
	reconcilerSetupCodeFragment = `if err = (&controller.%sReconciler{
		MaxConcurrentReconciles: concurrent,
		RateLimiter: utilcontroller.GetRateLimiter(rateLimiterOptions),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "%s")
		os.Exit(1)
	}
`
	multiGroupReconcilerSetupCodeFragment = `if err = (&%scontroller.%sReconciler{
		MaxConcurrentReconciles: concurrent,
		RateLimiter: utilcontroller.GetRateLimiter(rateLimiterOptions),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "%s")
		os.Exit(1)
	}
`
	webhookSetupCodeFragment = `if os.Getenv("DISABLE_WEBHOOKS") != "true" {
		if err = (&%s.%s{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "%s")
			os.Exit(1)
		}else{
			setupLog.Info("webhook disable", "webhook", "%s")
		}
	}
`
)

// GetCodeFragments implements file.Inserter
func (f *MainUpdater) GetCodeFragments() machinery.CodeFragmentsMap {
	fragments := make(machinery.CodeFragmentsMap, 3)

	// If resource is not being provided we are creating the file, not updating it
	if f.Resource == nil {
		return fragments
	}

	// Generate import code fragments
	imports := make([]string, 0)
	if f.WireResource {
		imports = append(imports, fmt.Sprintf(apiImportCodeFragment, f.Resource.ImportAlias(), f.Resource.Path))
	}

	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			if f.IsLegacyLayout {
				imports = append(imports, fmt.Sprintf(strings.ReplaceAll(controllerImportCodeFragment, "internal/controller", "controllers"), f.Repo))
			} else {
				imports = append(imports, fmt.Sprintf(controllerImportCodeFragment, f.Repo))
			}
		} else {
			if f.IsLegacyLayout {
				imports = append(imports, fmt.Sprintf(strings.ReplaceAll(multiGroupControllerImportCodeFragment, "internal/controller", "controllers"),
					f.Resource.PackageName(), f.Repo, f.Resource.Group))
			} else {
				imports = append(imports, fmt.Sprintf(multiGroupControllerImportCodeFragment,
					f.Resource.PackageName(), f.Repo, f.Resource.Group))
			}
		}
	}

	// Generate add scheme code fragments
	addScheme := make([]string, 0)
	if f.WireResource {
		addScheme = append(addScheme, fmt.Sprintf(addschemeCodeFragment, f.Resource.ImportAlias()))
	}

	// Generate setup code fragments
	setup := make([]string, 0)
	if f.WireController {
		if !f.MultiGroup || f.Resource.Group == "" {
			setup = append(setup, fmt.Sprintf(reconcilerSetupCodeFragment,
				f.Resource.Kind, f.Resource.Kind))
		} else {
			setup = append(setup, fmt.Sprintf(multiGroupReconcilerSetupCodeFragment,
				f.Resource.PackageName(), f.Resource.Kind, f.Resource.Kind))
		}
	}
	if f.WireWebhook {
		setup = append(setup, fmt.Sprintf(webhookSetupCodeFragment,
			f.Resource.ImportAlias(), f.Resource.Kind, f.Resource.Kind, f.Resource.Kind))
	}

	// Only store code fragments in the map if the slices are non-empty
	if len(imports) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), importMarker)] = imports
	}
	if len(addScheme) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), addSchemeMarker)] = addScheme
	}
	if len(setup) != 0 {
		fragments[machinery.NewMarkerFor(f.GetPath(), setupMarker)] = setup
	}

	return fragments
}

var mainTemplate = `{{ .Boilerplate }}

package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	%s
	utilcontroller "github.com/labring/operator-sdk/controller"
)

var (
	scheme = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	%s
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		concurrent           int
		rateLimiterOptions   utilcontroller.RateLimiterOptions
	)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. " +
		"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&concurrent, "concurrent", 5, "The number of concurrent cluster reconciles.")
	rateLimiterOptions.BindFlags(flag.CommandLine)
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "{{ hashFNV .Repo }}.{{ .Domain }}",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after 
		// the manager stops, so would be fine to enable this option. However, 
		// if you are doing or is intended to do any operation such as perform cleanups 
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	
	%s
	
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
`
