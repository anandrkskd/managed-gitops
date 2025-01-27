/*
Copyright 2021, 2022

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GitOpsDeploymentSpec defines the desired state of GitOpsDeployment
type GitOpsDeploymentSpec struct {
	Source ApplicationSource `json:"source"`

	// Destination is a reference to a target namespace/cluster to deploy to.
	// This field may be empty: if it is empty, it is assumed that the destination
	// is the same namespace as the GitOpsDeployment CR.
	Destination ApplicationDestination `json:"destination,omitempty"`

	// SyncPolicy controls when and how a sync will be performed.
	SyncPolicy *SyncPolicy `json:"syncPolicy,omitempty"`

	// Two possible values:
	// - Automated: whenever a new commit occurs in the GitOps repository, or the Argo CD Application is out of sync, Argo CD should be told to (re)synchronize.
	// - Manual: Argo CD should never be told to resynchronize. Instead, synchronize operations will be triggered via GitOpsDeploymentSyncRun operations only.
	// - See `GitOpsDeploymentSpecType*`
	//
	// Note: This is somewhat of a placeholder for more advanced logic that can be implemented in the future.
	// For an example of this type of logic, see the 'syncPolicy' field of Argo CD Application.
	Type string `json:"type"`
}

// ApplicationSource contains all required information about the source of an application
type ApplicationSource struct {
	// RepoURL is the URL to the repository (Git or Helm) that contains the application manifests
	RepoURL string `json:"repoURL"`
	// Path is a directory path within the Git repository, and is only valid for applications sourced from Git.
	Path string `json:"path"`
	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD.
	// In case of Helm, this is a semver tag for the Chart's version.
	TargetRevision string `json:"targetRevision,omitempty"`
}

// ApplicationDestination holds information about the application's destination
type ApplicationDestination struct {
	Environment string `json:"environment,omitempty"`

	// The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace
	Namespace string `json:"namespace,omitempty"`
}

const (
	// DeletionFinalizer will indicate the GitOpsDeployment to wait until all its dependencies are removed.
	// In the absence of this finalizer, GitOpsDeployment will be deleted first and its dependencies will be removed in the background.
	DeletionFinalizer string = "resources-finalizer.managed-gitops.redhat.com"
)

type SyncOption string

// Supported values for SyncOptions
const (
	SyncOptions_CreateNamespace_true  SyncOption = "CreateNamespace=true"
	SyncOptions_CreateNamespace_false SyncOption = "CreateNamespace=false"
)

type SyncPolicy struct {
	// Options allow you to specify whole app sync-options.
	// This option may be empty, if and when it is empty it is considered that there are no SyncOptions present.
	SyncOptions SyncOptions `json:"syncOptions,omitempty"`
}
type SyncOptions []SyncOption

const (
	GitOpsDeploymentSpecType_Automated = "automated"
	GitOpsDeploymentSpecType_Manual    = "manual"
)

func SyncOptionToStringSlice(syncOptions SyncOptions) []string {
	if syncOptions == nil {
		return nil
	}
	var res []string
	if len(syncOptions) > 0 {
		res = []string{}
		for _, syncOption := range syncOptions {
			res = append(res, string(syncOption))
		}
	}

	return res
}

// ResourceStatus holds the current sync and health status of a resource
type ResourceStatus struct {
	Group     string         `json:"group,omitempty"`
	Version   string         `json:"version,omitempty"`
	Kind      string         `json:"kind,omitempty"`
	Namespace string         `json:"namespace,omitempty"`
	Name      string         `json:"name,omitempty"`
	Status    SyncStatusCode `json:"status,omitempty"`
	Health    *HealthStatus  `json:"health,omitempty"`
}

// ReconciledState contains lists last thing ArgoCD Controller reconciled
type ReconciledState struct {
	Source      GitOpsDeploymentSource      `json:"source"`
	Destination GitOpsDeploymentDestination `json:"destination"`
}

// GitOpsDeploymentSource contains the information of .status.Sync.CompareTo.Source field of ArgoCD Application
type GitOpsDeploymentSource struct {
	// Path contains path from .status.Sync.CompareTo field of ArgoCD Application
	Path    string `json:"path"`
	RepoURL string `json:"repoURL"`
	Branch  string `json:"branch"`
}

// GitOpsDeploymentDestination contains the information of .status.Sync.CompareTo.Destination field of ArgoCD Application
type GitOpsDeploymentDestination struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// GitOpsDeploymentStatus defines the observed state of GitOpsDeployment
type GitOpsDeploymentStatus struct {
	Conditions []GitOpsDeploymentCondition `json:"conditions,omitempty"`
	Sync       SyncStatus                  `json:"sync,omitempty"`
	// Health contains information about the application's current health status
	Health HealthStatus `json:"health,omitempty"`

	// List of Resource created by a deployment
	Resources []ResourceStatus `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`

	// ReconciledState lists last deployment of ArgoCD Application
	ReconciledState ReconciledState `json:"reconciledState"`
}

// HealthStatus contains information about the currently observed health state of an application or resource
type HealthStatus struct {
	// Status holds the status code of the application or resource
	Status HealthStatusCode `json:"status,omitempty"`
	// Message is a human-readable informational message describing the health status
	Message string `json:"message,omitempty"`
}

type HealthStatusCode string

const (
	HeathStatusCodeHealthy     HealthStatusCode = "Healthy"
	HeathStatusCodeProgressing HealthStatusCode = "Progressing"
	HeathStatusCodeDegraded    HealthStatusCode = "Degraded"
	HeathStatusCodeSuspended   HealthStatusCode = "Suspended"
	HeathStatusCodeMissing     HealthStatusCode = "Missing"
	HeathStatusCodeUnknown     HealthStatusCode = "Unknown"
)

// SyncStatus contains information about the currently observed live and desired states of an application
type SyncStatus struct {
	// Status is the sync state of the comparison
	Status SyncStatusCode `json:"status"`
	// Revision contains information about the revision the comparison has been performed to
	Revision string `json:"revision,omitempty"`
}

// SyncStatusCode is a type which represents possible comparison results
type SyncStatusCode string

// Possible comparison results
const (
	// SyncStatusCodeUnknown indicates that the status of a sync could not be reliably determined
	SyncStatusCodeUnknown SyncStatusCode = "Unknown"
	// SyncStatusCodeOutOfSync indicates that desired and live states match
	SyncStatusCodeSynced SyncStatusCode = "Synced"
	// SyncStatusCodeOutOfSync indicates that there is a drift beween desired and live states
	SyncStatusCodeOutOfSync SyncStatusCode = "OutOfSync"
)

// GitOpsDeploymentCondition contains details about an GitOpsDeployment condition, which is usually an error or warning
type GitOpsDeploymentCondition struct {
	// Type is a GitOpsDeployment condition type
	Type GitOpsDeploymentConditionType `json:"type"`

	// Message contains human-readable message indicating details about the last condition.
	// +optional
	Message string `json:"message"`

	// LastProbeTime is the last time the condition was observed.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`

	// LastTransitionTime is the last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Status is the status of the condition.
	Status GitOpsConditionStatus `json:"status"`

	// Reason is a unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason GitOpsDeploymentReasonType `json:"reason"`
}

// GitOpsDeploymentConditionType represents type of GitOpsDeployment condition.
type GitOpsDeploymentConditionType string

const (
	GitOpsDeploymentConditionSyncError     GitOpsDeploymentConditionType = "SyncError"
	GitOpsDeploymentConditionErrorOccurred GitOpsDeploymentConditionType = "ErrorOccurred"
)

// GitOpsConditionStatus is a type which represents possible comparison results
type GitOpsConditionStatus string

// GitOpsDeployment Condition Status
const (
	// GitOpsConditionStatusTrue indicates that a condition type is true
	GitOpsConditionStatusTrue GitOpsConditionStatus = "True"
	// GitOpsConditionStatusFalse indicates that a condition type is false
	GitOpsConditionStatusFalse GitOpsConditionStatus = "False"
	// GitOpsConditionStatusUnknown indicates that the condition status could not be reliably determined
	GitOpsConditionStatusUnknown GitOpsConditionStatus = "Unknown"
)

type GitOpsDeploymentReasonType string

const (
	GitopsDeploymentReasonSyncError     GitOpsDeploymentReasonType = "SyncError"
	GitopsDeploymentReasonErrorOccurred GitOpsDeploymentReasonType = "ErrorOccurred"
)

const (
	GitOpsDeploymentUserError_InvalidPathSlash = "spec.source.path cannot be '/'"
	GitOpsDeploymentUserError_PathIsRequired   = "spec.source.path is a required field and it cannot be empty"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Sync Status",type=string,JSONPath=`.status.sync.status`
// +kubebuilder:printcolumn:name="Health Status",type=string,JSONPath=`.status.health.status`

// GitOpsDeployment is the Schema for the gitopsdeployments API
type GitOpsDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsDeploymentSpec   `json:"spec,omitempty"`
	Status GitOpsDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitOpsDeploymentList contains a list of GitOpsDeployment
type GitOpsDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOpsDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOpsDeployment{}, &GitOpsDeploymentList{})
}
