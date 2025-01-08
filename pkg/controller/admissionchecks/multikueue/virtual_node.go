package multikueue

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kueue "sigs.k8s.io/kueue/apis/kueue/v1beta1"
)

const (
	multiKueueVirtualNode = "multikueue-virtual-node"
)

type VirtualNodeController struct {
	client client.Client
}

func NewVirtualNodeReconciler(client client.Client) *VirtualNodeController {
	return &VirtualNodeController{
		client: client,
	}
}

func (c *VirtualNodeController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtual-node").
		//<TODO:> Temp fix - should be watches(MK, eventhandler)
		// Bandaiding with a For instead.
		For(&kueue.MultiKueueCluster{}).
		Complete(c)
}

// <TODO>: Add event handlers, where if there is an delete event to the
// virtual node, we look for pods scheduled on it, evict it and delete the node
// gracefully. Similarly, for update, make sure it is blocked for specific fields.

// +kubebuilder:rbac:groups="",resources=events,verbs=create;watch;update
// +kubebuilder:rbac:groups=kueue.x-k8s.io,resources=multikueueclusters,verbs=get;list;watch
// +kubebuilder:rbac:groups=kueue.x-k8s.io,resources=multikueueclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;update;patch

func (c *VirtualNodeController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Always ensure that the virtual node exists.
	// <TODO>: A separate controller is probably an overkill for
	// a single Virtual Node.
	// Raw thoughts:
	// 1. Using this so that the reconciler
	// gets triggered whenever a MultiKueueCluster object is created.
	// 2. If virtual Node is deleted it should get recreated.
	// 3. Ideally a reconciler for a VirtualNode, should get triggered
	// on a Virtual Node CRUD operation.
	// 4. Since this is a virtual Node and its lifecycle is going to be
	// maintained by us - is there anything else this reconciler needs
	// to implement from K8s perspective.
	// 5. Can this be a part of multiKueueCluster controller itself?
	// 6. Should probably create a separate API that encapsulates corev1.Node
	// and also has any helpers we need so that its cleaner.
	log := ctrl.LoggerFrom(ctx)
	log.V(2).Info("Reconciling Virtual Node")

	virtualNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: multiKueueVirtualNode,
		},
	}
	err := c.client.Get(ctx, req.NamespacedName, virtualNode)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			// If it's anything other than not found, log it.
			log.Error(err, "Virtual node not found")
			return ctrl.Result{}, err
		}

		// If not found, create one.
		// <TODO>: Verify if there is a better way to create a
		// virtual node. For now, manually setting the ready status
		// to true. Ideally a controller shouldn't be setting the status.
		virtualNode = &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: multiKueueVirtualNode,
				// <TODO> - better name? Visit later!
				Labels: map[string]string{
					"multikueue.kubernetes.io": "true",
				},
			},
			Spec: corev1.NodeSpec{
				Unschedulable: false,
				Taints:        []corev1.Taint{},
			},
			Status: corev1.NodeStatus{
				Phase: corev1.NodeRunning,
				Conditions: []corev1.NodeCondition{
					{
						Type:               corev1.NodeReady,
						Status:             corev1.ConditionTrue,
						Reason:             multiKueueVirtualNode,
						Message:            "Virtual node workaround",
						LastHeartbeatTime:  metav1.Now(), // Had to set this manually so that kubescheduler doesn't modify it.
						LastTransitionTime: metav1.Now(),
					},
				},
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000"), // Random nos, as they seem to be required.
					corev1.ResourceMemory: resource.MustParse("2000"),
				},
				Allocatable: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000"),
					corev1.ResourceMemory: resource.MustParse("2000"),
				},
			},
		}
	}

	// <TODO:> Implement logic where node is available, but status is not
	// as expected, and needs to be set to ready. If required we may need to delete
	// and recreate it.
	return ctrl.Result{}, c.client.Create(ctx, virtualNode)
}
