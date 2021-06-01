package cmd

import (
	"fmt"
	"github.com/roverliang/grpc2openapi/openapi/descriptor"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"os"
)

const (
	swaggerVersion         = "2.0"
	defaultSwaggerFileName = "api_swagger.json"
	defaultApiVersion      = "version not set"
	defaultConsumes        = "application/json"
	defaultProducts        = "application/json"
)

var GenCommand = &cobra.Command{
	Use:   "gen",
	Short: "gen swagger api",
	Run: func(cmd *cobra.Command, args []string) {
		//fds, err := openapi.LoadProtosetFile("/Users/roverliang/Downloads/taiping_api.bin")
		f, err := os.Open("/Users/roverliang/Downloads/taiping_api.bin")
		req, err := ParseRequest(f)
		if err != nil {
			klog.Error(err)
			return
		} else {
			//klog.Info(req)
		}


		reg := descriptor.NewRegistry()
		if err := reg.Load(req); err != nil {
			//klog.Info("hello", err)
			return
		}
		//gen := genopenapi.New(reg)

		var targets []*descriptor.File
		for _, target := range req.FileToGenerate {
			fmt.Println(target)
			f, err := reg.LookupFile(target)
			if err != nil {
				klog.Fatal(err)
			}
			targets = append(targets, f)
		}

		//fmt.Println(targets)

		//generator.Generator()


	},
}


func ParseRequest(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read code generator request: %v", err)
	}

	req := new(pluginpb.CodeGeneratorRequest)
	if err = proto.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal code generator request: %v", err)
	}

	return req, nil
}



