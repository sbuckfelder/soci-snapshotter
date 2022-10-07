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
	"encoding/csv"
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
	platform          = "linux/amd64"
)

type ImageDescriptor struct {
	shortName string
	imageRef  string
}

func main() {
	commit := os.Args[1]
	configCsv := os.Args[2]
	imageList, err := getImageListFromCsv(configCsv)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read csv file %s with error:%v\n", configCsv, err)
		panic(errMsg)
	}
	var drivers []framework.BenchmarkTestDriver
	for _, image := range imageList {

		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "OverlayFSPull" + image.shortName,
			NumberOfTests: 10,
			TestFunction: func(b *testing.B) {
				BenchmarkPullImage(b, image.imageRef)
			},
		})
		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "OverlayFSRun" + image.shortName,
			NumberOfTests: 10,
			TestFunction: func(b *testing.B) {
				BenchmarkRunContainer(b, image.imageRef)
			},
		})
	}

	benchmarks := framework.BenchmarkFramework{
		OutputFile: benchmarkOutput,
		CommitID:   commit,
		Drivers:    drivers,
	}
	benchmarks.Run()
}

func getImageListFromCsv(csvLoc string) ([]ImageDescriptor, error) {
	csvFile, err := os.Open(csvLoc)
	if err != nil {
		return nil, err
	}
	csv, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return nil, err
	}
	var images []ImageDescriptor
	for _, image := range csv {
		images = append(images, ImageDescriptor{
			shortName: image[0],
			imageRef:  image[1]})
	}
	return images, nil
}
