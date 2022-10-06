package featuretest

import (
	"fmt"
	"sync"
	"time"

	"github.com/greenbone/ospd-openvas/smoketest/connection"
	"github.com/greenbone/ospd-openvas/smoketest/nasl"
	"github.com/greenbone/ospd-openvas/smoketest/policies"
	"github.com/greenbone/ospd-openvas/smoketest/scan"
	"github.com/greenbone/ospd-openvas/smoketest/usecases"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
)

// Result is used to return some information about a test run
type Result struct {
	Name               string                 // Name of the test
	FailureDescription string                 // Description if there was a failure; if there was None then it is empty
	Duration           time.Duration          // How long did it take?
	TargetIDX          []int                  // which targts are used? To reduce memory just the index, empty for all
	Resp               *scan.GetScansResponse // Responses of a scan or nil when none available

}

// ExecInformation contains all information a FeatureTest may need to run
type ExecInformation struct {
	OSPDAddr    connection.OSPDSender // The address to the OSPD Socket
	Protocoll   string
	Targets     []kubeutils.Target // All available targets found via kubernetes
	NASLCache   *nasl.Cache        // All known NASL plugins; is mainly used to lookup OIDs when selecting a policy
	PolicyCache *policies.Cache    // All available Policies
}

// Runner is used to create the start command and a verifier.
//
// The actual start of a test is done in Delegator run. This is to control the output and runtime behaviour.
type Runner interface {
	Name() string                                   // The name of the test
	Start() scan.Start                              // Uses ExecInformation to create a scan.Start command
	Verify(*usecases.GetScanResponseFailure) Result // Uses response of a run to return a Result
}

// Delegator is used to start multiple feature tests
type Delegator struct {
	sync.RWMutex
	ExecInformation
	Runner []Runner
}

type ProgressHandler struct {
	name string
}

func (r *ProgressHandler) Each(resp scan.GetScansResponse) {
	fmt.Printf("\r%s: progress %d", r.name, resp.Scan.Progress)
}

func (r *ProgressHandler) Last(resp scan.GetScansResponse) {
	fmt.Printf("\r%s: progress %d; status: %s; ", r.name, resp.Scan.Progress, resp.Scan.Status)
}
func (d *Delegator) Run() ([]Result, error) {
	result := make([]Result, 0, len(d.Runner))
	if len(d.Runner) == 0 {
		return nil, fmt.Errorf("No runner given.")
	}
	fmt.Println("Running tests")
	for _, r := range d.Runner {
		start := time.Now()
		sr := usecases.StartScanGetLastStatus(r.Start(), d.ExecInformation.OSPDAddr, &ProgressHandler{name: r.Name()})
		elapsed := time.Now().Sub(start)

		re := r.Verify(&sr)
		re.Duration = elapsed
		if re.FailureDescription == "" {
			fmt.Printf("succeeded\n")
		} else {
			fmt.Printf("failed: %s\n", re.FailureDescription)

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
