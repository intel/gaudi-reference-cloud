// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/descriptorpb"
)

const validate = "/pkg/mod/github.com/envoyproxy/protoc-gen-validate@v0.10.1"

type ReadProtoOpts struct {
	IncludeImports bool
	IncludeDir     string
}

func ReadProto(opts *ReadProtoOpts, files ...string) ([]byte, error) {

	if len(files) == 0 {
		return []byte{}, nil
	}
	file, err := os.CreateTemp("", "pb")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
		if err := os.Remove(file.Name()); err != nil {
			panic(err)
		}
	}()
	cmdArgs := []string{
		"--proto_path", filepath.Dir(files[0]), "-o", file.Name(),
		"-I", (os.Getenv("GOPATH") + validate),
	}
	if opts.IncludeImports {
		cmdArgs = append(cmdArgs, "--include_imports")
	}
	cmdArgs = append(cmdArgs, files...)

	cmd := exec.Command("protoc", cmdArgs...)
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("protoc error: %w", err)
	}

	return io.ReadAll(file)
}

func ReadProtoFd(fileName string) (*descriptorpb.FileDescriptorProto, error) {
	buf, err := ReadProto(&ReadProtoOpts{}, fileName)
	if err != nil {
		return nil, err
	}

	desc := descriptorpb.FileDescriptorSet{}
	if err = proto.Unmarshal(buf, &desc); err != nil {
		return nil, fmt.Errorf("unmarshal error %v: %w", fileName, err)
	}

	if len(desc.File) != 1 {
		return nil, fmt.Errorf("%v: expecting 1 file descriptor, got %v", fileName, len(desc.File))
	}

	return desc.File[0], nil
}

func getIdcOptions[R proto.Message](msg proto.Message, ext *protoimpl.ExtensionInfo) R {
	if msg == nil {
		var nilR R
		return nilR
	}
	xt := proto.GetExtension(msg, ext)
	if rr, ok := xt.(R); ok {
		return rr
	}
	var nilR R
	return nilR
}

func GetIdcFileOptions(fd *descriptorpb.FileDescriptorProto) *pb.IdcFileOptions {
	return getIdcOptions[*pb.IdcFileOptions](fd.GetOptions(), pb.E_File)
}

func GetDeploy(fd *descriptorpb.FileDescriptorProto) string {
	fileOpt := GetIdcFileOptions(fd)
	if fileOpt == nil || fileOpt.Deploy == pb.DeploymentType_unspecified {
		return ""
	}
	return fileOpt.Deploy.String()
}

func GetIdcServiceOptions(svc *descriptorpb.ServiceDescriptorProto) *pb.IdcServiceOptions {
	return getIdcOptions[*pb.IdcServiceOptions](svc.GetOptions(), pb.E_Service)
}

func GetIdcMethodOptions(meth *descriptorpb.MethodDescriptorProto) *pb.IdcMethodOptions {
	return getIdcOptions[*pb.IdcMethodOptions](meth.GetOptions(), pb.E_Method)
}

func GetIdcFieldOptions(field *descriptorpb.FieldDescriptorProto) *pb.IdcFieldOptions {
	return getIdcOptions[*pb.IdcFieldOptions](field.GetOptions(), pb.E_Field)
}

func GetMessageDescriptor(msgType string) protoreflect.MessageDescriptor {
	msgType = msgType[1:]
	msg, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(msgType))
	if err != nil {
		log.Fatalf("Can't find type for %v: %v", msgType, err)
	}
	desc := msg.Descriptor()
	if desc == nil {
		log.Fatalf("Can't find descriptor for %v", msgType)
	}
	return desc
}

func findOptFields(desc protoreflect.MessageDescriptor, optName string) []protoreflect.FieldDescriptor {
	fields := desc.Fields()
	var result = []protoreflect.FieldDescriptor{}
	for ii := 0; ii < fields.Len(); ii++ {
		fd := fields.Get(ii)
		if fd.Kind() == protoreflect.MessageKind {
			fds := findOptFields(fd.Message(), optName)
			if len(fds) != 0 && len(result) != 0 {
				log.Fatalf("%v has multiple fields with %v option", desc.Name(), optName)
			}
			if len(fds) != 0 {
				result = append(result, fd)
				result = append(result, fds...)
			}
			continue
		}
		opts := fd.Options()
		if opts == nil {
			continue
		}
		oo, ok := opts.(*descriptorpb.FieldOptions)
		if !ok {
			continue
		}
		ee := proto.GetExtension(oo, pb.E_Field)
		if ee == nil {
			continue
		}
		fo, ok := ee.(*pb.IdcFieldOptions)
		if !ok || fo == nil {
			continue
		}
		foMsg := fo.ProtoReflect()
		foDesc := foMsg.Descriptor()
		optFd := foDesc.Fields().ByName(protoreflect.Name(optName))
		if optFd == nil {
			continue
		}
		val := foMsg.Get(optFd).Interface()
		if ok, set := val.(bool); !ok || !set {
			continue
		}
		if len(result) != 0 {
			log.Fatalf("%v has multiple fields with %v option", desc.Name(), optName)
		}
		result = append(result, fd)
	}
	return result

}

func FindOptField(msgType string, optName string) []protoreflect.FieldDescriptor {
	desc := GetMessageDescriptor(msgType)
	return findOptFields(desc, optName)
}

func JoinFieldNames(fds []protoreflect.FieldDescriptor, namer func(protoreflect.FieldDescriptor) string) string {
	builder := strings.Builder{}
	for _, fd := range fds {
		if builder.Len() > 0 {
			if err := builder.WriteByte('.'); err != nil {
				log.Fatal("Unexpected error observed while writing to builder")
			}
		}
		if _, err := builder.WriteString(namer(fd)); err != nil {
			log.Fatal("Unexpected error observed while writing to builder")
		}
	}
	return builder.String()
}

func FieldName(fd protoreflect.FieldDescriptor) string {
	return Capitalize(string(fd.Name()))
}

func FieldTextName(fd protoreflect.FieldDescriptor) string {
	return string(fd.TextName())
}

func Capitalize(str string) string {
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func Uncapitalize(str string) string {
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
