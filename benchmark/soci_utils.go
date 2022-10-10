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
        "errors"
	"fmt"
        "io/fs"
	"os"
	"os/exec"
	"time"

	"github.com/awslabs/soci-snapshotter/benchmark/framework"
	"github.com/awslabs/soci-snapshotter/fs/source"
	"github.com/containerd/containerd"
        "github.com/containerd/containerd/cio"
        "github.com/containerd/containerd/oci"
)

var (
    outputFilePerm fs.FileMode = 0644
    sociContainerId = "TEST_SOCI_CONTAINER"
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
        configFile string,
	outputDir string) (*SociProcess, error) {
	sociCmd := exec.Command(sociBinary,
		"-address", sociAddress,
		"-config", configFile,
		"-root", sociRoot)
        err := os.MkdirAll(outputDir, outputFilePerm)
	if err != nil {
		return nil, err
	}
	stdoutFile, err := os.Create(outputDir + "/soci-snapshotter-stdout")
	if err != nil {
		return nil, err
	}
	sociCmd.Stdout = stdoutFile
	stderrFile, err := os.Create(outputDir + "/soci-snapshotter-stderr")
	if err != nil {
		return nil, err
	}
	sociCmd.Stderr = stderrFile 
	err = sociCmd.Start()
	if err != nil {
                fmt.Printf("Soci Failed to Start %v\n", err)
		return nil, err
	}
        
        // The soci-snapshotter-grpc is not ready to be used until the
        // unix socket file is created
        sleepCount := 0
        loopExit := false
        for loopExit == false {
            time.Sleep(1* time.Second)
            sleepCount += 1
            if _, err := os.Stat(sociAddress); err == nil {
                loopExit = true
            }
            if sleepCount > 10 {
                return nil, errors.New("Could not create .sock in time")
            }
        }
        
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
            fmt.Printf("Error removing Address: %v\n", err)
        }
}

func (proc *SociContainerdProcess) SociRPullImageFromECR(
	imageRef string,
	sociIndexDigest string,
	awsSecretFile string) (containerd.Image, error) {
	image, err := proc.Client.Pull(proc.Context, imageRef, []containerd.RemoteOpt{
		containerd.WithResolver(framework.GetECRResolver(proc.Context, awsSecretFile)),
		containerd.WithSchema1Conversion,
		containerd.WithPullUnpack,
		containerd.WithPullSnapshotter("soci"),
		containerd.WithImageHandlerWrapper(source.AppendDefaultLabelsHandlerWrapper(
			imageRef,
			sociIndexDigest)),
	}...)
	if err != nil {
                fmt.Printf("Soci Pull Failed %v\n", err)
		return nil, err
	}
	return image, nil
}

func (proc *SociContainerdProcess) SociRunImageFromECR(
    image containerd.Image,
    sociIndexDigest string) error {
        id := fmt.Sprintf("%s-%d", sociContainerId, time.Now().UnixNano()) 
	container, err := proc.Client.NewContainer(
		proc.Context,
		id,
                containerd.WithSnapshotter("soci"),
		containerd.WithNewSnapshot(id, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)))
	if err != nil {
		return err
	}
	defer container.Delete(proc.Context, containerd.WithSnapshotCleanup)
	
        task, err := container.NewTask(proc.Context, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}
	defer task.Delete(proc.Context)

	exitStatusC, err := task.Wait(proc.Context)
	if err != nil {
		return err
	}

	if err := task.Start(proc.Context); err != nil {
		return err
	}
	status := <-exitStatusC
	_, _, err = status.Result()
	if err != nil {
		return err
	}
	return nil
}
