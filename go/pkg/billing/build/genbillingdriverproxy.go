package main

// This tool parses billing.proto and creates proxies for all services
// that are marked with the option idc.service.proxyBillingDriver = true.
// This means that after changes to billing.proto, "go generate" can be used
// to update the billing service without having to hand-write code.
// It also means that if we want to change the way the proxy code works,
// the change only needs to be made in the template in this file rather
// than in many nearly identical functions.

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/build/tools/util"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Method struct {
	Name           string
	Streams        bool
	InputType      string
	OutputType     string
	CloudAccountId string
}

type Service struct {
	Name            string
	DriverFieldName string
	ReceiverType    string
	Client          string
	Methods         []*Method
}

type TemplateData struct {
	Services []*Service
	Imports  map[string]any
}

var (
	templateData = TemplateData{Imports: make(map[string]any)}
	outputFile   string
	formatOutput bool
	includeDir   string
)

func init() {
	flag.StringVar(&outputFile, "output-file", "", "output file")
	flag.BoolVar(&formatOutput, "format", false, "format output")
	flag.StringVar(&includeDir, "I", "", "include dir")
}

func main() {
	flag.Parse()
	if outputFile == "" {
		log.Fatal("--output-file is required")
	}
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("usage: %v --ouput-file out path-to-billing.proto", os.Args[0])
	}
	inFileName := args[0]
	file, err := os.CreateTemp("", "pb")
	if err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(file.Name())
	cmd := exec.Command("protoc", "--proto_path",
		filepath.Dir(inFileName), "-o", file.Name(), inFileName, "-I", includeDir)
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		log.Fatalf("protoc error: %v", err)
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("read %v: %v", file.Name(), err)
	}

	desc := descriptorpb.FileDescriptorSet{}
	if err = proto.Unmarshal(buf, &desc); err != nil {
		log.Fatalf("unmarshal error %v: %v", file.Name(), err)
	}

	for _, ff := range desc.File {
		for _, service := range ff.Service {
			if oo := service.Options; oo != nil {
				ee := proto.GetExtension(oo, pb.E_Service)
				fo, ok := ee.(*pb.IdcServiceOptions)
				if ok && fo.ProxyBillingDriver {
					findProxies(service)
				} else {
					log.Println("idcserviceoptions assertion or proxybillingrriver failed")
				}
			}
			service.GetOptions().ProtoReflect().GetUnknown()
		}
	}

	tmpl, err := template.New("billing").Parse(templateCode)
	if err != nil {
		log.Fatalf("error parsing internal template: %v", err)
	}
	out := bytes.Buffer{}
	if err = tmpl.Execute(&out, &templateData); err != nil {
		log.Fatalf("error executing template: %v", err)
	}

	var outBytes []byte
	if formatOutput {
		outBytes, err = format.Source(out.Bytes())
		if err != nil {
			log.Fatalf("error formatting template output: %v", err)
		}
	} else {
		outBytes = out.Bytes()
	}

	tmpFile := outputFile + ".tmp"
	outf, err := os.OpenFile(tmpFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("open %v: %v", tmpFile, err)
	}
	if _, err = outf.Write(outBytes); err != nil {
		log.Fatalf("write %v: %v", tmpFile, err)
	}
	outf.Close()
	if err = os.Rename(tmpFile, outputFile); err != nil {
		log.Fatalf("rename %v %v: %v", tmpFile, outputFile, err)
	}
}

func pbType(protoType string) string {
	ii := strings.LastIndex(protoType, ".")
	pkg := protoType[:ii]
	typ := protoType[ii+1:]
	if pkg == ".proto" {
		return "*pb." + typ
	}
	if pkg == ".google.protobuf" {
		switch typ {
		case "Empty":
			templateData.Imports[`"github.com/golang/protobuf/ptypes/empty"`] = nil
			return "*empty.Empty"
		case "Timestamp":
			templateData.Imports[`"github.com/golang/protobuf/ptypes/timestamp"`] = nil
			return "*timestamp.Timestamp"
		}
	}
	log.Fatalf("Don't recognize type %v", protoType)
	return ""
}

func findProxies(svc *descriptorpb.ServiceDescriptorProto) {
	service := Service{
		Name:         svc.GetName(),
		ReceiverType: fmt.Sprintf("%vProxy", svc.GetName()),
		Client:       util.Uncapitalize(strings.Replace(svc.GetName(), "Service", "", 1)),
	}
	for _, meth := range svc.Method {
		method := Method{
			Name:       meth.GetName(),
			Streams:    meth.GetServerStreaming(),
			InputType:  pbType(meth.GetInputType()),
			OutputType: pbType(meth.GetOutputType()),
		}

		cloudAcctFds := util.FindOptField(meth.GetInputType(), "cloudAccount")
		if len(cloudAcctFds) == 0 {
			log.Fatalf("no cloudAccountId in %v", meth.GetInputType())
		}
		method.CloudAccountId = util.JoinFieldNames(cloudAcctFds, util.FieldName)
		service.Methods = append(service.Methods, &method)
	}
	templateData.Services = append(templateData.Services, &service)
}

var templateCode = `// Code generated by genbillingdriverproxy. DO NOT EDIT.
package server

import (
	"context"
	"errors"
	"io"
	grpc "google.golang.org/grpc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
{{- range $imp, $val := .Imports}}
	{{$imp}}
)
{{- end}}

type BillingDriver struct {
	name string
	conn *grpc.ClientConn
{{- range .Services}}
	{{.Client}} pb.{{.Name}}Client
{{- end}}
}

func NewBillingDriver(name string, conn *grpc.ClientConn) *BillingDriver {
	return &BillingDriver{
		name: name,
		conn: conn,
{{- range .Services}}
	    {{.Client}}: pb.New{{.Name}}Client(conn),
{{- end}}
	}
}

func RegisterProxies(grpcServer *grpc.Server) {
{{- range .Services}}
	pb.Register{{.Name}}Server(grpcServer, &{{.ReceiverType}}{})
{{- end}}
}

{{- range .Services}}
{{-   $svc := .}}

type {{$svc.ReceiverType}} struct{
	pb.Unimplemented{{$svc.Name}}Server
}
{{-   range .Methods}}
{{      if .Streams}}
func (_ *{{$svc.ReceiverType}}) {{.Name}}(in {{.InputType}}, outStream pb.{{- $svc.Name}}_{{.Name}}Server) error {
	ctx := outStream.Context()
{{-     else}}
func (_ *{{$svc.ReceiverType}}) {{.Name}}(ctx context.Context, in {{.InputType}}) ({{.OutputType}}, error) {
{{-     end}}
	logger := log.FromContext(ctx).WithName("{{$svc.ReceiverType}}")
	driver, err := GetDriver(ctx, in.Get{{.CloudAccountId}}())
	if err != nil {
		logger.Error(err, "unable to find driver", "cloudAccountId", in.Get{{.CloudAccountId}}())
{{-     if .Streams}}
		return err
{{-     else}}
		return nil, err
{{-     end}}
	}
	res, err := driver.{{$svc.Client}}.{{.Name}}(ctx, in)
	if err != nil {
		logger.Error(err, "calling {{.Name}}")
{{-     if .Streams}}
		return err
{{-     else}}
		return nil, err
{{-     end}}
	}
{{-     if .Streams}}
	for {
		out, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			logger.Error(err, "recv error in {{.Name}}")
			return err
		}
		send_err := outStream.Send(out)
		if send_err != nil {
			logger.Error(send_err, "send error in {{.Name}}")
			return send_err
		}
	}
{{-     else}}
	return res, err
{{-     end}}
}
{{-   end}}
{{- end}}
`
