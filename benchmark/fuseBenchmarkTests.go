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

package benchmark 

import (
        "fmt"
	"testing"
)

func BenchmarkOverlayFSRunContainerPrePull(
	b *testing.B,
	imageRef string,
        readyLine string) {
	containerdProcess, err := getContainerdProcess() 
	if err != nil {
                fmt.Printf("Failed to create containerd proc: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
        image, err := containerdProcess.PullImageFromECR(imageRef, platform, awsSecretFile)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        prepContainer, cleanupPrepContainer, err := containerdProcess.CreateContainer(image)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupPrepContainer()
        taskPrepDetails, cleanupPrepTask, err := containerdProcess.CreateTask(prepContainer)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupPrepTask()
        cleanupPrepRun, err := containerdProcess.RunContainerTaskForReadyLine(taskPrepDetails, readyLine)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        defer cleanupPrepRun()
	b.ResetTimer()
        container, cleanupContainer, err := containerdProcess.CreateContainer(image)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupContainer()
        taskDetails, cleanupTask, err := containerdProcess.CreateTask(container)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupTask()
        cleanupRun, err := containerdProcess.RunContainerTaskForReadyLine(taskDetails, readyLine)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        defer cleanupRun()
        b.StopTimer()
}

func BenchmarkSociRunContainerPrePull(
	b *testing.B,
	imageRef string,
	indexDigest string,
        readyLine string) {
	containerdProcess, err := getContainerdProcess() 
	if err != nil {
                fmt.Printf("Failed to create containerd proc: %v\n", err)
		b.Fatal(err)
	}
	defer containerdProcess.StopProcess()
	sociProcess, err := getSociProcess()	
        if err != nil {
                fmt.Printf("Failed to create soci proc: %v\n", err)
		b.Fatal(err)
	}
	defer sociProcess.StopProcess()
	sociContainerdProc := SociContainerdProcess{containerdProcess}
        image, err := sociContainerdProc.SociRPullImageFromECR(imageRef, indexDigest, awsSecretFile)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        prepContainer, cleanupPrepContainer, err := sociContainerdProc.CreateSociContainer(image)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupPrepContainer()
        prepTaskDetails, cleanupPrepTask, err := sociContainerdProc.CreateTask(prepContainer)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupPrepTask()
        cleanupPrepRun, err := sociContainerdProc.RunContainerTaskForReadyLine(prepTaskDetails, readyLine)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        defer cleanupPrepRun()
        b.ResetTimer()
        container, cleanupContainer, err := sociContainerdProc.CreateSociContainer(image)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupContainer()
        taskDetails, cleanupTask, err := sociContainerdProc.CreateTask(container)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
	defer cleanupTask()
        cleanupRun, err := sociContainerdProc.RunContainerTaskForReadyLine(taskDetails, readyLine)
	if err != nil {
		fmt.Println(err)
		b.Fatal(err)
	}
        defer cleanupRun()
        b.StopTimer()
}
