package repo

import (
	"context"
	"errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	repov1alpha1 "github.com/krateoplatformops/provider-github/apis/repo/v1alpha1"
	"github.com/krateoplatformops/provider-github/pkg/clients"
	"github.com/krateoplatformops/provider-github/pkg/clients/github"
)

const (
	errNotRepo = "managed resource is not a repo custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(repov1alpha1.RepoGroupKind)

	//opts := controller.Options{
	//	RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
	//}

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(repov1alpha1.RepoGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&repov1alpha1.Repo{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*repov1alpha1.Repo)
	if !ok {
		return nil, errors.New(errNotRepo)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	return &external{
		kube:  c.kube,
		log:   c.log,
		ghCli: github.NewClient(*cfg),
		rec:   c.recorder,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube  client.Client
	log   logging.Logger
	ghCli *github.Client
	rec   record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*repov1alpha1.Repo)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRepo)
	}

	spec := cr.Spec.ForProvider.DeepCopy()

	ok, err := e.ghCli.Repos().Exists(spec)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if ok {
		e.log.Debug("Repo already exists", "org", spec.Org, "name", spec.Name)
		e.rec.Eventf(cr, corev1.EventTypeNormal, "AlredyExists", "Repo '%s/%s' already exists", spec.Org, spec.Name)

		cr.SetConditions(xpv1.Available())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	e.log.Debug("Repo does not exists", "org", spec.Org, "name", spec.Name)

	return managed.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*repov1alpha1.Repo)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRepo)
	}

	cr.SetConditions(xpv1.Creating())

	spec := cr.Spec.ForProvider.DeepCopy()

	err := e.ghCli.Repos().Create(spec)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	e.log.Debug("Repo created", "org", spec.Org, "name", spec.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "RepoCreated", "Repo '%s/%s' created", spec.Org, spec.Name)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*repov1alpha1.Repo)
	if !ok {
		return errors.New(errNotRepo)
	}

	cr.SetConditions(xpv1.Deleting())

	spec := cr.Spec.ForProvider.DeepCopy()

	err := e.ghCli.Repos().Delete(spec)
	if err != nil {
		return err
	}
	e.log.Debug("Repo deleted", "org", spec.Org, "name", spec.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "RepDeleted", "Repo '%s/%s' deleted", spec.Org, spec.Name)

	return nil
}
