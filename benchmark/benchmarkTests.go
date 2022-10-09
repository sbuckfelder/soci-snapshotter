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
	"testing"

	"github.com/awslabs/soci-snapshotter/benchmark/framework"
)

func BenchmarkPullImage(b *testing.B, imageRef string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Error Starting Containerd: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	b.ResetTimer()
	_, err = containerdProcess.PullImage(imageRef, platform)
	if err != nil {
                fmt.Printf("Error Pulling Image: %v\n", err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkRunContainer(b *testing.B, imageRef string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Error Starting Containerd: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	image, err := containerdProcess.PullImage(imageRef, platform)
	if err != nil {
                fmt.Printf("Error Pulling Image: %v\n", err)
		b.Fatal(err)
	}
	b.ResetTimer()
	err = containerdProcess.RunContainer(image)
	if err != nil {
                fmt.Printf("Error Running Container: %v\n", err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkPullImageFromECR(b *testing.B, imageRef string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Error Starting Containerd: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	b.ResetTimer()
	_, err = containerdProcess.PullImageFromECR(imageRef, platform, awsSecretFile)
	if err != nil {
                fmt.Printf("Error Pulling Image: %v\n", err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkRunContainerFromECR(b *testing.B, imageRef string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Error Starting Containerd: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	image, err := containerdProcess.PullImageFromECR(imageRef, platform, awsSecretFile)
        if err != nil {
                fmt.Printf("Error Pulling Image: %v\n", err)
		b.Fatal(err)
	}
	b.ResetTimer()
	err = containerdProcess.RunContainer(image)
	if err != nil {
                fmt.Printf("Error Running Container: %v\n", err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkSociRPullPullImage(
	b *testing.B,
	imageRef string,
	indexDigest string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Failed to create containerd proc: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	sociProcess, err := getSociProcess(b)	
        if err != nil {
                fmt.Printf("Failed to create soci proc: %v\n", err)
		b.Fatal(err)
	}
	defer sociProcess.StopProcess()
	sociContainerdProc := SociContainerdProcess{containerdProcess}
	b.ResetTimer()
	_, err = sociContainerdProc.SociRPullImageFromECR(imageRef, indexDigest, awsSecretFile)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func BenchmarkSociRunContainer(b *testing.B, imageRef string, indexDigest string) {
	containerdProcess, err := getContainerdProcess(b) 
	if err != nil {
                fmt.Printf("Error Starting Containerd: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	sociProcess, err := getSociProcess(b)	
        if err != nil {
                fmt.Printf("Failed to create soci proc: %v\n", err)
		b.Fatal(err)
	}
	defer sociProcess.StopProcess()
	sociContainerdProc := SociContainerdProcess{containerdProcess}
        image, err := sociContainerdProc.SociRPullImageFromECR(imageRef, indexDigest, awsSecretFile)
        if err != nil {
                fmt.Printf("Error Pulling Image: %v\n", err)
		b.Fatal(err)
	}
	b.ResetTimer()
	err = sociContainerdProc.SociRunImageFromECR(image, indexDigest)
	if err != nil {
                fmt.Printf("Error Running Container: %v\n", err)
		b.Fatal(err)
	}
	b.StopTimer()
}

func getContainerdProcess(b *testing.B) (*framework.ContainerdProcess, error) {
	return framework.StartContainerd(
		b,
		containerdAddress,
		containerdRoot,
		containerdState,
		containerdConfig,
		outputDir)
}

func getSociProcess(b *testing.B) (*SociProcess, error) {
	return StartSoci(
		sociBinary,
		sociAddress,
		sociRoot,
		containerdAddress,
                sociConfig,
		outputDir)
            }
