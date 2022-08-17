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

// Note: the example only works with the code within the same release/branch.
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
	certPath := flag.String("cert-path", "/var/lib/gvm/CA/servercert.pem", "(optional) path to the certificate used by ospd.")
	certKeyPath := flag.String("certkey-path", "/var/lib/gvm/private/CA/serverkey.pem", "(optional) path to certificate key used by ospd.")
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()

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

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// get pods
	pods, err := kubeutils.GetPodIPsLabel(clientset, "default")
	if err != nil {
		panic(err.Error())
	}
	ospd, err := findFirstOpenVAS(pods)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Using certPath %s and keyPath %s\n", *certPath, *certKeyPath)
	address := fmt.Sprintf("%s:%s", ospd.IP, ospd.ExposedPorts[0])
	sender := connection.New("tcp", address, *certPath, *certKeyPath, false)

	d, err := featuretest.New(pods, *vtDIR, *policyPath, sender)
	if err != nil {
		panic(err.Error())
	}
	fst := findservice.New(d.NASLCache, d.PolicyCache)
	d.RegisterTest(fst.Discovery)
	if results, err := d.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error while running the tests: %s", err)
	} else {
		for _, r := range results {
			fmt.Printf("%s: %s took %s", r.Name, r.FailureDescription, r.Duration)
		}
	}

}
