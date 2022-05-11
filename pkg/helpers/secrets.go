package helpers

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

func SetSecret(ctx context.Context, k client.Client, ref *xpv1.SecretKeySelector, val string) error {
	if ref == nil {
		return errors.New("no credentials secret referenced")
	}

	s := &corev1.Secret{}
	s.Name = ref.Name
	s.Namespace = ref.Namespace
	s.StringData = map[string]string{
		ref.Key: val,
	}

	return k.Create(ctx, s)
}

func GetSecret(ctx context.Context, k client.Client, ref *xpv1.SecretKeySelector) (string, error) {
	if ref == nil {
		return "", errors.New("no credentials secret referenced")
	}

	s := &corev1.Secret{}
	if err := k.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s); err != nil {
		return "", errors.Wrapf(err, "cannot get %s secret", ref.Name)
	}

	return string(s.Data[ref.Key]), nil
}

func DeleteSecret(ctx context.Context, k client.Client, ref *xpv1.SecretKeySelector) error {
	if ref == nil {
		return errors.New("no credentials secret referenced")
	}

	s := &corev1.Secret{}
	s.Name = ref.Name
	s.Namespace = ref.Namespace

	return k.Delete(ctx, s)
}

/*
func ErrorIsNotFound(err error) bool {
	ex, ok := err.(*apierrors.StatusError)
	return ok || (ex.Status().Code == http.StatusNotFound)
}
*/
