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

package tcp

import (
	"net"
	"strconv"
	"time"

	"k8s.io/kubernetes/pkg/probe"

	"github.com/golang/glog"
)

func New() TCPProber {
	return tcpProber{}
}

type TCPProber interface {
	Probe(host string, port int, timeout time.Duration) (probe.Result, string, error)
}

type tcpProber struct{}

func (pr tcpProber) Probe(host string, port int, timeout time.Duration) (probe.Result, string, error) {
	return DoTCPProbe(net.JoinHostPort(host, strconv.Itoa(port)), timeout)
}

// DoTCPProbe checks that a TCP socket to the address can be opened.
// If the socket can be opened, it returns Success
// If the socket fails to open, it returns Failure.
// This is exported because some other packages may want to do direct TCP probes.
func DoTCPProbe(addr string, timeout time.Duration) (probe.Result, string, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		// Convert errors to failures to handle timeouts.
		return probe.Failure, err.Error(), nil
	}
	err = conn.Close()
	if err != nil {
		glog.Errorf("Unexpected error closing TCP probe socket: %v (%#v)", err, err)
	}
	return probe.Success, "", nil
}
