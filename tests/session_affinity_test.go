package tests

import (
	"context"
	"github.com/k8sbykeshed/k8s-service-validator/entities/kubernetes"
	"testing"

	"github.com/k8sbykeshed/k8s-service-validator/entities"
	"github.com/k8sbykeshed/k8s-service-validator/matrix"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSessionAffinity(t *testing.T) {


	testSessionAffinity := features.New("SessionAffinity").WithLabel("type", "sessionAffinity").
		Setup(func(context.Context, *testing.T, *envconf.Config) context.Context {
			service := entities.NewServiceFromTemplate(entities.Service{
															Name: "ServiceWithSessionAffinity",
															Namespace: namespace,
															Selector: {},
														})



			nodes, err := ma.GetReadyNodes()
			if err != nil {
				t.Fatal(err)
			}

			newPod, err = createHostNetworkPod("pod-5", nodes[0])
			if err != nil {
				t.Fatal(err)
			}
			model.AddPod(newPod, namespace)

			return ctx
		}).
		Teardown(func(context.Context, *testing.T, *envconf.Config) context.Context {
			logger.Info("delete newly created pod, which use host network.")
			if err := ma.DeletePod(newPod.Name, newPod.Namespace); err != nil {
				t.Fatal(err)
			}
			err := model.RemovePod(newPod.Name, namespace)
			if err != nil {
				ma.Logger.Debug(err.Error())
			}
			return ctx
		}).
		Assess("should function for pods using hostNetwork", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ma.Logger.Info("testing pod with hostNetwork connections.")
			// Expect pod-5 can connect with pods in the cluster
			reachability := matrix.NewReachability(model.AllPods(), true)

			testCase := matrix.TestCase{ToPort: 80, Protocol: v1.ProtocolTCP, Reachability: reachability, ServiceType: entities.PodIP}
			wrong := matrix.ValidateOrFail(ma, model, &testCase, false)
			if wrong > 0 {
				t.Error("Wrong result number ")
			}
			return ctx
		}).Feature()

	testenv.Test(t, testSessionAffinity)
}