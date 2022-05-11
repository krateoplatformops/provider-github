package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "github.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Firewall type metadata.
var (
	RepoKind             = reflect.TypeOf(Repo{}).Name()
	RepoGroupKind        = schema.GroupKind{Group: Group, Kind: RepoKind}.String()
	RepoKindAPIVersion   = RepoKind + "." + SchemeGroupVersion.String()
	RepoGroupVersionKind = SchemeGroupVersion.WithKind(RepoKind)
)

func init() {
	SchemeBuilder.Register(&Repo{}, &RepoList{})
}
