package openapi

import (
	"encoding/json"
	"github.com/golang/protobuf/proto"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"io/ioutil"
)

//加载protoset
func LoadProtosetFile(filepath string) ([]*desc.FileDescriptor, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var fileSet descpb.FileDescriptorSet
	if err := proto.Unmarshal(bytes, &fileSet); err != nil {
		return nil, err
	}

	test, err := desc.CreateFileDescriptorsFromSet(&fileSet)
	var FileDs []*desc.FileDescriptor
	for _, val := range test {
		if len(val.GetServices()) > 0 {
			FileDs = append(FileDs, val.GetFile())
		}
	}
	return FileDs, nil
}

func WriteSwaggerJsonToFile(swagger *openapiSwaggerObject)error{
	v, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("api_swagger.json", v, 0777)
	if err != nil {
		return err
	}
	return nil
}

