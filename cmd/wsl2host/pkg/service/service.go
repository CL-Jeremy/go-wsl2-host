package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CL-Jeremy/go-wsl2-host/pkg/hostsapi"

	"github.com/CL-Jeremy/go-wsl2-host/pkg/wslapi"
)

const tld = ".wsl"

var hostnamereg, _ = regexp.Compile("[^A-Za-z0-9]+")

func distroNameToHostname(distroname string) string {
	// Ubuntu-18.04
	// => ubuntu1804.wsl
	hostname := strings.ToLower(distroname)
	hostname = hostnamereg.ReplaceAllString(hostname, "")
	return hostname + tld
}

// Run main entry point to service logic
func Run() error {
	infos, err := wslapi.GetAllInfo()
	if err != nil {
		return fmt.Errorf("failed to get infos: %w", err)
	}

	hapi, err := hostsapi.CreateAPI(tld)
	if err != nil {
		return fmt.Errorf("failed to create hosts api: %w", err)
	}

	updated := false
	hostentries := hapi.Entries()
	for _, i := range infos {
		hostname := distroNameToHostname(i.Name)
		// remove stopped distros
		if i.Running == false {
			err := hapi.RemoveEntry(hostname)
			if err == nil {
				updated = true
			}
			continue
		}

		// update IPs of running distros
		ip, err := wslapi.GetIP(i.Name)
		if he, exists := hostentries[hostname]; exists {
			if err != nil {
				return fmt.Errorf("failed to get IP for distro %q: %w", i.Name, err)
			}
			if he.IP != ip {
				updated = true
				he.IP = ip
			}
		} else {
			updated = true
			// add running distros not present
			hapi.AddEntry(&hostsapi.HostEntry{
				Hostname: hostname,
				IP:       ip,
			})
		}
	}

	if updated {
		err = hapi.Write()
		if err != nil {
			return fmt.Errorf("failed to write hosts file: %w", err)
		}
	}

	return nil
}
