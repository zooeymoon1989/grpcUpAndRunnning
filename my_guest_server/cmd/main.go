package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"learningGRPC/my_guest_server/metrics"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"learningGRPC/my_guest_server/config"
	"learningGRPC/my_guest_server/core"
	pb "learningGRPC/my_guest_server/grpc/cdp/v1/my_guest"
	"net"
	"os"
	"strings"
	"time"
)

var (
	log                = logrus.New()
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid credentials")
)

func init() {
	log.Out = os.Stdout
	// 注册gzip压缩
	encoding.RegisterCompressor(encoding.GetCompressor(gzip.Name))
	// 初始化 prometheus
	metrics.Reg.MustRegister(metrics.GrpcMetrics, metrics.CustomMetricCounter)
}

func main() {

	//  ---------
	// 启动zapage 协程
	//go func() {
	//	mux := http.NewServeMux()
	//	zpages.Handle(mux, "/debug")
	//	log.Fatal(http.ListenAndServe("127.0.0.1:8081", mux))
	//}()
	//
	//view.RegisterExporter(&exporter.PrintExporter{}) // 这个是展示用的，生产环境不宜使用
	//
	//if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
	//	log.Fatal(err)
	//}
	//  ----------

	httpServer := &http.Server{
		Handler: promhttp.HandlerFor(metrics.Reg, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", 9092),
	}

	cert, err := tls.LoadX509KeyPair(config.CrtFile, config.KeyFile)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(config.CaFile)
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(ca) {
		panic("failed to append ca certificate")
	}

	opts := []grpc.ServerOption{
		//grpc.StatsHandler(&ocgrpc.ServerHandler{}), // 添加handler
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
		grpc.ChainUnaryInterceptor(
			//ensureValidBasicCredentials,
			metrics.GrpcMetrics.UnaryServerInterceptor(),
			ensureValidOauth2Credentials,
			MyGuestUnaryServerInterceptor),
		grpc.StreamInterceptor(MyGuestStreamServerInterceptor),
	}
	server := grpc.NewServer(
		opts...,
	)
	pb.RegisterGuestServicesServer(server, core.NewService())

	metrics.GrpcMetrics.InitializeMetrics(server)

	// 开始prometheus服务
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server")
		}
	}()

	reflection.Register(server)
	log.Printf("Starting gRPC listener on port " + config.Port)

	listen, err := net.Listen("tcp", config.Port)
	if err != nil {
		panic(err)
	}
	if err := server.Serve(listen); err != nil {
		panic(err)
	}
}

func ensureValidOauth2Credentials(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	if !validAuth2(md["authorization"]) {
		return nil, errInvalidToken
	}
	return handler(ctx, req)

}

func ensureValidBasicCredentials(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errMissingMetadata
	}

	if !valid(md["authorization"]) {
		return nil, errInvalidToken
	}

	return handler(ctx, req)
}

func validAuth2(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	return token == "some-token"
}

func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Basic ")
	return token == base64.StdEncoding.EncodeToString([]byte("admin:admin"))
}

type wrappedStream struct {
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] "+
		"Receive a message (Type: %T) at %s",
		m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Server Stream Interceptor Wrapper] "+
		"Send a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

// MyGuestStreamServerInterceptor 流的拦截器
func MyGuestStreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("======= [Server Stream interceptor]", info.FullMethod)
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}
	return err
}

// MyGuestUnaryServerInterceptor 普通拦截器
func MyGuestUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Println("======= [Server interceptor]", info.FullMethod)
	m, err := handler(ctx, req)

	log.Printf("Post Proc Message : %s \n", m)
	return m, err
}
