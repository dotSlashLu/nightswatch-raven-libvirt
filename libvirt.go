package main

import (
	"fmt"
	"time"
	"strconv"
	"github.com/libvirt/libvirt-go"
	ri "github.com/dotSlashLu/nightswatch/agent/raven_interface"
)

const pluginName = "libvirt"

func Report(ch chan *ri.PluginReport, errCh chan *error) {
	conn, err := connect()
	if err != nil {
		errCh <- &err
		return
	}
	defer conn.Close()
	plugin := ri.NewPlugin(pluginName, ch, errCh)
	intervalReport(conn, plugin)
}


func intervalReport(c *libvirt.Connect, plugin *ri.Plugin) {
	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			doms, stats, err := getStats(c)
			if err != nil {
				plugin.ReportError(err)
				return
			}
			err = processRes(plugin, doms, stats)
			if err != nil {
				plugin.ReportError(err)
				return
			}
		}
	}
}

func connect() (*libvirt.Connect, error) {
	return libvirt.NewConnectReadOnly("qemu:///system")
}

func getStats(c *libvirt.Connect) ([]*libvirt.Domain,
	[]libvirt.DomainStats,
	error) {
	allStatsFlag := libvirt.DOMAIN_STATS_STATE |
		libvirt.DOMAIN_STATS_CPU_TOTAL |
		libvirt.DOMAIN_STATS_BALLOON |
		libvirt.DOMAIN_STATS_VCPU |
		libvirt.DOMAIN_STATS_INTERFACE |
		libvirt.DOMAIN_STATS_BLOCK |
		libvirt.DOMAIN_STATS_PERF
	allDomFlag := libvirt.CONNECT_LIST_DOMAINS_ACTIVE |
		libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	doms, err := c.ListAllDomains(allDomFlag)
	if err != nil {
		return nil, nil, err
	}
	// GetAllDomainStats needs []*Domain
	domsPtr := []*libvirt.Domain{}
	for i := range doms {
		domsPtr = append(domsPtr, &doms[i])
	}
	stats, err := c.GetAllDomainStats(domsPtr,
		allStatsFlag,
		libvirt.CONNECT_GET_ALL_DOMAINS_STATS_ACTIVE)
	return domsPtr, stats, err
}

func processRes(p *ri.Plugin,
	domPtrs []*libvirt.Domain, stats []libvirt.DomainStats) error {
	// generated metrics:
	// <machine>.vm.<uuid>.<name, uuid>
	// <machine>.vm.<uuid>.<stats name>
	// <machine>.vm.count
	p.SingleReport(ri.ReportValInt, "vm.count", len(domPtrs))
	for _, dom := range domPtrs {
		domName, _ := dom.GetName()
		domUUID, _ := dom.GetUUIDString()
		kPrefix := "vm." + domUUID

		// name
		nameStr := fmt.Sprintf("%s.%s", kPrefix, "name")
		p.SingleReport(ri.ReportValStr, nameStr, domName)

		// id
		domID, err := dom.GetID()
		if err != nil {
			return err
		}
		domIDString := strconv.Itoa(int(domID))
		idStr := fmt.Sprintf("%s.%s", kPrefix, "id")
		p.SingleReport(ri.ReportValStr, idStr, domIDString)
	}
	return nil
}

