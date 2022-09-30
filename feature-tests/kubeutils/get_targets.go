package kubeutils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/cp"
	"k8s.io/kubectl/pkg/cmd/util"
)

var portlookup = map[string][]string{
	"victim": {"80", "445", "139", "6667", "6697", "8067"},
	"slsw":   {"80"},
}

type Target struct {
	App          string
	ID           string
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
				ID:           pod.Name,
			}
			result = append(result, t)
			i++
		}
	}
	return result, nil
}

type PodCopy interface {
	FromPod(sourceFilePath, destinationFilePath string) error
}

type podCP struct {
	RestConfig *rest.Config
	ClientSet  *kubernetes.Clientset
	Container  string
	Pod        string
}

func NewPodCP(
	config rest.Config,
	clientset *kubernetes.Clientset,
	container, pod string) PodCopy {

	config.APIPath = "/api"
	config.GroupVersion = &schema.GroupVersion{Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}

	return &podCP{
		RestConfig: &config,
		ClientSet:  clientset,
		Container:  container,
		Pod:        pod,
	}

}

func (p *podCP) FromPod(sourceFilePath, destinationFilePath string) error {
	ioStreams, _, out, errout := genericclioptions.NewTestIOStreams()

	var copt genericclioptions.RESTClientGetter = &genericclioptions.ConfigFlags{
		WrapConfigFn: func(c *rest.Config) *rest.Config {
			return p.RestConfig
		},
	}

	nf := util.NewFactory(copt)
	cmd := cp.NewCmdCp(nf, ioStreams)

	sourceFilePath = p.Pod + ":" + sourceFilePath
	cmd.SetArgs([]string{"-c", p.Container, sourceFilePath, destinationFilePath})

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf(
			"%s; out: %s; err out: %s",
			err.Error(), out.String(), errout.String(),
		)
	}
	return nil
}
