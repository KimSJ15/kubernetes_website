/*
Copyright 2014 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/pkg/term"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/remotecommand"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
)

const (
	attach_example = `# Get output from running pod 123456-7890, using the first container by default
kubectl attach 123456-7890

# Get output from ruby-container from pod 123456-7890
kubectl attach 123456-7890 -c ruby-container

# Switch to raw terminal mode, sends stdin to 'bash' in ruby-container from pod 123456-7890
# and sends stdout/stderr from 'bash' back to the client
kubectl attach 123456-7890 -c ruby-container -i -t`
)

func NewCmdAttach(f *cmdutil.Factory, cmdIn io.Reader, cmdOut, cmdErr io.Writer) *cobra.Command {
	options := &AttachOptions{
		In:  cmdIn,
		Out: cmdOut,
		Err: cmdErr,

		Attach: &DefaultRemoteAttach{},
	}
	cmd := &cobra.Command{
		Use:     "attach POD -c CONTAINER",
		Short:   "Attach to a running container.",
		Long:    "Attach to a process that is already running inside an existing container.",
		Example: attach_example,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(options.Complete(f, cmd, args))
			cmdutil.CheckErr(options.Validate())
			cmdutil.CheckErr(options.Run())
		},
	}
	// TODO support UID
	cmd.Flags().StringVarP(&options.ContainerName, "container", "c", "", "Container name. If omitted, the first container in the pod will be chosen")
	cmd.Flags().BoolVarP(&options.Stdin, "stdin", "i", false, "Pass stdin to the container")
	cmd.Flags().BoolVarP(&options.TTY, "tty", "t", false, "Stdin is a TTY")
	return cmd
}

// RemoteAttach defines the interface accepted by the Attach command - provided for test stubbing
type RemoteAttach interface {
	Attach(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error
}

// DefaultRemoteAttach is the standard implementation of attaching
type DefaultRemoteAttach struct{}

func (*DefaultRemoteAttach) Attach(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewExecutor(config, method, url)
	if err != nil {
		return err
	}
	return exec.Stream(stdin, stdout, stderr, tty)
}

// AttachOptions declare the arguments accepted by the Exec command
type AttachOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Stdin         bool
	TTY           bool

	In  io.Reader
	Out io.Writer
	Err io.Writer

	Attach RemoteAttach
	Client *client.Client
	Config *restclient.Config
}

// Complete verifies command line arguments and loads data from the command environment
func (p *AttachOptions) Complete(f *cmdutil.Factory, cmd *cobra.Command, argsIn []string) error {
	if len(argsIn) == 0 {
		return cmdutil.UsageError(cmd, "POD is required for attach")
	}
	if len(argsIn) > 1 {
		return cmdutil.UsageError(cmd, fmt.Sprintf("expected a single argument: POD, saw %d: %s", len(argsIn), argsIn))
	}
	p.PodName = argsIn[0]

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}
	p.Namespace = namespace

	config, err := f.ClientConfig()
	if err != nil {
		return err
	}
	p.Config = config

	client, err := f.Client()
	if err != nil {
		return err
	}
	p.Client = client

	return nil
}

// Validate checks that the provided attach options are specified.
func (p *AttachOptions) Validate() error {
	allErrs := []error{}
	if len(p.PodName) == 0 {
		allErrs = append(allErrs, fmt.Errorf("pod name must be specified"))
	}
	if p.Out == nil || p.Err == nil {
		allErrs = append(allErrs, fmt.Errorf("both output and error output must be provided"))
	}
	if p.Attach == nil || p.Client == nil || p.Config == nil {
		allErrs = append(allErrs, fmt.Errorf("client, client config, and attach must be provided"))
	}
	return utilerrors.NewAggregate(allErrs)
}

// Run executes a validated remote execution against a pod.
func (p *AttachOptions) Run() error {
	pod, err := p.Client.Pods(p.Namespace).Get(p.PodName)
	if err != nil {
		return err
	}

	if pod.Status.Phase != api.PodRunning {
		return fmt.Errorf("pod %s is not running and cannot be attached to; current phase is %s", p.PodName, pod.Status.Phase)
	}

	var stdin io.Reader
	tty := p.TTY

	containerToAttach := p.GetContainer(pod)
	if tty && !containerToAttach.TTY {
		tty = false
		fmt.Fprintf(p.Err, "Unable to use a TTY - container %s doesn't allocate one\n", containerToAttach.Name)
	}

	// TODO: refactor with terminal helpers from the edit utility once that is merged
	if p.Stdin {
		stdin = p.In
		if tty {
			if file, ok := stdin.(*os.File); ok {
				inFd := file.Fd()
				if term.IsTerminal(inFd) {
					oldState, err := term.SetRawTerminal(inFd)
					if err != nil {
						glog.Fatal(err)
					}
					fmt.Fprintln(p.Out, "\nHit enter for command prompt")
					// this handles a clean exit, where the command finished
					defer term.RestoreTerminal(inFd, oldState)

					// SIGINT is handled by term.SetRawTerminal (it runs a goroutine that listens
					// for SIGINT and restores the terminal before exiting)

					// this handles SIGTERM
					sigChan := make(chan os.Signal, 1)
					signal.Notify(sigChan, syscall.SIGTERM)
					go func() {
						<-sigChan
						term.RestoreTerminal(inFd, oldState)
						os.Exit(0)
					}()
				} else {
					fmt.Fprintln(p.Err, "STDIN is not a terminal")
				}
			} else {
				tty = false
				fmt.Fprintln(p.Err, "Unable to use a TTY - input is not the right kind of file")
			}
		}
	}

	// TODO: consider abstracting into a client invocation or client helper
	req := p.Client.RESTClient.Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("attach")
	req.VersionedParams(&api.PodAttachOptions{
		Container: containerToAttach.Name,
		Stdin:     stdin != nil,
		Stdout:    p.Out != nil,
		Stderr:    p.Err != nil,
		TTY:       tty,
	}, api.ParameterCodec)

	err = p.Attach.Attach("POST", req.URL(), p.Config, stdin, p.Out, p.Err, tty)
	if err != nil {
		return err
	}
	if p.Stdin && tty && pod.Spec.RestartPolicy == api.RestartPolicyAlways {
		fmt.Fprintf(p.Out, "Session ended, resume using 'kubectl attach %s -c %s -i -t' command when the pod is running\n", pod.Name, containerToAttach.Name)
	}
	return nil
}

// GetContainer returns the container to attach to, with a fallback.
func (p *AttachOptions) GetContainer(pod *api.Pod) api.Container {
	if len(p.ContainerName) > 0 {
		for _, container := range pod.Spec.Containers {
			if container.Name == p.ContainerName {
				return container
			}
		}
	}

	glog.V(4).Infof("defaulting container name to %s", pod.Spec.Containers[0].Name)
	return pod.Spec.Containers[0]
}

// GetContainerName returns the name of the container to attach to, with a fallback.
func (p *AttachOptions) GetContainerName(pod *api.Pod) string {
	return p.GetContainer(pod).Name
}
