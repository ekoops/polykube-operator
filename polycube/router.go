package polycube

import (
	"context"
	"errors"
	"fmt"
	"github.com/ekoops/polykube-operator/polycube/clients/router"
	"net"
	"net/url"
)

// SetRouterToIntK8sLbrpPortMAC set the MAC of the polycube router interface representing the default gateway for
// the pods subnet.
func SetRouterToIntK8sLbrpPortMAC(MAC net.HardwareAddr) error {
	rName := conf.rName
	rToIklPortName := conf.rToIklPortName
	MACStr := MAC.String()
	log := log.WithValues("name", rName, "port", rToIklPortName, "MAC", MACStr)

	resp, err := routerAPI.UpdateRouterPortsMacByID(context.TODO(), rName, rToIklPortName, MACStr)
	if err != nil {
		log.Error(err, "failed to set router port MAC", "response", fmt.Sprintf("%+v", resp))
		return errors.New("failed to set router port MAC")
	}

	log.V(1).Info("set router to internal k8s lbrp port MAC")
	return nil
}

// GetRouterToIntK8sLbrpPortMAC returns the MAC of the polycube router interface representing the default gateway for
// the pods subnet.
func GetRouterToIntK8sLbrpPortMAC() (net.HardwareAddr, error) {
	rName := conf.rName
	rToIklPortName := conf.rToIklPortName
	log := log.WithValues("name", rName, "port", rToIklPortName)

	MACStr, resp, err := routerAPI.ReadRouterPortsMacByID(context.TODO(), rName, rToIklPortName)
	if err != nil {
		log.Error(err, "failed to retrieve router port MAC", "response", fmt.Sprintf("%+v", resp))
		return nil, errors.New("failed to retrieve router port MAC")
	}

	MAC, err := net.ParseMAC(MACStr)
	if err != nil {
		log.Error(err, "failed to parse router port MAC")
		return nil, errors.New("failed to parse router port MAC")
	}
	log.V(1).Info("retrieved router to internal k8s lbrp port MAC", "MAC", MAC.String())
	return MAC, nil
}

// CheckRouterRouteExistence allows to verify the existence of the route towards nodePodCIDR through nodeVtepIP on
// the polycube router.
func CheckRouterRouteExistence(nodePodCIDR *net.IPNet, nodeVtepIP net.IP) (bool, error) {
	rName := conf.rName
	network := nodePodCIDR.String()
	nexthop := nodeVtepIP.String()
	log := log.WithValues("router", rName, "network", network, "nexthop", nexthop)

	routes, resp, err := routerAPI.ReadRouterRouteListByID(context.TODO(), rName)
	if err != nil {
		log.Error(err, "failed to retrieve router routes list", "response", fmt.Sprintf("%+v", resp))
		return false, errors.New("failed to retrieve router routes list")
	}

	for _, r := range routes {
		if r.Network == network && r.Nexthop == nexthop {
			log.V(1).Info("verified router route existence")
			return true, nil
		}
	}
	log.V(1).Info("router route not found")
	return false, nil
}

// CreateRouterRoute creates a new route on the polycube router.
func CreateRouterRoute(nodePodCIDR *net.IPNet, nodeVtepIP net.IP) error {
	rName := conf.rName
	network := nodePodCIDR.String()
	nexthop := nodeVtepIP.String()
	log := log.WithValues("router", rName, "network", network, "nexthop", nexthop)

	route := router.Route{
		Network:    network,
		Nexthop:    nexthop,
		Interface_: conf.rToVxlanPortName,
	}

	// adding route to router in order to make node pod CIDR reachable throw vxlan interface
	if resp, err := routerAPI.CreateRouterRouteByID(
		context.TODO(), rName, url.QueryEscape(route.Network), route.Nexthop, route,
	); err != nil && resp.StatusCode != 409 {
		log.Error(err, "failed to create router route", "response", fmt.Sprintf("%+v", resp))
		return errors.New("failed to create router route")
	}
	log.V(1).Info("created router route")
	return nil
}

// deleteRouterRoute deletes the route on the polycube router.
func deleteRouterRoute(network, nexthop string) error {
	rName := conf.rName
	log := log.WithValues("router", rName, "network", network, "nexthop", nexthop)
	// delete route to router in order to make node pod CIDR unreachable throw vxlan interface
	if resp, err := routerAPI.DeleteRouterRouteByID(
		context.TODO(), rName, url.QueryEscape(network), nexthop,
	); err != nil && resp.StatusCode != 409 {
		log.Error(err, "failed to delete router route", "response", fmt.Sprintf("%+v", resp))
		return errors.New("failed to delete router route")
	}
	log.V(1).Info("deleted router route")
	return nil
}

// DeleteRouterRoute deletes the route on the polycube router in order to disable connectivity toward
// the node Pod CIDR.
func DeleteRouterRoute(nodePodCIDR *net.IPNet, nodeVtepIP net.IP) error {
	return deleteRouterRoute(nodePodCIDR.String(), nodeVtepIP.String())
}

// CleanupRouterRoutes deletes all the non-local routes of the polycube router.
func CleanupRouterRoutes() error {
	rName := conf.rName
	log := log.WithValues("router", rName)
	routes, resp, err := routerAPI.ReadRouterRouteListByID(context.TODO(), rName)
	if err != nil {
		log.Error(err, "failed to retrieve router routes", "response", fmt.Sprintf("%+v", resp))
		return errors.New("failed to retrieve router routes")
	}
	for _, route := range routes {
		if route.Nexthop != "local" {
			if err := deleteRouterRoute(route.Network, route.Nexthop); err != nil {
				return err
			}
		}
	}
	log.V(1).Info("deleted all non-local router routes")
	return nil
}
