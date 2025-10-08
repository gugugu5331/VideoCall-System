package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor gRPC 服务端一元拦截器
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		tracer := opentracing.GlobalTracer()

		// 从 metadata 提取 span context
		md, ok := metadata.FromIncomingContext(ctx)
		var spanCtx opentracing.SpanContext
		if ok {
			spanCtx, _ = tracer.Extract(
				opentracing.TextMap,
				metadataTextMap(md),
			)
		}

		// 创建 span
		span := tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanCtx),
		)
		defer span.Finish()

		// 设置标签
		ext.Component.Set(span, "gRPC")
		ext.SpanKindRPCServer.Set(span)

		// 将 span 注入到 context
		ctx = opentracing.ContextWithSpan(ctx, span)

		// 调用处理器
		resp, err := handler(ctx, req)

		// 如果有错误，标记 span
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("error.message", err.Error())
		}

		return resp, err
	}
}

// UnaryClientInterceptor gRPC 客户端一元拦截器
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		tracer := opentracing.GlobalTracer()

		// 从 context 获取父 span
		var parentSpan opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentSpan = parent.Context()
		}

		// 创建 span
		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentSpan),
		)
		defer span.Finish()

		// 设置标签
		ext.Component.Set(span, "gRPC")
		ext.SpanKindRPCClient.Set(span)

		// 将 span context 注入到 metadata
		md := metadata.New(nil)
		if err := tracer.Inject(
			span.Context(),
			opentracing.TextMap,
			metadataTextMap(md),
		); err != nil {
			span.SetTag("error.inject", err.Error())
		}

		// 将 metadata 添加到 context
		ctx = metadata.NewOutgoingContext(ctx, md)

		// 调用 RPC
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 如果有错误，标记 span
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("error.message", err.Error())
		}

		return err
	}
}

// StreamServerInterceptor gRPC 服务端流式拦截器
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		tracer := opentracing.GlobalTracer()

		// 从 metadata 提取 span context
		md, ok := metadata.FromIncomingContext(ss.Context())
		var spanCtx opentracing.SpanContext
		if ok {
			spanCtx, _ = tracer.Extract(
				opentracing.TextMap,
				metadataTextMap(md),
			)
		}

		// 创建 span
		span := tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanCtx),
		)
		defer span.Finish()

		// 设置标签
		ext.Component.Set(span, "gRPC")
		ext.SpanKindRPCServer.Set(span)
		span.SetTag("grpc.stream", true)

		// 创建包装的 ServerStream
		wrappedStream := &tracedServerStream{
			ServerStream: ss,
			span:         span,
		}

		// 调用处理器
		err := handler(srv, wrappedStream)

		// 如果有错误，标记 span
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("error.message", err.Error())
		}

		return err
	}
}

// StreamClientInterceptor gRPC 客户端流式拦截器
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		tracer := opentracing.GlobalTracer()

		// 从 context 获取父 span
		var parentSpan opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentSpan = parent.Context()
		}

		// 创建 span
		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentSpan),
		)

		// 设置标签
		ext.Component.Set(span, "gRPC")
		ext.SpanKindRPCClient.Set(span)
		span.SetTag("grpc.stream", true)

		// 将 span context 注入到 metadata
		md := metadata.New(nil)
		if err := tracer.Inject(
			span.Context(),
			opentracing.TextMap,
			metadataTextMap(md),
		); err != nil {
			span.SetTag("error.inject", err.Error())
		}

		// 将 metadata 添加到 context
		ctx = metadata.NewOutgoingContext(ctx, md)

		// 创建流
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("error.message", err.Error())
			span.Finish()
			return nil, err
		}

		// 返回包装的 ClientStream
		return &tracedClientStream{
			ClientStream: clientStream,
			span:         span,
		}, nil
	}
}

// tracedServerStream 包装 ServerStream 以支持追踪
type tracedServerStream struct {
	grpc.ServerStream
	span opentracing.Span
}

func (s *tracedServerStream) Context() context.Context {
	return opentracing.ContextWithSpan(s.ServerStream.Context(), s.span)
}

// tracedClientStream 包装 ClientStream 以支持追踪
type tracedClientStream struct {
	grpc.ClientStream
	span opentracing.Span
}

func (s *tracedClientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	if err != nil {
		ext.Error.Set(s.span, true)
		s.span.SetTag("error.message", err.Error())
	}
	s.span.Finish()
	return err
}

// metadataTextMap 实现 opentracing.TextMapReader 和 opentracing.TextMapWriter
type metadataTextMap metadata.MD

func (m metadataTextMap) Set(key, val string) {
	metadata.MD(m).Set(key, val)
}

func (m metadataTextMap) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range metadata.MD(m) {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
