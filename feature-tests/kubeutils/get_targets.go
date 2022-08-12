package kubeutils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var portlookup = map[string][]string{
	"victim": {"80", "445", "139", "6667", "6697", "8067"},
	"slsw":   {"80"},
}

type Target struct {
	App          string
	IP           string
	ExposedPorts []string
}

// RequiredFoundPorts returns the ports required to be found when running discovery
//
// Those ports can differ from the defined ports within the deployment since there may not relevant
// for testing.
func (t *Target) RequiredFoundPorts() []string {
	if r, ok := portlookup[t.App]; ok {
		return r
	}
	return nil
}

// GetPodIPsLabel get targets via clientset within namespace
func GetPodIPsLabel(clientset *kubernetes.Clientset, namespace string) ([]Target, error) {
	var result []Target
	i := 0
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	result = make([]Target, 0, len(pods.Items))

	for _, pod := range pods.Items {
		if app, ok := pod.GetLabels()["app"]; ok {

			exposedPorts := make([]string, 0, 0)
			for _, c := range pod.Spec.Containers {
				for _, p := range c.Ports {
					exposedPorts = append(exposedPorts, fmt.Sprintf("%d", p.ContainerPort))
				}
			}
			t := Target{
				App:          app,
				IP:           pod.Status.PodIP,
				ExposedPorts: exposedPorts,
			}
			result = append(result, t)
			i++
		}
	}
	return result, nil
}
