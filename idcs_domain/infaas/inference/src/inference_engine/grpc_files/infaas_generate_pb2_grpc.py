# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

import src.inference_engine.grpc_files.infaas_generate_pb2 as infaas__generate__pb2

GRPC_GENERATED_VERSION = '1.65.1'
GRPC_VERSION = grpc.__version__
EXPECTED_ERROR_RELEASE = '1.66.0'
SCHEDULED_RELEASE_DATE = 'August 6, 2024'
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    warnings.warn(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in infaas_generate_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
        + f' This warning will become an error in {EXPECTED_ERROR_RELEASE},'
        + f' scheduled for release on {SCHEDULED_RELEASE_DATE}.',
        RuntimeWarning
    )


class TextGeneratorStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GenerateStream = channel.unary_stream(
                '/proto.TextGenerator/GenerateStream',
                request_serializer=infaas__generate__pb2.GenerateStreamRequest.SerializeToString,
                response_deserializer=infaas__generate__pb2.GenerateStreamResponse.FromString,
                _registered_method=True)
        self.ChatCompletionStream = channel.unary_stream(
                '/proto.TextGenerator/ChatCompletionStream',
                request_serializer=infaas__generate__pb2.ChatCompletionStreamRequest.SerializeToString,
                response_deserializer=infaas__generate__pb2.ChatCompletionStreamResponse.FromString,
                _registered_method=True)


class TextGeneratorServicer(object):
    """Missing associated documentation comment in .proto file."""

    def GenerateStream(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ChatCompletionStream(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_TextGeneratorServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GenerateStream': grpc.unary_stream_rpc_method_handler(
                    servicer.GenerateStream,
                    request_deserializer=infaas__generate__pb2.GenerateStreamRequest.FromString,
                    response_serializer=infaas__generate__pb2.GenerateStreamResponse.SerializeToString,
            ),
            'ChatCompletionStream': grpc.unary_stream_rpc_method_handler(
                    servicer.ChatCompletionStream,
                    request_deserializer=infaas__generate__pb2.ChatCompletionStreamRequest.FromString,
                    response_serializer=infaas__generate__pb2.ChatCompletionStreamResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'proto.TextGenerator', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('proto.TextGenerator', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class TextGenerator(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def GenerateStream(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(
            request,
            target,
            '/proto.TextGenerator/GenerateStream',
            infaas__generate__pb2.GenerateStreamRequest.SerializeToString,
            infaas__generate__pb2.GenerateStreamResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)

    @staticmethod
    def ChatCompletionStream(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_stream(
            request,
            target,
            '/proto.TextGenerator/ChatCompletionStream',
            infaas__generate__pb2.ChatCompletionStreamRequest.SerializeToString,
            infaas__generate__pb2.ChatCompletionStreamResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
