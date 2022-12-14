/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/greenbone/ospd-openvas/smoketest/connection"
	"github.com/greenbone/scanner-lab/feature-tests/featuretest"
	"github.com/greenbone/scanner-lab/feature-tests/featuretest/findservice"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
	"github.com/greenbone/scanner-lab/feature-tests/sink"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func findFirstOpenVAS(pods []kubeutils.Target) (*kubeutils.Target, error) {
	for _, p := range pods {
		if p.App == "openvas" {
			return &p, nil
		}
	}
	return nil, errors.New("no openvas pod found")
}

func main() {
	vtDIR := flag.String("vt-dir", "/var/lib/openvas/plugins", "(optional) a path to existing plugins.")
	policyPath := flag.String("policy-path", "/var/lib/gvm/data-objects/gvmd/22.04/scan-configs", "(optional) path to policies.")
	certPath := flag.String("cert-path", "", "(optional) path to the certificate used by ospd.")
	certKeyPath := flag.String("certkey-path", "", "(optional) path to certificate key used by ospd.")
	mattermostChannelID := flag.String("mattermost-channel-id", "wsgmdikbjiyn8m5ifa5njqngwr", "(optional) a channel id to send mattermost messages to")
	mattermostToken := flag.String("mattermost-token", "", "password for mattermost user; can also be env variable: MATTERMOST_TOKEN")
	mattermostAddress := flag.String("mattermost-address", "https://mattermost.greenbone.net", "The address of the mattermost server.")
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()
	if *mattermostToken == "" {
		if mt, ok := os.LookupEnv("MATTERMOST_TOKEN"); ok {
			*mattermostToken = mt
		} else {
			fmt.Printf("no MATTERMOST_TOKEN is set. Results will be printed into stdout only.\n")
		}
	}

	var config *rest.Config
	if f, err := os.Open(*kubeconfig); err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		f.Close()
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}

	}

	errRetry := 0
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get pods
retry:
	pods, err := kubeutils.GetPodIPsLabel(clientset, "default")
	if err != nil {
		errRetry += 1
		if errRetry < 10 {
			goto retry
		}
		panic(err.Error())
	}
	ospd, err := findFirstOpenVAS(pods)
	if err != nil {
		errRetry += 1
		if errRetry < 10 {
			goto retry
		}
		panic(err.Error())
	}
	pd := kubeutils.NewPodCP(*config, clientset, "ospd", ospd.ID)
	if *certPath == "" {
		if err := pd.FromPod("/var/lib/gvm/CA/cacert.pem", "/tmp/ca.pem"); err != nil {
			panic(err.Error())
		}
		*certPath = "/tmp/ca.pem"

	}
	if *certKeyPath == "" {
		if err := pd.FromPod("/var/lib/gvm/private/CA/serverkey.pem", "/tmp/key.pem"); err != nil {
			panic(err.Error())
		}
		*certKeyPath = "/tmp/key.pem"
	}

	address := fmt.Sprintf("%s:%s", ospd.IP, ospd.ExposedPorts[0])
	sender := connection.New("tcp", address, *certPath, *certKeyPath, false)

	m, err := sink.NewMattermost(*mattermostAddress, *mattermostChannelID, *mattermostToken)
	if err != nil {
		panic(err.Error())
	}

	d, err := featuretest.New(pods, *vtDIR, *policyPath, sender)
	if err != nil {
		m.Error(err)
		panic(err.Error())
	}
	fst := findservice.New(&d.ExecInformation)

	d.RegisterTest(fst)
	if results, err := d.Run(); err != nil {
		if *mattermostToken != "" {
			m.Error(err)
		}
		panic(err.Error())
	} else {
		if *mattermostToken != "" {
			m.Send(results)
		}
		for _, r := range results {
			fmt.Printf("%s: %s took %s", r.Name, r.FailureDescription, r.Duration)
		}
	}

}
