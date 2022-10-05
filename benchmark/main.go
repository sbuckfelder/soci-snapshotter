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
	"os"
	"testing"

	"github.com/awslabs/soci-snapshotter/benchmark/framework"
)

var (
	benchmarkOutput   = "./output.json"
	containerdAddress = "/tmp/containerd-grpc/containerd.sock"
	containerdRoot    = "/tmp/lib/containerd"
	containerdState   = "/tmp/containerd"
	containerdConfig  = "./containerd-config.toml"
	containerdOutput  = "./containerd-out"
	dockerImage       = "docker.io/library/hello-world:latest"
	platform          = "linux/amd64"
)

func main() {
	commit := os.Args[1]
	var drivers []framework.BenchmarkTestDriver
	drivers = append(drivers, framework.BenchmarkTestDriver{
		TestName:      "DockerPullImage",
		NumberOfTests: 10,
		TestFunction: func(b *testing.B) {
			BenchmarkPullImage(b, dockerImage)
		},
	})
	drivers = append(drivers, framework.BenchmarkTestDriver{
		TestName:      "DockerRunContainer",
		NumberOfTests: 10,
		TestFunction: func(b *testing.B) {
			BenchmarkRunContainer(b, dockerImage)
		},
	})

	benchmarks := framework.BenchmarkFramework{
		OutputFile: benchmarkOutput,
		CommitID:   commit,
		Drivers:    drivers,
	}
	benchmarks.Run()
}

func BenchmarkPullImage(b *testing.B, imageRef string) {
	containerdProcess, err := framework.StartContainerd(
		b,
		containerdAddress,
		containerdRoot,
		containerdState,
		containerdConfig,
		containerdOutput)
	if err != nil {
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	b.ResetTimer()
	_, err = containerdProcess.PullImage(imageRef, platform)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkRunContainer(b *testing.B, imageRef string) {
	containerdProcess, err := framework.StartContainerd(
		b,
		containerdAddress,
		containerdRoot,
		containerdState,
		containerdConfig,
		containerdOutput)
	if err != nil {
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	image, err := containerdProcess.PullImage(imageRef, platform)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	err = containerdProcess.RunContainer(image)
	if err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
}
