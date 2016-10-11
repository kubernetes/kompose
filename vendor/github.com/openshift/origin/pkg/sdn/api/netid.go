package api

// Accessor methods to annotate NetNamespace for multitenant support
import (
	"fmt"
	"strings"
)

type PodNetworkAction string

const (
	// Maximum VXLAN Virtual Network Identifier(VNID) as per RFC#7348
	MaxVNID = uint32((1 << 24) - 1)
	// VNID: 1 to 9 are internally reserved for any special cases in the future
	MinVNID = uint32(10)
	// VNID: 0 reserved for default namespace and can reach any network in the cluster
	GlobalVNID = uint32(0)

	// ChangePodNetworkAnnotation is an annotation on NetNamespace to request change of pod network
	ChangePodNetworkAnnotation string = "pod.network.openshift.io/multitenant.change-network"

	// Acceptable values for ChangePodNetworkAnnotation
	GlobalPodNetwork  PodNetworkAction = "global"
	JoinPodNetwork    PodNetworkAction = "join"
	IsolatePodNetwork PodNetworkAction = "isolate"
)

var (
	ErrorPodNetworkAnnotationNotFound = fmt.Errorf("ChangePodNetworkAnnotation not found")
)

// Check if the given vnid is valid or not
func ValidVNID(vnid uint32) error {
	if vnid == GlobalVNID {
		return nil
	}
	if vnid < MinVNID {
		return fmt.Errorf("VNID must be greater than or equal to %d", MinVNID)
	}
	if vnid > MaxVNID {
		return fmt.Errorf("VNID must be less than or equal to %d", MaxVNID)
	}
	return nil
}

// GetChangePodNetworkAnnotation fetches network change intent from NetNamespace
func GetChangePodNetworkAnnotation(netns *NetNamespace) (PodNetworkAction, string, error) {
	value, ok := netns.Annotations[ChangePodNetworkAnnotation]
	if !ok {
		return PodNetworkAction(""), "", ErrorPodNetworkAnnotationNotFound
	}

	args := strings.Split(value, ":")
	switch PodNetworkAction(args[0]) {
	case GlobalPodNetwork:
		return GlobalPodNetwork, "", nil
	case JoinPodNetwork:
		if len(args) != 2 {
			return PodNetworkAction(""), "", fmt.Errorf("invalid namespace for join pod network: %s", value)
		}
		namespace := args[1]
		return JoinPodNetwork, namespace, nil
	case IsolatePodNetwork:
		return IsolatePodNetwork, "", nil
	}

	return PodNetworkAction(""), "", fmt.Errorf("invalid ChangePodNetworkAnnotation: %s", value)
}

// SetChangePodNetworkAnnotation sets network change intent on NetNamespace
func SetChangePodNetworkAnnotation(netns *NetNamespace, action PodNetworkAction, params string) {
	if netns.Annotations == nil {
		netns.Annotations = make(map[string]string)
	}

	value := string(action)
	if len(params) != 0 {
		value = fmt.Sprintf("%s:%s", value, params)
	}
	netns.Annotations[ChangePodNetworkAnnotation] = value
}

// DeleteChangePodNetworkAnnotation removes network change intent from NetNamespace
func DeleteChangePodNetworkAnnotation(netns *NetNamespace) {
	delete(netns.Annotations, ChangePodNetworkAnnotation)
}
