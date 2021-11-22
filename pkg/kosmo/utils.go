package kosmo

import (
	"fmt"
	"io"
	"os"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	DeploymentGVK = schema.GroupVersionKind{
		Group:   appsv1.GroupName,
		Version: "v1",
		Kind:    "Deployment",
	}
	ServiceGVK = schema.GroupVersionKind{
		Group:   corev1.GroupName,
		Version: "v1",
		Kind:    "Service",
	}
	IngressGVK = schema.GroupVersionKind{
		Group:   netv1.GroupName,
		Version: "v1",
		Kind:    "Ingress",
	}
)

type Comparable interface {
	GetManagedFields() []metav1.ManagedFieldsEntry
	SetManagedFields(managedFields []metav1.ManagedFieldsEntry)
	SetResourceVersion(resourceVersion string)
}

func resetManagedFieldTime(obj Comparable) {
	mf := obj.GetManagedFields()
	for i := range mf {
		mf[i].Time = nil
	}
	obj.SetManagedFields(mf)
}

func resetResourceVersion(obj Comparable) {
	obj.SetResourceVersion("")
}

type DeepEqualOption interface {
	Apply(x, y Comparable)
}

type printDiff struct {
	out io.Writer
}

func (o printDiff) Apply(x, y Comparable) {
	clog.PrintObjectDiff(o.out, x, y)
}

func WithPrintDiff() DeepEqualOption {
	return printDiff{out: os.Stderr}
}

// LooseDeepEqual deep equal objects without dynamic values
// This function removes some fields, so you should give deep-copied objects.
func LooseDeepEqual(x, y Comparable, opts ...DeepEqualOption) bool {
	resetManagedFieldTime(x)
	resetManagedFieldTime(y)

	resetResourceVersion(x)
	resetResourceVersion(y)

	for _, o := range opts {
		o.Apply(x, y)
	}

	return equality.Semantic.DeepEqual(x, y)
}

func PodStatusReason(pod corev1.Pod) string {
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	if reason == "Running" && readyContainers != totalContainers {
		reason = fmt.Sprintf("Running:%d/%d", readyContainers, totalContainers)
	}

	// if pod.DeletionTimestamp != nil && pod.Status.Reason == node.NodeUnreachablePodReason {
	// 	reason = "Unknown"
	// } else
	if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	return reason
}
