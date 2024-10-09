package main

import (
	"bytes"
	_ "embed"
	"flag"
	"html/template"
	"strings"

	"github.com/Masterminds/sprig"
	"github.com/huandu/xstrings"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

//go:embed sparrow.gotmpl
var sparrowTmp []byte

func main() {
	var flags flag.FlagSet
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(
		func(plugin *protogen.Plugin) error {
			plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
			return codegen(plugin)
		},
	)
}

func codegen(plugin *protogen.Plugin) error {
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}
		filename := file.GeneratedFilenamePrefix + ".srpc.go"
		g := plugin.NewGeneratedFile(filename, file.GoImportPath)

		tpl, err := template.New("sparrow").Funcs(FuncMap()).Parse(string(sparrowTmp))
		if err != nil {
			return err
		}

		var data = &FileProto{
			FileDescriptorProto: file.Proto,
			Attrs:               make(Attrs),
		}

		for _, s := range file.Proto.Service {
			service := &Service{
				ServiceDescriptorProto: s,
				Attrs:                  make(Attrs),
			}
			for _, m := range s.GetMethod() {
				method := &Method{
					MethodDescriptorProto: m,
					Attrs:                 make(Attrs),
				}

				if m.GetClientStreaming() && m.GetServerStreaming() {
					method.CallType = "BidiStream"
				} else if m.GetServerStreaming() {
					method.CallType = "ServerStream"
				} else if m.GetClientStreaming() {
					method.CallType = "ClientStream"
				} else {
					method.CallType = "Unary"
				}

				if len(m.GetInputType()) > 0 {
					lastIdx := strings.LastIndexByte(m.GetInputType(), '.')
					if lastIdx >= 0 {
						inputType := m.GetInputType()[lastIdx+1:]
						method.InputTypeName = inputType
					}
				}
				if len(m.GetOutputType()) > 0 {
					lastIdx := strings.LastIndexByte(m.GetOutputType(), '.')
					if lastIdx >= 0 {
						outType := m.GetOutputType()[lastIdx+1:]
						method.OutputTypeName = outType
					}
				}

				service.Method = append(service.Method, method)
			}
			data.Service = append(data.Service, service)
		}

		var buffer bytes.Buffer
		err = tpl.Execute(&buffer, data)
		if err != nil {
			return err
		}

		_, err = g.Write(buffer.Bytes())
		return err
	}
	return nil
}

func FuncMap() template.FuncMap {
	fm := sprig.FuncMap()
	fm["camelcase"] = xstrings.ToPascalCase
	return fm
}

type Attrs map[string]any

type FileProto struct {
	*descriptorpb.FileDescriptorProto
	Attrs
	Service []*Service
}

type Service struct {
	*descriptorpb.ServiceDescriptorProto
	Attrs
	Method []*Method
}

type Method struct {
	*descriptorpb.MethodDescriptorProto
	Attrs
	InputTypeName  string
	OutputTypeName string
	CallType       string
}
