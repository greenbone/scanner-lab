// findservice contains tests cases around finding services
package findservice

import (
	"fmt"
	"strings"

	"github.com/greenbone/ospd-openvas/smoketest/nasl"
	"github.com/greenbone/ospd-openvas/smoketest/policies"
	"github.com/greenbone/ospd-openvas/smoketest/scan"
	"github.com/greenbone/ospd-openvas/smoketest/usecases"
	"github.com/greenbone/scanner-lab/feature-tests/converter"
	"github.com/greenbone/scanner-lab/feature-tests/featuretest"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
)

type FindService struct {
	startscan converter.TargetStartScanConverter
	data      *featuretest.ExecInformation
}
type VerifyFoundServicePorts struct {
	Targets []kubeutils.Target
	// host ports
	FoundServicePorts map[string]map[string]bool
}

func NewVFSP(tl int, ts []kubeutils.Target) *VerifyFoundServicePorts {
	return &VerifyFoundServicePorts{
		Targets:           ts,
		FoundServicePorts: make(map[string]map[string]bool, tl),
	}
}

func (pr *VerifyFoundServicePorts) mayAdd(host string, port string) {

	if port != "" && host != "" {
		port = strings.SplitN(port, "/", 2)[0]
		if port != "general" {
			if sp, ok := pr.FoundServicePorts[host]; ok {
				if _, ok := sp[port]; !ok {
					sp[port] = true
					pr.FoundServicePorts[host] = sp
				}
			} else {
				sp = make(map[string]bool)
				sp[port] = true
				pr.FoundServicePorts[host] = sp
			}
		}

	}
}

func (pr *VerifyFoundServicePorts) findTarget(h string) (*kubeutils.Target, error) {
	for _, t := range pr.Targets {
		if t.IP == h {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("%s not found.", h)

}

func (pr *VerifyFoundServicePorts) Add(resp scan.GetScansResponse) {
	for _, r := range resp.Scan.Results.Results {
		pr.mayAdd(r.Host, r.Port)
	}
}

func (pr *VerifyFoundServicePorts) MissingPorts() ([]string, bool) {
	if len(pr.FoundServicePorts) == 0 {
		return nil, false
	}
	missing := make([]string, 0, len(pr.Targets))
	for k, v := range pr.FoundServicePorts {
		if t, err := pr.findTarget(k); err == nil {
			for _, p := range t.RequiredFoundPorts() {
				for k := range v {
					if p == k {
						goto found
					}
				}
				missing = append(missing, p)
			found:
				// needed to continue parent loop without having to loop through RequiredPorts
			}
		}

	}
	return missing, true
}

func (fs *FindService) Name() string {
	return "Full and fast"
}

func (fs *FindService) Start() scan.Start {
	// we always have to start Discovery first to enable full and fast requirements
	return fs.startscan.Convert(fs.data.Targets, []string{"Discovery", fs.Name()}, nil)
}

func (fs *FindService) Verify(resp *usecases.GetScanResponseFailure) featuretest.Result {
	data := fs.data

	vfsp := NewVFSP(len(data.Targets), data.Targets)
	vfsp.Add(resp.Resp)
	var failure string
	if missing, ok := vfsp.MissingPorts(); ok {
		if len(missing) != 0 {
			failure = fmt.Sprintf("ports: %+v not found.", missing)
		}
	} else {
		failure = "No known host information found."
	}
	return featuretest.Result{
		Name:               fs.Name(),
		FailureDescription: failure,
		Resp:               &resp.Resp,
	}

}

func New(data *featuretest.ExecInformation) featuretest.Runner {
	return &FindService{
		startscan: converter.NewTargetStartScan(data.NASLCache, data.PolicyCache),
		data:      data,
	}
}
