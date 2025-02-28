// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package machine_image

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"sigs.k8s.io/yaml"
)

// GitToGrpcSynchronizer is intended to maintain a database table that can be modified using GRPC.
// The source of truth is a directory (intended to be maintained by Git) that contains YAML files.
// The YAML files must be be directly serializable to Protobuf messages that will be sent to the GRPC methods.
// Users can provide GRPC methods for Search, Create, Update, and Delete.
type GitToGrpcSynchronizer struct {
	// The file system containing YAML files.
	// All *.yaml files directly in Fsys will be applied (recursion will not occur).
	Fsys fs.FS
	// A connection to the GRPC server.
	ClientConn *grpc.ClientConn
	// An empty Protobuf message of the type that will be synchronized.
	EmptyMessage proto.Message
	// An empty Protobuf message of the type that will be returned by SearchMethod.
	EmptySearchMethodResponseMessage proto.Message
	// The name of the search method. This method must return a stream of the same type as EmptyMessage.
	// If empty, the CreateMethod will be called for each YAML file.
	SearchMethod string
	// The name of the create method. This method must accept the same type as EmptyMessage.
	CreateMethod string
	// The name of the delete method. This method must accept the same type returned by MessageToDeleteRequestFunc.
	// Ignored if SearchMethod is empty.
	DeleteMethod string
	// The name of the update method. This method must accept the same type returned by MessageToUpdateRequestFunc.
	// Ignored if SearchMethod is empty.
	UpdateMethod string
	// A function that accepts a Protobuf message and returns an object that represents its unique key. Returned type
	// has to be comparable.
	// Must handle Protobuf messages of type EmptyMessage and EmptySearchMethodResponseMessage.
	// Ignored if SearchMethod is empty.
	MessageKeyFunc func(proto.Message) (any, error)
	// A function that accepts a Protobuf message and returns a Protobuf message that can be sent to DeleteMethod.
	// Ignored if SearchMethod is empty.
	MessageToDeleteRequestFunc func(proto.Message) (proto.Message, error)
	// A function that accepts a Protobuf message deserialized from a file and a message with matching key found on the
	// server. Returns a Protobuf message that can be sent to UpdateMethod.
	// Ignored if SearchMethod is empty.
	MessageToUpdateRequestFunc func(fileMessage proto.Message, serverMessage proto.Message) (proto.Message, error)
	// A function that checks if a Protobuf message deserialized from a file is equal to a Protobuf message with matching
	// key returned by the gRPC server. If not specified, proto.Equal() function will be used for comparison.
	// Ignored if SearchMethod is empty.
	MessageComparatorFunc func(fileMessage proto.Message, serverMessage proto.Message) (bool, error)
}

// Perform synchronization.
// Returns (true, nil) if changes were made.
// Returns (false, nil) if no changes were needed.
func (g *GitToGrpcSynchronizer) Synchronize(ctx context.Context) (bool, error) {
	g.init(ctx)
	log := log.FromContext(ctx).WithName("GitToGrpcSynchronizer.Synchronize")
	fileMessages, err := g.listFileMessages(ctx)
	if err != nil {
		return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
	}
	serverMessages, err := g.listServerMessages(ctx)
	if err != nil {
		return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
	}
	log.V(1).Info("List results", "serverMessages", fmt.Sprintf("%#v", serverMessages), "fileMessages", fmt.Sprintf("%#v", fileMessages))
	// Determine messages that need to be Created or Updated.
	createMessages := make(map[any]proto.Message)
	updateMessages := make(map[any]proto.Message)
	for key, fileMessage := range fileMessages {
		serverMessage, exists := serverMessages[key]
		if exists {
			equals, err := g.MessageComparatorFunc(fileMessage, serverMessage)
			if err != nil {
				return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
			}
			if equals {
				log.Info("Unmodified", "key", key)
			} else {
				log.Info("Modified", "key", key)
				updateMessage, err := g.MessageToUpdateRequestFunc(fileMessage, serverMessage)
				if err != nil {
					return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
				}
				updateMessages[key] = updateMessage
			}
			// Delete from serverMessages so we can identify extra messages.
			delete(serverMessages, key)
		} else {
			log.Info("New", "key", key)
			createMessages[key] = fileMessage
		}
	}
	// Determine messages that need to be Deleted.
	deleteMessages := make(map[any]proto.Message)
	for key, serverMessage := range serverMessages {
		deleteMessage, err := g.MessageToDeleteRequestFunc(serverMessage)
		if err != nil {
			return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
		}
		log.Info("Extra (will be deleted from server)", "key", key)
		deleteMessages[key] = deleteMessage
	}
	log.V(1).Info("Actions",
		"createMessages", fmt.Sprintf("%#v", createMessages),
		"updateMessages", fmt.Sprintf("%#v", updateMessages),
		"deleteMessages", fmt.Sprintf("%#v", deleteMessages))

	invocations := []struct {
		Method   string
		Messages map[any]proto.Message
	}{
		{Method: g.DeleteMethod, Messages: deleteMessages},
		{Method: g.CreateMethod, Messages: createMessages},
		{Method: g.UpdateMethod, Messages: updateMessages},
	}
	for _, invoc := range invocations {
		log.V(1).Info("Calling API", "Method", invoc.Method, "Messages", fmt.Sprintf("%#v", invoc.Messages))
		if err := g.sendMessages(ctx, invoc.Method, invoc.Messages); err != nil {
			return false, fmt.Errorf("GitToGrpcSynchronizer.Synchronize: %w", err)
		}
	}
	changed := len(createMessages)+len(updateMessages)+len(deleteMessages) > 0
	return changed, nil
}

func (g *GitToGrpcSynchronizer) init(ctx context.Context) {
	// If custom comparator func wasn't provided, use proto.Equal() as a default comparator
	if g.MessageComparatorFunc == nil {
		g.MessageComparatorFunc = func(fileMessage proto.Message, serverMessage proto.Message) (bool, error) {
			return proto.Equal(fileMessage, serverMessage), nil
		}
	}
}

func (g *GitToGrpcSynchronizer) listFileMessages(ctx context.Context) (map[any]proto.Message, error) {
	log := log.FromContext(ctx).WithName("GitToGrpcSynchronizer.listFileMessages")
	// Glob all YAML files in "source"
	filenames, err := fs.Glob(g.Fsys, "*.yaml")
	if err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: Glob: %w", err)
	}

	// Glob all YAML files in "source" subdirectories
	subdirFiles, err := fs.Glob(g.Fsys, "*/*.yaml")
	if err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: Glob: %w", err)
	}
	// Merge the two slices of filenames
	filenames = append(filenames, subdirFiles...)

	log.V(1).Info("Glob", "Fsys", g.Fsys, "filenames", filenames)
	if len(filenames) == 0 {
		// It is possible that the user deleted all YAML files and wants a blank table but this is unlikely.
		// The most likely cause is an incorrect directory so we error on the safe side and stop.
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: no files found")
	}
	messages := make(map[any]proto.Message)
	for _, filename := range filenames {
		log.V(1).Info("Read", "filename", filename)
		reader, err := g.Fsys.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: Open: %w", err)
		}
		defer reader.Close()
		message, err := g.readFileMessage(ctx, reader)
		if err != nil {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: %w", err)
		}
		if err := reader.Close(); err != nil {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: Close: %w", err)
		}
		var key any
		if g.MessageKeyFunc == nil {
			key = filename
		} else {
			key, err = g.MessageKeyFunc(message)
			if err != nil {
				return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: MessageKeyFunc: %w", err)
			}
		}
		_, exists := messages[key]
		if exists {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listFileMessages: more than one file has key %s: %s", key, filename)
		}
		messages[key] = message
	}
	return messages, nil
}

func (g *GitToGrpcSynchronizer) readFileMessage(ctx context.Context, reader io.Reader) (proto.Message, error) {
	log := log.FromContext(ctx).WithName("GitToGrpcSynchronizer.readFileMessage")

	// Convert YAML file to JSON.
	yamlBuffer := new(bytes.Buffer)

	if _, err := yamlBuffer.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.readFileMessage: ReadFrom: %w", err)
	}
	jsonBytes, err := yaml.YAMLToJSON(yamlBuffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.readFileMessage: YAMLToJSON: %w", err)
	}
	log.V(1).Info("YAMLToJSON", "jsonBytes", string(jsonBytes))

	// Deserialize JSON to Protobuf message.
	marshaler := &runtime.JSONPb{}
	marshaler.DiscardUnknown = true
	message := proto.Clone(g.EmptyMessage)
	err = marshaler.Unmarshal(jsonBytes, message)
	if err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.readFileMessage: Unmarshal: %w", err)
	}
	log.V(1).Info("Unmarshal", "message", message)
	return message, nil
}

func (g *GitToGrpcSynchronizer) listServerMessages(ctx context.Context) (map[any]proto.Message, error) {
	log := log.FromContext(ctx).WithName("GitToGrpcSynchronizer.listServerMessages")
	messages := make(map[any]proto.Message)
	if g.SearchMethod == "" {
		// If no SearchMethod was provided, return an empty map.
		// This will cause all files to be Created and there will be no Deletes.
		return messages, nil
	}
	streamDesc := &grpc.StreamDesc{
		ServerStreams: true,
	}
	stream, err := g.ClientConn.NewStream(ctx, streamDesc, g.SearchMethod)
	if err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: NewStream: %w", err)
	}
	req := &emptypb.Empty{}
	if err := stream.SendMsg(req); err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: SendMsg: %w", err)
	}
	if err := stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: CloseSend: %w", err)
	}
	for {
		message := proto.Clone(g.EmptySearchMethodResponseMessage)
		if err := stream.RecvMsg(message); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: RecvMsg: %w", err)
		}
		log.V(1).Info("Search", "message", message)
		key, err := g.MessageKeyFunc(message)
		if err != nil {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: MessageKeyFunc: %w", err)
		}
		_, exists := messages[key]
		if exists {
			return nil, fmt.Errorf("GitToGrpcSynchronizer.listServerMessages: more than one server message has key %s", key)
		}
		messages[key] = message
	}
	return messages, nil
}

func (g *GitToGrpcSynchronizer) sendMessages(ctx context.Context, method string, messages map[any]proto.Message) error {
	for key, message := range messages {
		reply := &emptypb.Empty{}
		if err := g.ClientConn.Invoke(ctx, method, message, reply); err != nil {
			return fmt.Errorf("GitToGrpcSynchronizer.sendMessages: %s: %w", key, err)
		}
	}
	return nil
}
