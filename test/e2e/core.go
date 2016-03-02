/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"
)

type command struct {
	cmd       string
	component string
}

func CoreDump(dir string) {
	c, err := loadClient()
	if err != nil {
		fmt.Printf("Error creating client: %v", err)
		return
	}
	provider := testContext.Provider

	// requires ssh to master machine
	if !providerIs(providersWithMasterSSH...) {
		fmt.Printf("Skipping SSH core dump, which is not implemented for %s", provider)
		return
	}

	// I wish there was a better way to get the master IP...
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v", err)
	}
	ix := strings.LastIndex(config.Host, "/")
	master := net.JoinHostPort(config.Host[ix+1:], "22")

	cmds := []command{
		{"cat /var/log/kube-apiserver.log", "kube-apiserver"},
		{"cat /var/log/kube-scheduler.log", "kube-scheduler"},
		{"cat /var/log/kube-controller-manager.log", "kube-controller-manager"},
		{"cat /var/log/etcd.log", "kube-etcd"},
	}
	if isUsingSystemdKubelet(provider, master) {
		cmds = append(cmds, command{"sudo journalctl --output=cat -u kubelet.service", "kubelet"})
	} else {
		cmds = append(cmds, []command{
			{"cat /var/log/kubelet.log", "kubelet"},
			{"cat /var/log/supervisor/supervisord.log", "supervisord"},
			{"cat /var/log/supervisor/kubelet-stdout.log", "supervisord-kubelet-stdout"},
			{"cat /var/log/supervisor/kubelet-stderr.log", "supervisord-kubelet-stderr"},
			{"cat /var/log/kern.log", "kern.log"},
			{"cat /var/log/docker.log", "docker.log"},
		}...)
	}

	logCore(cmds, []string{master}, dir, provider)

	// requires ssh
	if !providerIs(providersWithSSH...) {
		fmt.Printf("Skipping SSH core dump for nodes, which is not implemented for %s", provider)
		return
	}

	// Get all nodes' external IPs.
	hosts, err := NodeSSHHosts(c)
	if err != nil {
		fmt.Printf("Error getting node hostnames: %v", err)
		return
	}

	cmds = []command{{"cat /var/log/kube-proxy.log", "kube-proxy"}}
	if isUsingSystemdKubelet(provider, hosts...) {
		cmds = append(cmds, command{"sudo journalctl --output=cat -u kubelet.service", "kubelet"})
	} else {
		cmds = append(cmds, []command{
			{"cat /var/log/kubelet.log", "kubelet"},
			{"cat /var/log/supervisor/supervisord.log", "supervisord"},
			{"cat /var/log/supervisor/kubelet-stdout.log", "supervisord-kubelet-stdout"},
			{"cat /var/log/supervisor/kubelet-stderr.log", "supervisord-kubelet-stderr"},
			{"cat /var/log/kern.log", "kern.log"},
			{"cat /var/log/docker.log", "docker.log"},
		}...)
	}

	logCore(cmds, hosts, dir, provider)
}

func logCore(cmds []command, hosts []string, dir, provider string) {
	wg := &sync.WaitGroup{}
	// Run commands on all nodes via SSH.
	for _, cmd := range cmds {
		fmt.Printf("SSH'ing to all nodes and running %s\n", cmd.cmd)
		for _, host := range hosts {
			wg.Add(1)
			go func(cmd command, host string) {
				defer wg.Done()
				logfile := fmt.Sprintf("%s/%s-%s.log", dir, host, cmd.component)
				fmt.Printf("Writing to %s.\n", logfile)
				result, err := SSH(cmd.cmd, host, provider)
				if err != nil {
					fmt.Printf("Error running command: %v\n", err)
				}
				if err := ioutil.WriteFile(logfile, []byte(result.Stdout+result.Stderr), 0777); err != nil {
					fmt.Printf("Error writing logfile: %v\n", err)
				}
			}(cmd, host)
		}
	}
	wg.Wait()
}

func isUsingSystemdKubelet(provider string, hosts ...string) bool {
	wg := &sync.WaitGroup{}
	results := make([]bool, len(hosts))
	cmd := "sudo systemctl status kubelet.service"

	wg.Add(len(hosts))
	for i := range hosts {
		go func(i int) {
			defer wg.Done()
			result, err := SSH(cmd, hosts[i], provider)
			if err != nil {
				fmt.Printf("Error running command: %v\n", err)
				return
			}
			if result.Code == 0 {
				results[i] = true
			}
		}(i)
	}
	wg.Wait()

	for _, r := range results {
		if r {
			return true
		}
	}
	return false
}
