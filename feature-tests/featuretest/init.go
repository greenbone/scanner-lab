package featuretest

import (
	"fmt"
	"sync"
	"time"

	"github.com/greenbone/ospd-openvas/smoketest/nasl"
	"github.com/greenbone/ospd-openvas/smoketest/policies"
	"github.com/greenbone/ospd-openvas/smoketest/connection"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
)

// Result is used to return some information about a test run
type Result struct {
	Name               string        // Name of the test
	FailureDescription string        // Description if there was a failure; if there was None then it is empty
	Duration           time.Duration // How long did it take?
	TargetIDX          []int         // which targts are used? To reduce memory just the index, empty for all

}

// ExecInformation contains all information a FeatureTest may need to run
type ExecInformation struct {
	OSPDAddr    connection.OSPDSender // The address to the OSPD Socket
	Protocoll   string
	Targets     []kubeutils.Target // All available targets found via kubernetes
	NASLCache   *nasl.Cache        // All known NASL plugins; is mainly used to lookup OIDs when selecting a policy
	PolicyCache *policies.Cache    // All available Policies
}

type Runner func(*ExecInformation) Result

// Delegator is used to start multiple feature tests
type Delegator struct {
	sync.RWMutex
	ExecInformation
	Runner []Runner
}

func (d *Delegator) Run() ([]Result, error) {
	result := make([]Result, 0, len(d.Runner))
	if len(d.Runner) == 0 {
		return nil, fmt.Errorf("No runner given.")
	}
	fmt.Println("Running tests")
	for _, r := range d.Runner {
		re := r(&d.ExecInformation)
		if re.FailureDescription == "" {
			fmt.Printf("%s: succeeded\n", re.Name)
		} else {
			fmt.Printf("%s: failed: %s\n", re.Name, re.FailureDescription)

		}
		result = append(result, re)
	}

	return result, nil
}

func (d *Delegator) RegisterTest(t Runner) {
	d.Lock()
	d.Runner = append(d.Runner, t)
	d.Unlock()
}

// New initializes the needed caches, get targets and creates a Delegator
func New(ts []kubeutils.Target, vtDIR string, policyPath string, address connection.OSPDSender) (result *Delegator, err error) {
	naslCache, err := nasl.InitCache(vtDIR)
	if err != nil {
		return
	}

	policyCache, err := policies.InitCache(policyPath)
	if err != nil {
		return
	}
	result = &Delegator{
		ExecInformation: ExecInformation{
			OSPDAddr:    address,
			Targets:     ts,
			NASLCache:   naslCache,
			PolicyCache: policyCache,
			Protocoll:   "tcp",
		},
		Runner: make([]Runner, 0, 10),
	}
	return
}
