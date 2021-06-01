package main

import (
	"github.com/roverliang/grpc2openapi/cmd"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var rootCommand = &cobra.Command{
	Use:     "grpc2openapi",
	Short:   "grpc2openapi relies on reflection or protos or protoset to generate swagger json",
	Version: "v0.4.7",
}

func main() {
	rootCommand.AddCommand()
	err := rootCommand.Execute()
	if err != nil {
		klog.Error(err)
		return
	}
}
