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

package framework

import (
	"context"
        "io/fs"
	"os"
	"os/exec"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/log/logtest"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
)

var (
	testNamespace = "BENCHMARK_TESTING"
        testContainerId = "TEST_RUN_CONTAINER"
        testEnvironment = "TEST_RUNTIME"
        outputFilePerm fs.FileMode = 0644
)

type ContainerdProcess struct {
	command       *exec.Cmd
	address       string
	root          string
	state         string
	stdout        *os.File
	stderr        *os.File
	Client        *containerd.Client
	Context       context.Context
	cancelContext context.CancelFunc
}

func StartContainerd(
	b *testing.B,
	containerdAddress string,
	containerdRoot string,
	containerdState string,
	containerdConfig string,
	containerdOutput string) (*ContainerdProcess, error) {
	containerdCmd := exec.Command("containerd",
		"-a", containerdAddress,
		"--root", containerdRoot,
		"--state", containerdState,
		"-c", containerdConfig)
        err := os.MkdirAll(containerdOutput, outputFilePerm)
	if err != nil {
		return nil, err
	}
	stdoutFile, err := os.Create(containerdOutput + "/containerd-stdout")
	if err != nil {
		return nil, err
	}
	containerdCmd.Stdout = stdoutFile 
	stderrFile, err := os.Create(containerdOutput + "/containerd-stderr")
	if err != nil {
		return nil, err
	}
	containerdCmd.Stderr = stderrFile
	err = containerdCmd.Start()
	if err != nil {
		return nil, err
	}
	client, err := newClient(containerdAddress)
	if err != nil {
		return nil, err
	}
	ctx, cancelCtx := testContext(b)

	return &ContainerdProcess{
		command:       containerdCmd,
		address:       containerdAddress,
		root:          containerdRoot,
		stdout:        stdoutFile,
		stderr:        stderrFile,
		state:         containerdState,
		Client:        client,
		Context:       ctx,
		cancelContext: cancelCtx}, nil
}

func (proc *ContainerdProcess) StopProcess() {
	if proc.Client != nil {
		proc.Client.Close()
	}
	if proc.stdout != nil {
		proc.stdout.Close()
	}
	if proc.stderr != nil {
		proc.stderr.Close()
	}
	if proc.cancelContext != nil {
		proc.cancelContext()
	}
	if proc.command != nil {
		proc.command.Process.Kill()
	}
	os.RemoveAll(proc.root)
	os.RemoveAll(proc.state)
	os.RemoveAll(proc.address)
}

func (proc *ContainerdProcess) PullImage(
	imageRef string,
	platform string) (containerd.Image, error) {
	image, pullErr := proc.Client.Pull(proc.Context, imageRef, GetRemoteOpts(proc.Context, platform)...)
	if pullErr != nil {
		return nil, pullErr
	}
	return image, nil
}


func (proc *ContainerdProcess) RunContainer(image containerd.Image) error {
	container, err := proc.Client.NewContainer(
		proc.Context,
		testContainerId,
		containerd.WithNewSnapshot(testContainerId, image),
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

func GetRemoteOpts(ctx context.Context, platform string) []containerd.RemoteOpt {
	var opts []containerd.RemoteOpt
	opts = append(opts, containerd.WithPlatform(platform))
	opts = append(opts, containerd.WithPullUnpack)

	return opts
}

func testContext(t testing.TB) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = namespaces.WithNamespace(ctx, testNamespace)
	if t != nil {
		ctx = logtest.WithT(ctx, t)
	}
	return ctx, cancel
}

func newClient(address string) (*containerd.Client, error) {
	opts := []containerd.ClientOpt{}
	if rt := os.Getenv(testEnvironment); rt != "" {
		opts = append(opts, containerd.WithDefaultRuntime(rt))
	}

	return containerd.New(address, opts...)
}
