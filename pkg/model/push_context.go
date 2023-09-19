package model

import (
	"sort"
	"strings"
	"time"

	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/features"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/constants"
)

type DestinationType string

const (
	Single DestinationType = "Single"

	Multiple DestinationType = "Multiple"
)

type BackendService struct {
	Namespace string
	Name      string
	Port      uint32
	Weight    int32
}

type IngressRoute struct {
	Name            string
	Host            string
	PathType        string
	Path            string
	DestinationType DestinationType
	// Deprecated
	ServiceName string
	ServiceList []BackendService
	Error       string
}

type IngressRouteCollection struct {
	Valid   []IngressRoute
	Invalid []IngressRoute
}

type IngressDomain struct {
	Host string

	// tls for HTTPS
	// default HTTP
	Protocol string

	// cluster id/namespace/name
	SecretName string

	// creation time of ingress resource
	CreationTime time.Time
	Error        string
}

type IngressDomainCollection struct {
	Valid   []IngressDomain
	Invalid []IngressDomain
}

func SortStableForIngressRoutes(routes []IngressRoute) {
	isAllMatch := func(route IngressRoute) bool {
		return route.PathType == "prefix" && route.Path == "/"
	}

	sort.SliceStable(routes, func(i, j int) bool {
		if routes[i].Host != routes[j].Host {
			return len(routes[i].Host) > len(routes[j].Host)
		}

		if isAllMatch(routes[i]) {
			return false
		}
		if isAllMatch(routes[j]) {
			return true
		}

		if routes[i].PathType == routes[j].PathType {
			// sort canary
			if routes[i].Path == routes[j].Path {
				return strings.HasSuffix(routes[i].Name, "canary")
			}

			return len(routes[i].Path) > len(routes[j].Path)
		}

		if routes[i].PathType == "exact" {
			return true
		}

		if routes[i].PathType != "exact" &&
			routes[j].PathType != "exact" {
			return routes[i].PathType == "prefix"
		}

		return false
	})
}

// IngressStore provide uniform access to get convert resources from ingress for mse ops via debug interface.
type IngressStore interface {
	GetIngressRoutes() IngressRouteCollection

	GetIngressDomains() IngressDomainCollection
}

func createCRName(clusterId, autoGenerated string) string {
	stripName := strings.TrimPrefix(autoGenerated, constants.IstioIngressGatewayName+"-")
	if clusterId != "" && clusterId != "Kubernetes" {
		return clusterId + "-" + stripName
	}
	return stripName
}

// virtualServiceFilter will modify copied configs from underlying store.
// We merge routes into pre host of virtual service.
func virtualServiceFilter(configs []config.Config) []config.Config {
	var autoGenerated []*config.Config
	configsForName := make(map[string]*config.Config, len(configs))

	istioClusterId := features.ClusterName

	for idx := range configs {
		c := configs[idx]
		if strings.HasPrefix(c.Name, constants.IstioIngressGatewayName) {
			autoGenerated = append(autoGenerated, &c)
		} else {
			configsForName[c.Name] = &c
		}
	}

	log.Infof("Auto generator virtual services number %d", len(autoGenerated))

	for _, c := range autoGenerated {
		targetName := createCRName(istioClusterId, c.Name)
		rawVS, exist := configsForName[targetName]
		if exist {
			vs := rawVS.Spec.(*networking.VirtualService)
			autoGeneratedVS := c.Spec.(*networking.VirtualService)
			// TODO: Upgrade fix
			//if len(vs.HostHTTPFilters) == 0 {
			//	vs.HostHTTPFilters = autoGeneratedVS.HostHTTPFilters
			//}
			// TODO(special.fy) make configurable for priority of routes between OPS, ACK and ASM
			vs.Http = append(vs.Http, autoGeneratedVS.Http...)
		} else {
			// We change the auto-generated config name to the format of cr name same with ops when ops
			// don't have this host.
			c.Name = targetName
			configsForName[targetName] = c
		}
	}

	var out []config.Config
	for _, c := range configsForName {
		out = append(out, *c)
	}
	return out
}

// destinationFilter will modify copied configs from underlying store.
func destinationFilter(configs []config.Config) []config.Config {
	var autoGenerated []*config.Config
	configsForName := make(map[string]*config.Config, len(configs))

	for idx := range configs {
		c := configs[idx]
		if strings.HasPrefix(c.Name, constants.IstioIngressGatewayName) {
			autoGenerated = append(autoGenerated, &c)
		} else {
			configsForName[c.Name] = &c
		}
	}

	log.Infof("Auto generator destination rule number %d", len(autoGenerated))

	for _, c := range autoGenerated {
		// DestinationRule name of ops is md5 without cluster id.
		targetName := strings.TrimPrefix(c.Name, constants.IstioIngressGatewayName+"-")
		_, exist := configsForName[targetName]
		if !exist {
			// We change the auto-generated config name to the format of cr name same with ops when ops
			// don't have destination rule for this service.
			c.Name = targetName
			configsForName[targetName] = c
		}
	}

	var out []config.Config
	for _, c := range configsForName {
		out = append(out, *c)
	}
	return out
}

// gatewayFilter will modify copied configs from underlying store.
// We merge routes into pre host of virtual service.
func gatewayFilter(configs []config.Config) []config.Config {
	var autoGenerated []*config.Config
	configsForName := make(map[string]*config.Config, len(configs))

	istioClusterId := features.ClusterName

	for idx := range configs {
		c := configs[idx]
		if strings.HasPrefix(c.Name, constants.IstioIngressGatewayName) {
			autoGenerated = append(autoGenerated, &c)
		} else {
			configsForName[c.Name] = &c
		}
	}

	log.Infof("Auto generator gateways number %d", len(autoGenerated))

	for _, c := range autoGenerated {
		targetName := createCRName(istioClusterId, c.Name)
		_, exist := configsForName[targetName]
		// Note, if ops already has the host without tls and ingress has the same host with tls,
		// we don't merge tls settings, i.e, we don't adopt ingress tls for this host.
		if !exist {
			// We change the auto-generated config name to the format of cr name same with ops when ops
			// don't have this host.
			c.Name = targetName
			configsForName[targetName] = c
		}
	}

	var out []config.Config
	for _, c := range configsForName {
		out = append(out, *c)
	}
	return out
}

// TODO: Upgrade fix
//func GetGatewayByName(ps *istiomodel.PushContext, name string) *config.Config {
//	parts := strings.Split(name, "/")
//	if len(parts) != 2 {
//		return nil
//	}
//
//	for _, cfg := range ps.gatewayIndex.all {
//		if cfg.Namespace == parts[0] && cfg.Name == parts[1] {
//			return &cfg
//		}
//	}
//
//	return nil
//}
//
//func GetHTTPFiltersFromEnvoyFilter(ps *istiomodel.PushContext, node *istiomodel.Proxy) []*httpConn.HttpFilter {
//	var out []*httpConn.HttpFilter
//	envoyFilterWrapper := ps.EnvoyFilters(node)
//	if envoyFilterWrapper != nil && len(envoyFilterWrapper.Patches) > 0 {
//		httpFilters := envoyFilterWrapper.Patches[networking.EnvoyFilter_HTTP_FILTER]
//		if len(httpFilters) > 0 {
//			for _, filter := range httpFilters {
//				if filter.Operation == networking.EnvoyFilter_Patch_INSERT_AFTER ||
//					filter.Operation == networking.EnvoyFilter_Patch_ADD ||
//					filter.Operation == networking.EnvoyFilter_Patch_INSERT_BEFORE ||
//					filter.Operation == networking.EnvoyFilter_Patch_INSERT_FIRST {
//					out = append(out, proto.Clone(filter.Value).(*httpConn.HttpFilter))
//				}
//			}
//		}
//	}
//
//	return out
//}
