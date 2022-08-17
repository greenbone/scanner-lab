// findservice contains tests cases around finding services
package findservice

import (
	"fmt"
	"strings"
	"time"

	"github.com/greenbone/ospd-openvas/smoketest/nasl"
	"github.com/greenbone/ospd-openvas/smoketest/policies"
	"github.com/greenbone/ospd-openvas/smoketest/scan"
	"github.com/greenbone/ospd-openvas/smoketest/usecases"
	"github.com/greenbone/ospd-openvas/smoketest/connection"
	"github.com/greenbone/scanner-lab/feature-tests/converter"
	"github.com/greenbone/scanner-lab/feature-tests/featuretest"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
)

type FindService struct {
	startscan converter.TargetStartScanConverter
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
					fmt.Printf("Add %s:%s\n", host, port)
				}
			} else {
				sp = make(map[string]bool)
				sp[port] = true
				pr.FoundServicePorts[host] = sp
				fmt.Printf("Add %s:%s\n", host, port)

			}
		}

	}
}

func (pr *VerifyFoundServicePorts) Each(resp scan.GetScansResponse) {

	for _, r := range resp.Scan.Results.Results {
		pr.mayAdd(r.Host, r.Port)
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

func (pr *VerifyFoundServicePorts) Last(resp scan.GetScansResponse) {
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
		fmt.Printf("checking %s", k)
		if t, err := pr.findTarget(k); err == nil {
			for _, p := range t.RequiredFoundPorts() {
				for k := range v {
					if p == k {
						goto found
					}
				}
				missing = append(missing, p)
			found:
			}
			fmt.Printf(" missing: %+v\n", missing)
		}

	}
	return missing, true
}

// just for debug purposes it will hide the actual run by sending a stop command when all ports found
var stopWhenFoundAllPorts = false

func startScanPopulateResults(start scan.Start, proto string, sender connection.OSPDSender, mh *VerifyFoundServicePorts) string {
	var result usecases.GetScanResponseFailure
	var startR scan.StartResponse

	if err := sender.SendCommand(start, &startR); err != nil {
		panic(err)
	}
	if startR.Code != "200" {
		return fmt.Sprintf("WrongStatusCode: %s", startR.Code)
	}
	get := scan.GetScans{ID: startR.ID, PopResults: true}

	for !usecases.ScanStatusFinished(result.Resp.Scan.Status) {
		if stopWhenFoundAllPorts {
			if missing, ok := mh.MissingPorts(); ok && len(missing) == 0 {
				break
			}
		}
		// reset to not contain previous results
		result.Resp = scan.GetScansResponse{}
		if err := sender.SendCommand(get, &result.Resp); err != nil {
			panic(err)
		}
		mh.Each(result.Resp)
		if result.Resp.Code != "200" {
			return fmt.Sprintf("WrongStatusCode: %s", startR.Code)
		}
	}
	if stopWhenFoundAllPorts {
		stop := scan.Stop{ID: startR.ID}
		var stopSR scan.StopResponse
		if err := sender.SendCommand(stop, &stopSR); err != nil {
			fmt.Printf("Unable to stop scan %s: %s\n", stop.ID, err)
		} else if stopSR.Code != "200" {
			fmt.Printf("Unable to stop scan %s: %s (%s)\n", stop.ID, stopSR.Code, stopSR.Text)
		}
	}
	mh.Last(result.Resp)
	if missing, ok := mh.MissingPorts(); ok {
		if len(missing) != 0 {
			return fmt.Sprintf("ports: %+v not found.", missing)
		}

	} else {
		return "No host information returned"
	}
	return ""

}

func (fs *FindService) Discovery(data *featuretest.ExecInformation) featuretest.Result {
	cmd := fs.startscan.Convert(data.Targets, []string{"Discovery"}, nil)
	start := time.Now()
	r := startScanPopulateResults(cmd, data.Protocoll, data.OSPDAddr, NewVFSP(len(data.Targets), data.Targets))
	elapsed := time.Now().Sub(start)
	return featuretest.Result{
		Name:               "discovery",
		FailureDescription: r,
		Duration:           elapsed,
	}

}

func New(naslCache *nasl.Cache, policyCache *policies.Cache) *FindService {
	return &FindService{
		startscan: converter.NewTargetStartScan(naslCache, policyCache),
	}
}
