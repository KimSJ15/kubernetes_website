// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"time"

	// TODO(rjnagal): Remove dependency after moving all stats structs from v1.
	// using v1 now for easy conversion.
	"github.com/google/cadvisor/info/v1"
)

const (
	TypeName   = "name"
	TypeDocker = "docker"
)

type CpuSpec struct {
	// Requested cpu shares. Default is 1024.
	Limit uint64 `json:"limit"`
	// Requested cpu hard limit. Default is unlimited (0).
	// Units: milli-cpus.
	MaxLimit uint64 `json:"max_limit"`
	// Cpu affinity mask.
	// TODO(rjnagal): Add a library to convert mask string to set of cpu bitmask.
	Mask string `json:"mask,omitempty"`
}

type MemorySpec struct {
	// The amount of memory requested. Default is unlimited (-1).
	// Units: bytes.
	Limit uint64 `json:"limit,omitempty"`

	// The amount of guaranteed memory.  Default is 0.
	// Units: bytes.
	Reservation uint64 `json:"reservation,omitempty"`

	// The amount of swap space requested. Default is unlimited (-1).
	// Units: bytes.
	SwapLimit uint64 `json:"swap_limit,omitempty"`
}

type ContainerInfo struct {
	// Describes the container.
	Spec ContainerSpec `json:"spec,omitempty"`

	// Historical statistics gathered from the container.
	Stats []*ContainerStats `json:"stats,omitempty"`
}

type ContainerSpec struct {
	// Time at which the container was created.
	CreationTime time.Time `json:"creation_time,omitempty"`

	// Other names by which the container is known within a certain namespace.
	// This is unique within that namespace.
	Aliases []string `json:"aliases,omitempty"`

	// Namespace under which the aliases of a container are unique.
	// An example of a namespace is "docker" for Docker containers.
	Namespace string `json:"namespace,omitempty"`

	// Metadata labels associated with this container.
	Labels map[string]string `json:"labels,omitempty"`
	// Metadata envs associated with this container. Only whitelisted envs are added.
	Envs map[string]string `json:"envs,omitempty"`

	HasCpu bool    `json:"has_cpu"`
	Cpu    CpuSpec `json:"cpu,omitempty"`

	HasMemory bool       `json:"has_memory"`
	Memory    MemorySpec `json:"memory,omitempty"`

	HasCustomMetrics bool            `json:"has_custom_metrics"`
	CustomMetrics    []v1.MetricSpec `json:"custom_metrics,omitempty"`

	// Following resources have no associated spec, but are being isolated.
	HasNetwork    bool `json:"has_network"`
	HasFilesystem bool `json:"has_filesystem"`
	HasDiskIo     bool `json:"has_diskio"`

	// Image name used for this container.
	Image string `json:"image,omitempty"`
}

type DeprecatedContainerStats struct {
	// The time of this stat point.
	Timestamp time.Time `json:"timestamp"`
	// CPU statistics
	HasCpu bool `json:"has_cpu"`
	// In nanoseconds (aggregated)
	Cpu v1.CpuStats `json:"cpu,omitempty"`
	// In nanocores per second (instantaneous)
	CpuInst *CpuInstStats `json:"cpu_inst,omitempty"`
	// Disk IO statistics
	HasDiskIo bool           `json:"has_diskio"`
	DiskIo    v1.DiskIoStats `json:"diskio,omitempty"`
	// Memory statistics
	HasMemory bool           `json:"has_memory"`
	Memory    v1.MemoryStats `json:"memory,omitempty"`
	// Network statistics
	HasNetwork bool         `json:"has_network"`
	Network    NetworkStats `json:"network,omitempty"`
	// Filesystem statistics
	HasFilesystem bool         `json:"has_filesystem"`
	Filesystem    []v1.FsStats `json:"filesystem,omitempty"`
	// Task load statistics
	HasLoad bool         `json:"has_load"`
	Load    v1.LoadStats `json:"load_stats,omitempty"`
	// Custom Metrics
	HasCustomMetrics bool                      `json:"has_custom_metrics"`
	CustomMetrics    map[string][]v1.MetricVal `json:"custom_metrics,omitempty"`
}

type ContainerStats struct {
	// The time of this stat point.
	Timestamp time.Time `json:"timestamp"`
	// CPU statistics
	// In nanoseconds (aggregated)
	Cpu *v1.CpuStats `json:"cpu,omitempty"`
	// In nanocores per second (instantaneous)
	CpuInst *CpuInstStats `json:"cpu_inst,omitempty"`
	// Disk IO statistics
	DiskIo *v1.DiskIoStats `json:"diskio,omitempty"`
	// Memory statistics
	Memory *v1.MemoryStats `json:"memory,omitempty"`
	// Network statistics
	Network *NetworkStats `json:"network,omitempty"`
	// Filesystem statistics
	Filesystem *FilesystemStats `json:"filesystem,omitempty"`
	// Task load statistics
	Load *v1.LoadStats `json:"load_stats,omitempty"`
	// Custom Metrics
	CustomMetrics map[string][]v1.MetricVal `json:"custom_metrics,omitempty"`
}

type Percentiles struct {
	// Indicates whether the stats are present or not.
	// If true, values below do not have any data.
	Present bool `json:"present"`
	// Average over the collected sample.
	Mean uint64 `json:"mean"`
	// Max seen over the collected sample.
	Max uint64 `json:"max"`
	// 50th percentile over the collected sample.
	Fifty uint64 `json:"fifty"`
	// 90th percentile over the collected sample.
	Ninety uint64 `json:"ninety"`
	// 95th percentile over the collected sample.
	NinetyFive uint64 `json:"ninetyfive"`
}

type Usage struct {
	// Indicates amount of data available [0-100].
	// If we have data for half a day, we'll still process DayUsage,
	// but set PercentComplete to 50.
	PercentComplete int32 `json:"percent_complete"`
	// Mean, Max, and 90p cpu rate value in milliCpus/seconds. Converted to milliCpus to avoid floats.
	Cpu Percentiles `json:"cpu"`
	// Mean, Max, and 90p memory size in bytes.
	Memory Percentiles `json:"memory"`
}

// latest sample collected for a container.
type InstantUsage struct {
	// cpu rate in cpu milliseconds/second.
	Cpu uint64 `json:"cpu"`
	// Memory usage in bytes.
	Memory uint64 `json:"memory"`
}

type DerivedStats struct {
	// Time of generation of these stats.
	Timestamp time.Time `json:"timestamp"`
	// Latest instantaneous sample.
	LatestUsage InstantUsage `json:"latest_usage"`
	// Percentiles in last observed minute.
	MinuteUsage Usage `json:"minute_usage"`
	// Percentile in last hour.
	HourUsage Usage `json:"hour_usage"`
	// Percentile in last day.
	DayUsage Usage `json:"day_usage"`
}

type FsInfo struct {
	// The block device name associated with the filesystem.
	Device string `json:"device"`

	// Path where the filesystem is mounted.
	Mountpoint string `json:"mountpoint"`

	// Filesystem usage in bytes.
	Capacity uint64 `json:"capacity"`

	// Bytes available for non-root use.
	Available uint64 `json:"available"`

	// Number of bytes used on this filesystem.
	Usage uint64 `json:"usage"`

	// Labels associated with this filesystem.
	Labels []string `json:"labels"`
}

type RequestOptions struct {
	// Type of container identifier specified - "name", "dockerid", dockeralias"
	IdType string `json:"type"`
	// Number of stats to return
	Count int `json:"count"`
	// Whether to include stats for child subcontainers.
	Recursive bool `json:"recursive"`
}

type ProcessInfo struct {
	User          string  `json:"user"`
	Pid           int     `json:"pid"`
	Ppid          int     `json:"parent_pid"`
	StartTime     string  `json:"start_time"`
	PercentCpu    float32 `json:"percent_cpu"`
	PercentMemory float32 `json:"percent_mem"`
	RSS           uint64  `json:"rss"`
	VirtualSize   uint64  `json:"virtual_size"`
	Status        string  `json:"status"`
	RunningTime   string  `json:"running_time"`
	CgroupPath    string  `json:"cgroup_path"`
	Cmd           string  `json:"cmd"`
}

type TcpStat struct {
	Established uint64
	SynSent     uint64
	SynRecv     uint64
	FinWait1    uint64
	FinWait2    uint64
	TimeWait    uint64
	Close       uint64
	CloseWait   uint64
	LastAck     uint64
	Listen      uint64
	Closing     uint64
}

type NetworkStats struct {
	// Network stats by interface.
	Interfaces []v1.InterfaceStats `json:"interfaces,omitempty"`
	// TCP connection stats (Established, Listen...)
	Tcp TcpStat `json:"tcp"`
	// TCP6 connection stats (Established, Listen...)
	Tcp6 TcpStat `json:"tcp6"`
}

// Instantaneous CPU stats
type CpuInstStats struct {
	Usage CpuInstUsage `json:"usage"`
}

// CPU usage time statistics.
type CpuInstUsage struct {
	// Total CPU usage.
	// Units: nanocores per second
	Total uint64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanocores per second
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanocores per second
	User uint64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanocores per second
	System uint64 `json:"system"`
}

// Filesystem usage statistics.
type FilesystemStats struct {
	// Total Number of bytes consumed by container.
	TotalUsageBytes *uint64 `json:"totalUsageBytes,omitempty"`
	// Number of bytes consumed by a container through its root filesystem.
	BaseUsageBytes *uint64 `json:"baseUsageBytes,omitempty"`
}
