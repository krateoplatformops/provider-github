package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	repov1alpha1 "github.com/krateoplatformops/provider-github/apis/repo/v1alpha1"
	gitv1alpha1 "github.com/krateoplatformops/provider-github/apis/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		gitv1alpha1.SchemeBuilder.AddToScheme,
		repov1alpha1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
