package model

import (
	"strings"
	"testing"

	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/constants"
)

func createName(domain string) string {
	domain = strings.ReplaceAll(domain, ".", "-")
	return features.ClusterName + "-" + domain
}

func createAutoGeneratedName(domain string) string {
	domain = strings.ReplaceAll(domain, ".", "-")
	return constants.IstioIngressGatewayName + "-" + domain
}

func TestVirtualServiceFilter(t *testing.T) {
	features.ClusterName = "gw-123-istio"

	inputConfigs := []config.Config{
		{
			Meta: config.Meta{
				Name:      createName("test.com"),
				Namespace: "test",
			},
			Spec: &networking.VirtualService{
				Http: []*networking.HTTPRoute{
					{
						Name: "route-1",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("test.com"),
				Namespace: "test",
			},
			Spec: &networking.VirtualService{
				Http: []*networking.HTTPRoute{
					{
						Name: "route-2",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("foo.com"),
				Namespace: "test",
			},
			Spec: &networking.VirtualService{
				Http: []*networking.HTTPRoute{
					{
						Name: "route-1",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("bar.com"),
				Namespace: "test",
			},
			Spec: &networking.VirtualService{
				Http: []*networking.HTTPRoute{
					{
						Name: "route-1",
					},
				},
			},
		},
	}

	out := virtualServiceFilter(inputConfigs)
	if len(out) != 3 {
		t.Fatal("filter error")
	}

	for _, c := range out {
		if !strings.HasPrefix(c.Name, features.ClusterName) {
			t.Fatalf("CRD name %s mush has prefix %s", c.Name, features.ClusterName)
		}
	}
}

func TestGatewayFilter(t *testing.T) {
	features.ClusterName = "gw-123-istio"

	inputConfigs := []config.Config{
		{
			Meta: config.Meta{
				Name:      createName("test.com"),
				Namespace: "test",
			},
			Spec: &networking.Gateway{
				Servers: []*networking.Server{
					{
						Name: "server-1",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("test.com"),
				Namespace: "test",
			},
			Spec: &networking.Gateway{
				Servers: []*networking.Server{
					{
						Name: "server-2",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("foo.com"),
				Namespace: "test",
			},
			Spec: &networking.Gateway{
				Servers: []*networking.Server{
					{
						Name: "server-1",
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      createAutoGeneratedName("bar.com"),
				Namespace: "test",
			},
			Spec: &networking.Gateway{
				Servers: []*networking.Server{
					{
						Name: "server-1",
					},
				},
			},
		},
	}

	out := gatewayFilter(inputConfigs)
	if len(out) != 3 {
		t.Fatal("filter error")
	}

	for _, c := range out {
		if !strings.HasPrefix(c.Name, features.ClusterName) {
			t.Fatalf("CRD name %s mush has prefix %s", c.Name, features.ClusterName)
		}
	}
}

func TestDestinationRuleFilter(t *testing.T) {
	inputConfigs := []config.Config{
		{
			Meta: config.Meta{
				Name:      "e3431b3db77d88642015e60647514d2f",
				Namespace: "test",
			},
			Spec: &networking.DestinationRule{
				Host: "test.default.svc.cluster.local",
				TrafficPolicy: &networking.TrafficPolicy{
					LoadBalancer: &networking.LoadBalancerSettings{
						LbPolicy: &networking.LoadBalancerSettings_Simple{
							Simple: networking.LoadBalancerSettings_LEAST_CONN,
						},
					},
				},
			},
		},
		{
			Meta: config.Meta{
				Name:      constants.IstioIngressGatewayName + "-e3431b3db77d88642015e60647514d2f",
				Namespace: "test",
			},
			Spec: &networking.DestinationRule{
				Host: "test.default.svc.cluster.local",
				TrafficPolicy: &networking.TrafficPolicy{
					LoadBalancer: &networking.LoadBalancerSettings{
						LbPolicy: &networking.LoadBalancerSettings_Simple{
							Simple: networking.LoadBalancerSettings_RANDOM,
						},
					},
				},
			},
		},
	}

	out := destinationFilter(inputConfigs)
	if len(out) != 1 {
		t.Fatal("filter error")
	}
}
