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
	"fmt"
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
	ecrImage          = "761263122156.dkr.starport.us-west-2.amazonaws.com/tensortest:hello-world"
	// dockerImage = "docker.io/library/python:3"
	// dockerImage = "docker.io/tensorflow/tensorflow:latest"
	dockerImage     = "docker.io/library/hello-world:latest"
	platform        = "linux/amd64"
	awsSecretFile   = "./aws_secret"
	sociBinary      = "../out/soci-snapshotter-grpc"
	sociAddress     = "/tmp/soci-snapshotter-grpc/soci-snapshotter-grpc.sock"
	sociRoot        = "/tmp/lib/soci-snapshotter-grpc"
	sociOutput      = "./soci-snapshotter-grpc-out"
	sociIndexDigest = "sha256:6b4dfbf07ffb0e1349f733eb610c9768f20a2472312509a2d003aa21abbdfc2d"
)

func main() {
	commit := os.Args[1]
	var drivers []framework.BenchmarkTestDriver
	drivers = append(drivers, framework.BenchmarkTestDriver{
		TestName:      "SociRPullTensorHelloWorld",
		NumberOfTests: 10,
		TestFunction: func(b *testing.B) {
			BenchmarkSociRPullPullImage(b, ecrImage, sociIndexDigest)
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
	_, err = containerdProcess.PullImageFromECR(imageRef, platform, awsSecretFile)
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
	image, err := containerdProcess.PullImageFromECR(imageRef, platform, awsSecretFile)
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

func BenchmarkSociRPullPullImage(
	b *testing.B,
	imageRef string,
	indexDigest string) {
	containerdProcess, err := framework.StartContainerd(
		b,
		containerdAddress,
		containerdRoot,
		containerdState,
		containerdConfig,
		containerdOutput)
	if err != nil {
                fmt.Printf("Failed to create containerd proc: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	sociProcess, err := StartSoci(
		sociBinary,
		sociAddress,
		sociRoot,
		containerdAddress,
		sociOutput)
	if err != nil {
                fmt.Printf("Failed to create soci proc: %v\n", err)
		b.Fatal(err)
	}
	defer sociProcess.StopProcess()
	b.ResetTimer()
	sociContainerdProc := SociContainerdProcess{containerdProcess}
	_, err = sociContainerdProc.SociRPullImageFromECR(imageRef, indexDigest, awsSecretFile)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	b.StopTimer()
}
