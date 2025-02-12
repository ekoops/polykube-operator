package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
)

// GetPods returns the list of all the pods contained in the specified cluster node.
// The request is performed using directly a kubernetes clientset (without using caching mechanisms from
// the controller-runtime library
func GetPods(nodeName string) (*v1.PodList, error) {
	log := log.WithValues("node", nodeName)
	opts := metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	}
	pods, err := Conf.clientset.CoreV1().Pods("").List(context.TODO(), opts)
	if err != nil {
		log.Error(err, "failed to retrieve cluster node pods")
		return nil, fmt.Errorf("failed to retrieve cluster node pods")
	}
	log.V(1).Info("retrieved cluster node pods")
	return pods, nil
}

// ParsePodCIDR returns the pod CIDR of the provided node
func ParsePodCIDR(n *v1.Node) (*net.IPNet, error) {
	rawPodCIDR := n.Spec.PodCIDR
	log := log.WithValues("node", n.Name, "rawPodCIDR", rawPodCIDR)

	// parsing pod CIDR
	_, podCIDR, err := net.ParseCIDR(rawPodCIDR)
	if err != nil {
		log.Error(err, "failed to parse cluster node pod CIDR")
		return nil, errors.New("failed to parse cluster node pod CIDR")
	}

	// making sure that the pods CIDR is IPv4
	podCIDR.IP = podCIDR.IP.To4()
	if podCIDR.IP == nil {
		log.Error(errors.New("unsupported IPv6 pod CIDR"), "failed to parse cluster node pod CIDR")
		return nil, errors.New("failed to parse cluster node pod CIDR")
	}

	log.V(1).Info("parsed cluster node pod CIDR", "podCIDR", podCIDR)
	return podCIDR, nil
}

// CalcPodsDefaultGatewayIPNet calculates the pods default gateway IP and prefix length starting from the pod CIDR.
// The convention that the IP of the pods default gateway is the last IP of pod CIDR other than the broadcast address
// (e.g.: if the pod CIDR is /24, then the default gateway IP will be .254) is used.
func CalcPodsDefaultGatewayIPNet(podCIDR *net.IPNet) (*net.IPNet, error) {
	// calculating the broadcast address
	subIP := podCIDR.IP
	subMask := podCIDR.Mask
	subBroadcastIP := net.IP(make([]byte, 4))
	for i := range subIP {
		subBroadcastIP[i] = subIP[i] | ^subMask[i]
	}

	// using the address preceding the broadcast address as default gateway for pods
	gwIP := ip.PrevIP(subBroadcastIP)
	gwIPNet := &net.IPNet{
		IP:   gwIP,
		Mask: subMask,
	}

	log.V(1).Info(
		"calculated pods default gateway address from the Pod CIDR", "podCIDR", podCIDR, "IP", gwIPNet,
	)
	return gwIPNet, nil
}
