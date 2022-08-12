package converter

import (
	"strings"

	"github.com/greenbone/ospd-openvas/smoketest/nasl"
	"github.com/greenbone/ospd-openvas/smoketest/policies"
	"github.com/greenbone/ospd-openvas/smoketest/scan"
	"github.com/greenbone/scanner-lab/feature-tests/kubeutils"
)

// work around for ospd bug that we must provide ScannerParams although we set nothing
var defaultScannerParams = []scan.ScannerParam{
	{},
}

type TargetStartScanConverter interface {
	Convert(target []kubeutils.Target, pols []string, oids []string) scan.Start
}

type targetStartScanConverter struct {
	naslCache     *nasl.Cache
	policiesCache *policies.Cache
	alivemethod   scan.AliveTestMethods
}

// selection transform given policy names and oids to a scan.VTSelection
func (tc *targetStartScanConverter) selection(pols []string, oids []string) []scan.VTSelection {
	selection := scan.VTSelection{
		Single: make([]scan.VTSingle, 0),
		Group:  make([]scan.VTGroup, 0),
	}
	for _, p := range pols {
		ps := tc.policiesCache.ByName(p).AsVTSelection(tc.naslCache)
		selection.Group = append(selection.Group, ps.Group...)
		selection.Single = append(selection.Single, ps.Single...)
	}

	for _, oid := range oids {
		selection.Single = append(selection.Single, scan.VTSingle{
			ID: oid,
		})
	}

	return []scan.VTSelection{selection}
}

func (tc *targetStartScanConverter) transformtarget(t kubeutils.Target) (string, string) {
	ports := strings.Join(t.RequiredFoundPorts(), ",")
	return t.IP, ports
}

// transformtargets transforms given kubeutils.Target into scan.Targets
//
// Since ospd does create a new scan for each target we unfortunately have to artifically create
// a comma separated hosts string instead of using a proper list of scan.Target therefore it
// currently just has one element.
func (tc *targetStartScanConverter) transformtargets(ts []kubeutils.Target) scan.Targets {
	hosts := make([]string, 0, len(ts))
	ports := make([]string, 0, len(ts))
	allPorts := make(map[string]bool, len(ts))
	for _, t := range ts {
		for _, p := range t.RequiredFoundPorts() {
			if _, ok := allPorts[p]; !ok {
				allPorts[p] = true
				ports = append(ports, p)
			}
		}
		hosts = append(hosts, t.IP)
	}
	t := scan.Target{
		Hosts:            strings.Join(hosts, ","),
		Ports:            strings.Join(ports, ","),
		AliveTestMethods: tc.alivemethod,
	}

	return scan.Targets{
		Targets: []scan.Target{t},
	}
}

// Convert converts given targets and pols as well as oids to a start scan.Start command
func (tc *targetStartScanConverter) Convert(target []kubeutils.Target, pols []string, oids []string) scan.Start {
	sel := tc.selection(pols, oids)
	ts := tc.transformtargets(target)
	return scan.Start{
		Targets:       ts,
		VTSelection:   sel,
		ScannerParams: defaultScannerParams,
	}

}

func NewTargetStartScan(
	naslCache *nasl.Cache,
	policiesCache *policies.Cache,
) TargetStartScanConverter {
	alive := scan.AliveTestMethods{
		ICMP:          1,
		TCPSYN:        1,
		TCPACK:        1,
		ARP:           1,
		ConsiderAlive: 0,
	}
	return &targetStartScanConverter{
		naslCache,
		policiesCache,
		alive,
	}

}
