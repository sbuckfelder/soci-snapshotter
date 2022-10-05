/*
   Copyright The Soci Snapshotter Authors.

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

package main

import (
	"context"
	"fmt"
        "io/fs"
	"os"
	"os/exec"
	"time"

	"github.com/awslabs/soci-snapshotter/benchmark/framework"
	"github.com/awslabs/soci-snapshotter/fs/source"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var (
    outputFilePerm fs.FileMode = 0644
)

type SociContainerdProcess struct {
	*framework.ContainerdProcess
}

type SociProcess struct {
	command *exec.Cmd
	address string
	root    string
	stdout  *os.File
	stderr  *os.File
}

func StartSoci(
	sociBinary string,
	sociAddress string,
	sociRoot string,
	containerdAddress string,
	sociOutput string) (*SociProcess, error) {
	sociCmd := exec.Command(sociBinary,
		"-address", sociAddress,
		"-image-service-address", containerdAddress,
		"-root", sociRoot)
        err := os.MkdirAll(sociOutput, outputFilePerm)
	if err != nil {
		return nil, err
	}
	stdoutFile, err := os.Create(sociOutput + "soci-snapshotter-stdout")
	if err != nil {
		return nil, err
	}
	sociCmd.Stdout = stdoutFile
	stderrFile, err := os.Create(sociOutput + "soci-snapshotter-stderr")
	if err != nil {
		return nil, err
	}
	sociCmd.Stderr = stderrFile 
	err = sociCmd.Start()
	if err != nil {
                fmt.Println(err)
		return nil, err
	}
	time.Sleep(4 * time.Second)
	return &SociProcess{
		command: sociCmd,
		address: sociAddress,
		root:    sociRoot,
                stdout: stdoutFile,
                stderr: stderrFile}, nil
}

func (proc *SociProcess) StopProcess() {
	if proc.stdout != nil {
		proc.stdout.Close()
	}
	if proc.stderr != nil {
		proc.stderr.Close()
	}
	if proc.command != nil {
		proc.command.Process.Kill()
	}
        err := os.RemoveAll(proc.address)
        if err != nil {
            fmt.Println(err)
        }
}

func (proc *SociContainerdProcess) SociRPullImageFromECR(
	imageRef string,
	sociIndexDigest string,
	awsSecretFile string) (containerd.Image, error) {
	h := images.HandlerFunc(
		func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			if desc.MediaType != images.MediaTypeDockerSchema1Manifest {
				fmt.Printf("fetching %v... %v\n", desc.Digest.String()[:15], desc.MediaType)
			}
			return nil, nil
		})

	labels := commands.LabelArgs([]string{})
	image, err := proc.Client.Pull(proc.Context, imageRef, []containerd.RemoteOpt{
		containerd.WithPullLabels(labels),
		containerd.WithResolver(framework.GetECRResolver(proc.Context, awsSecretFile)),
		containerd.WithImageHandler(h),
		containerd.WithSchema1Conversion,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("soci"),
		containerd.WithImageHandlerWrapper(source.AppendDefaultLabelsHandlerWrapper(
			imageRef,
			sociIndexDigest)),
	}...)
	if err != nil {
		return nil, err
	}
	return image, nil
}
