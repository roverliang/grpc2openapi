package openapi

import (
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

const reflectionProto = "reflection.proto"

type httpInfo struct {
	path        string
	requestType string
	body        string
}

type pathInfo struct {
	isAnnotation bool
	ds     grpcurl.DescriptorSource
	fd     *desc.FileDescriptor
	svc    *desc.ServiceDescriptor
	method *desc.MethodDescriptor
	remark string
	httpInfo
}

func (swagger *openapiSwaggerObject) SetExtensions(extensions []extension) {
	swagger.extensions = extensions
}

func (swagger *openapiSwaggerObject) SetExternalDocs(ExternalDocs *openapiExternalDocumentationObject) {
	swagger.ExternalDocs = ExternalDocs
}

func (swagger *openapiSwaggerObject) SetSecurity(Security []openapiSecurityRequirementObject) {
	swagger.Security = Security
}

func (swagger *openapiSwaggerObject) SetSecurityDefinitions(SecurityDefinitions openapiSecurityDefinitionsObject) {
	swagger.SecurityDefinitions = SecurityDefinitions
}

func (swagger *openapiSwaggerObject) SetDefinitions(Definitions openapiDefinitionsObject) {
	swagger.Definitions = Definitions
}

func (swagger *openapiSwaggerObject) SetPaths(Paths openapiPathsObject) {
	swagger.Paths = Paths
}

func (swagger *openapiSwaggerObject) SetProduces(Produces string) {
	swagger.Produces = append([]string{}, Produces)
}

func (swagger *openapiSwaggerObject) SetConsumes(Consumes string) {
	swagger.Consumes = append([]string{}, Consumes)
}

func (swagger *openapiSwaggerObject) SetSchemes(Schemes []string) {
	swagger.Schemes = Schemes
}

func (swagger *openapiSwaggerObject) SetBasePath(BasePath string) {
	swagger.BasePath = BasePath
}

func (swagger *openapiSwaggerObject) SetHost(Host string) {
	swagger.Host = Host
}

func (swagger *openapiSwaggerObject) SetTags(tags []string) {
	var t []openapiTagObject
	for _, name := range tags {
		t = append(t, openapiTagObject{Name: name})
	}
	swagger.Tags = t
}

func (swagger *openapiSwaggerObject) SetInfo(Info openapiInfoObject) {
	swagger.Info = Info
}

func NewSwaggerObject() *openapiSwaggerObject {
	return &openapiSwaggerObject{}
}

func (swagger *openapiSwaggerObject) SetSwagger(Swagger string) {
	swagger.Swagger = Swagger
}

// NewOpenapiInfoObject 生成OpenAPI info
//    "info":{
//        "title":"grpc_transcoding_http.proto",
//        "version":"version not set"
//    },
func NewOpenapiInfoObject(title string, version string) openapiInfoObject {
	return openapiInfoObject{
		Title:   title,
		Version: version,
	}
}

//获取protoset 的Path
func Tags(fds []*desc.FileDescriptor) (tags []string) {
	for _, fd := range fds {
		if strings.Contains(fd.GetFile().GetName(), reflectionProto) {
			continue
		}
		svcDs := fd.GetServices()
		for _, svc := range svcDs {
			tags = append(tags, svc.GetName())
		}
	}
	return tags
}

func Paths(fds []*desc.FileDescriptor) (openapiPathsObject, error) {
	var openapiPathObj = make(openapiPathsObject)

	for _, fd := range fds {
		if strings.Contains(fd.GetFile().GetName(), reflectionProto) {
			continue
		}

		svcs := fd.GetServices()
		if len(svcs) == 0 {
			continue
		}

		for _, svc := range svcs {
			methods := svc.GetMethods()
			if len(methods) == 0 {
				continue
			}


			for _, method := range methods {
				paths, err :=getMethodOptions(svc, method)
				if err != nil {
					return nil, err
				}
				openapiPathObj.setOpenapiPathsObject(paths)
			}
		}
	}
	return openapiPathObj, nil
}


//处理方法的Options
func getMethodOptions(svc *desc.ServiceDescriptor, method *desc.MethodDescriptor) ([]pathInfo, error) {
	var paths []pathInfo

	remarkOrigin := method.GetSourceInfo().GetLeadingComments()
	remark := strings.Trim(strings.ReplaceAll(remarkOrigin, "\n", ""), "\n")
	ds, _ := grpcurl.DescriptorSourceFromFileDescriptors(svc.GetFile())

	fd := svc.GetFile()
	md := method.AsMethodDescriptorProto()

	//提取无option 的情况
	if !proto.HasExtension(md.Options, annotations.E_Http) {
		paths = append(paths, pathInfo{
			isAnnotation: false,
			ds:     ds,
			fd:     fd,
			svc:    svc,
			method: method,
			remark: remark,
			httpInfo: httpInfo{
				path:        fmt.Sprintf("/%s.%s/%s", svc.GetFile().GetPackage(), svc.GetName(), method.GetName()),
				requestType: "POST",
				body:        "*",
			},
		})
		return paths, nil
	}

	optExt, err := extractMethodOptions(method.AsMethodDescriptorProto())
	if err != nil {
		return nil, errors.Wrap(err, "Parse HTTP OPTIONS errors")
	}

	//提取单option
	info, err := extractGoogleApiHttpMethodOptions(optExt)
	if err != nil {
		return nil, err
	}

	paths = append(paths, pathInfo{
		isAnnotation: true,
		ds:       ds,
		fd:       fd,
		svc:      svc,
		method:   method,
		remark:   remark,
		httpInfo: info,
	})


	//提取多Option
	if len(optExt.AdditionalBindings) != 0 {
		for _, hd := range optExt.AdditionalBindings {
			if info, err := extractGoogleApiHttpMethodOptions(hd); err != nil {
				continue
			} else {
				paths = append(paths, pathInfo{
					isAnnotation: true,
					ds:       ds,
					fd:       fd,
					svc:      svc,
					method:   method,
					remark:   remark,
					httpInfo: info,
				})
			}
		}
	}

	return paths, nil
}

func extractGoogleApiHttpMethodOptions(optExt *annotations.HttpRule) (info httpInfo, err error) {
	if optExt.GetGet() != "" {
		info = httpInfo{
			path:        optExt.GetGet(),
			requestType: "GET",
			body:        "",
		}
		return
	}

	if optExt.GetPost() != "" {
		info = httpInfo{
			path:        optExt.GetPost(),
			requestType: "POST",
			body:        optExt.GetBody(),
		}
		return
	}

	if optExt.GetPatch() != "" {
		info = httpInfo{
			path:        optExt.GetPatch(),
			requestType: "PATCH",
			body:        optExt.GetBody(),
		}
		return
	}

	if optExt.GetPut() != "" {
		info = httpInfo{
			path:        optExt.GetPut(),
			requestType: "PUT",
			body:        optExt.GetBody(),
		}
		return
	}

	if optExt.GetDelete() != "" {
		info = httpInfo{
			path:        optExt.GetDelete(),
			requestType: "PUT",
			body:        optExt.GetBody(),
		}
		return
	}

	return info, errors.New("Failed to parse option")
}

func extractMethodOptions(md *descriptorpb.MethodDescriptorProto) (*annotations.HttpRule, error) {
	ext, _ := proto.GetExtension(md.Options, annotations.E_Http)
	opts, ok := ext.(*annotations.HttpRule)

	if !ok {
		return nil, fmt.Errorf("extension is %T; want an HttpRule", ext)
	}

	return opts, nil
}

func (PathObj openapiPathsObject)setOpenapiPathsObject(pathInfo []pathInfo) openapiPathsObject{
	for _,pathInfoItem := range pathInfo {
		PathObj[pathInfoItem.path] = openapiPathItemObject{
			Post:   &openapiOperationObject{
				Summary:      pathInfoItem.remark,
				OperationID:  fmt.Sprintf("%s_%s", pathInfoItem.svc.GetName(), pathInfoItem.method.GetName()),
				Responses:    openapiResponsesObject{
					"200": openapiResponseObject{
						Description: "A successful response.",
						Schema: openapiSchemaObject{
							schemaCore:           schemaCore{
								Ref: fmt.Sprintf("$/definitions/%s", pathInfoItem.method.GetInputType().GetName()),
							},
						},
					},
				},
				Parameters:   getOpenapiParametersObject(pathInfoItem),
				Tags:         []string{pathInfoItem.svc.GetName()},
			} ,
		}
	}

	return nil
}

func getOpenapiParametersObject(info pathInfo)openapiParametersObject{
	var parameters openapiParametersObject
	if info.isAnnotation {
		parameter := openapiParameterObject{
			Name:             "body",
			In: "body",
			Required: true,
			Schema: &openapiSchemaObject{
				schemaCore:           schemaCore{
					Ref: fmt.Sprintf("$/definitions/%s", info.method.GetInputType().GetName()),
				},
			},
		}
		parameters = append(parameters, parameter)
		return parameters
	}


	switch info.httpInfo.requestType {
	case "GET":

		break
	case "POST":
		break
	case "PUT":
		break
	case "PATCH":
		break
	case "DELETE":
		break
	}

	return nil
}