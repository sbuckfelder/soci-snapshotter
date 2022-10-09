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
        outputDir = "./output"
	containerdAddress = "/tmp/containerd-grpc/containerd.sock"
	containerdRoot    = "/tmp/lib/containerd"
	containerdState   = "/tmp/containerd"
	containerdConfig  = "./containerd-config.toml"
	platform          = "linux/amd64"
        sociBinary = "../out/soci-snapshotter-grpc"
        sociAddress = "/tmp/soci-snapshotter-grpc/soci-snapshotter-grpc.sock"
        sociRoot = "/tmp/lib/soci-snapshotter-grpc"
        sociConfig = "./soci_config.toml"
        awsSecretFile = "./aws_secret"
)

type ImageDescriptor struct {
	shortName string
	imageRef  string
        sociIndexManifestRef string
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
/*
		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "OverlayFSPull" + image.shortName,
			NumberOfTests: 5,
			TestFunction: func(b *testing.B) {
				BenchmarkPullImageFromECR(b, image.imageRef)
			},
		})
*/
		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "OverlayFSRun" + image.shortName,
			NumberOfTests: 5,
			TestFunction: func(b *testing.B) {
				BenchmarkRunContainerFromECR(b, image.imageRef)
			},
		})
/*
		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "SociRPull" + image.shortName,
			NumberOfTests: 5,
			TestFunction: func(b *testing.B) {
				BenchmarkSociRPullPullImage(b, image.imageRef, image.sociIndexManifestRef)
			},
		})
		drivers = append(drivers, framework.BenchmarkTestDriver{
			TestName:      "SociRun" + image.shortName,
			NumberOfTests: 5,
			TestFunction: func(b *testing.B) {
				BenchmarkSociRunContainer(b, image.imageRef, image.sociIndexManifestRef)
			},
		})
*/
	}

	benchmarks := framework.BenchmarkFramework{
		OutputDir: outputDir,
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
			imageRef:  image[1],
			sociIndexManifestRef: image[2]})
	}
	return images, nil
}
