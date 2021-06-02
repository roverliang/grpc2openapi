package cmd

import (
	"github.com/roverliang/grpc2openapi/openapi"
	"github.com/roverliang/grpc2openapi/openapi/descriptor"
	"github.com/roverliang/grpc2openapi/openapi/genopenapi"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/klog/v2"
)

var (
	importPrefix               string
	file                       string
	allowDeleteBody            bool
	grpcAPIConfiguration       string
	allowMerge                 bool
	mergeFileName              string
	useJSONNamesForFields      bool
	repeatedPathParamSeparator string
	versionFlag                bool
	allowRepeatedFieldsInBody  bool
	includePackageInTags       bool
	useFQNForOpenAPIName       bool
	useGoTemplate              bool
	disableDefaultErrors       bool
	enumsAsInts                bool
	simpleOperationIDs         bool
	openAPIConfiguration       string
	generateUnboundMethods     bool
	namespace string
)

func init() {
	GenCommand.Flags().StringVar(&namespace, "namespace", "", "RESTful API prefix")
	GenCommand.Flags().StringVar(&importPrefix, "import_prefix", "", "prefix to be added to go package paths for imported proto files")
	GenCommand.Flags().StringVar(&file, "file", "-", "where to load data from")
	GenCommand.Flags().BoolVar(&allowDeleteBody, "allow_delete_body", false, "unless set, HTTP DELETE methods may not have a body")
	GenCommand.Flags().StringVar(&grpcAPIConfiguration, "grpc_api_configuration", "", "path to file which describes the gRPC API Configuration in YAML format")
	GenCommand.Flags().BoolVar(&allowMerge, "allow_merge", true, "if set, generation one OpenAPI file out of multiple protos")
	GenCommand.Flags().StringVar(&mergeFileName, "merge_file_name", "api", "target OpenAPI file name prefix after merge")
	GenCommand.Flags().BoolVar(&useJSONNamesForFields, "json_names_for_fields", true, "if disabled, the original proto name will be used for generating OpenAPI definitions")
	GenCommand.Flags().StringVar(&repeatedPathParamSeparator, "repeated_path_param_separator", "csv", "configures how repeated fields should be split. Allowed values are `csv`, `pipes`, `ssv` and `tsv`")
	GenCommand.Flags().BoolVar(&versionFlag, "version", false, "print the current version")
	GenCommand.Flags().BoolVar(&allowRepeatedFieldsInBody, "allow_repeated_fields_in_body", false, "allows to use repeated field in `body` and `response_body` field of `google.api.http` annotation option")
	GenCommand.Flags().BoolVar(&includePackageInTags, "include_package_in_tags", false, "if unset, the gRPC service name is added to the `Tags` field of each operation. If set and the `package` directive is shown in the proto file, the package name will be prepended to the service name")
	GenCommand.Flags().BoolVar(&useFQNForOpenAPIName, "fqn_for_openapi_name", false, "if set, the object's OpenAPI names will use the fully qualified names from the proto definition (ie my.package.MyMessage.MyInnerMessage")
	GenCommand.Flags().BoolVar(&useGoTemplate, "use_go_templates", false, "if set, you can use Go templates in protofile comments")
	GenCommand.Flags().BoolVar(&disableDefaultErrors, "disable_default_errors", false, "if set, disables generation of default errors. This is useful if you have defined custom error handling")
	GenCommand.Flags().BoolVar(&enumsAsInts, "enums_as_ints", false, "whether to render enum values as integers, as opposed to string values")
	GenCommand.Flags().BoolVar(&simpleOperationIDs, "simple_operation_ids", false, "whether to remove the service prefix in the operationID generation. Can introduce duplicate operationIDs, use with caution.")
	GenCommand.Flags().StringVar(&openAPIConfiguration, "openapi_configuration", "", "path to file which describes the OpenAPI Configuration in YAML format")
	GenCommand.Flags().BoolVar(&generateUnboundMethods, "generate_unbound_methods", true, "generate swagger metadata even for RPC methods that have no HttpRule annotation")
}

var GenCommand = &cobra.Command{
	Use:   "gen",
	Short: "gen swagger api",
	Run: func(cmd *cobra.Command, args []string) {
		fds, err := openapi.LoadProtosetFile("/Users/roverliang/Downloads/taiping_api.bin")

		if err != nil {
			klog.Error(err)
			return
		}

		reg := descriptor.NewRegistry()

		reg.SetNamespace(namespace)
		reg.SetPrefix(importPrefix)
		reg.SetAllowDeleteBody(allowDeleteBody)
		reg.SetAllowMerge(allowMerge)
		reg.SetMergeFileName(mergeFileName)
		reg.SetUseJSONNamesForFields(useJSONNamesForFields)
		reg.SetAllowRepeatedFieldsInBody(allowRepeatedFieldsInBody)
		reg.SetIncludePackageInTags(includePackageInTags)
		reg.SetUseFQNForOpenAPIName(useFQNForOpenAPIName)
		reg.SetUseGoTemplate(useGoTemplate)
		reg.SetEnumsAsInts(enumsAsInts)
		reg.SetDisableDefaultErrors(disableDefaultErrors)
		reg.SetSimpleOperationIDs(simpleOperationIDs)
		reg.SetGenerateUnboundMethods(generateUnboundMethods)

		gen := genopenapi.New(reg)

		if err := reg.Load(fds); err != nil {
			klog.Info(err)
			return
		}

		var targets []*descriptor.File
		for _, f := range fds {
			f.AsFileDescriptorProto()
			filePath := f.GetFile().GetName()
			f, err := reg.LookupFile(filePath)
			if err != nil {
				klog.Fatal(err)
			}
			targets = append(targets, f)
		}

		out, err := gen.Generate(targets)
		if err != nil {
			klog.Error(err)
			return
		}
		emitResp(out)
	},
}




func emitResp(resp []*descriptor.ResponseFile) {
	if len(resp) == 1 && allowMerge {
		fileName := resp[0].GetName()
		fileContent := resp[0].GetContent()
		err := writeContentToFile(fileName, fileContent)
		if err != nil {
			klog.Error(err)
			return
		}
		return
	}

	for _,file := range resp {
		fileName := file.GetName()
		fileContent := file.GetContent()
		err := writeContentToFile(fileName, fileContent)
		if err != nil {
			klog.Fatal(err)
			return
		}
	}
}



//将文件内容写入文件
func writeContentToFile(filePath string, content string)error{
	return ioutil.WriteFile(filePath, []byte(content),0777)
}
