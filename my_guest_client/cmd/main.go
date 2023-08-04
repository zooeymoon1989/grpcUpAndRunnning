package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/brianvoe/gofakeit"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"golang.org/x/oauth2"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"learningGRPC/my_guest_client/core"
	pb "learningGRPC/my_guest_client/grpc/cdp/v1/my_guest"
	"log"
	"net/http"
	"os"
	"time"
)

const address = "localhost:54321"
const hostname = "localhost"
const crtFile = "ca/client.crt"
const keyFile = "ca/client.key"
const caFile = "ca/ca.crt"

func init() {
	view.RegisterExporter(&exporter.PrintExporter{})

	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		panic(err)
	}
}

func main() {

	reg := prometheus.NewRegistry()
	grpcMetrics := grpc_prometheus.NewClientMetrics()
	reg.MustRegister(grpcMetrics)

	auth := oauth.NewOauthAccess(fetchToken())
	cert, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		panic(err)
	}
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(caFile)
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(ca) {
		panic("failed to append ca certificate")
	}

	if err != nil {
		panic(err)
	}

	opts := []grpc.DialOption{
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:         hostname,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            certPool,
			InsecureSkipVerify: true,
		})),
		//grpc.WithPerRPCCredentials(security.NewBasicAuth("admin", "admin")), //basic auth
		grpc.WithPerRPCCredentials(auth), //oauth 2.0
		grpc.WithUnaryInterceptor(MyGuestsUnaryClientInterceptor),
		grpc.WithStreamInterceptor(MyGuestsStreamClientInterceptor),
	}
	// 普通grpc连接
	conn, err := grpc.Dial(
		address,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := pb.NewGuestServicesClient(conn)

	httpSerer := &http.Server{
		Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", 9094),
	}

	go func() {
		if err := httpSerer.ListenAndServe(); err != nil {
			log.Fatal("Unable to start a http server.")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second * 5))
	defer cancel()
	err = core.MyGuestAdd(&c, &ctx)
	if err != nil {
		println(err)
	}
	boolValue, err := c.Add(ctx, &pb.MyGuest{
		Firstname: gofakeit.FirstName(),
		Lastname:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		RegTime:   timestamppb.Now(),
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		println(err.Error())
		errorCode := status.Code(err)
		if errorCode == codes.InvalidArgument {
			log.Printf("Invalid argument Error : %s", errorCode)
			errorStatus := status.Convert(err)
			for _, d := range errorStatus.Details() {
				switch info := d.(type) {
				case *errdetails.BadRequest_FieldViolation:
					log.Printf("Request Field Invalid: %s", info)
				default:
					log.Printf("Unexpected error type: %s", info)
				}
			}
		} else {
			log.Printf("Unhandled error : %s ", errorCode)
		}
		return
	}
	if boolValue.Value {
		println("yes")
	} else {
		println("no")
	}

	//getStream, err := c.Get(ctx, &emptypb.Empty{})
	//defer getStream.CloseSend()
	//if err != nil {
	//	panic(err)
	//}
	//for {
	//	myGuest, err := getStream.Recv()
	//	if err == io.EOF {
	//		break
	//	}
	//
	//	if err != nil {
	//		break
	//	}
	//
	//	println(myGuest.GetId())
	//	println(myGuest.GetFirstname())
	//	println(myGuest.GetLastname())
	//	println(myGuest.GetEmail())
	//	println(myGuest.GetRegTime())
	//	println("----------")
	//}

	//stream, err := c.StreamAdd(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//for i := 0; i < 10; i++ {
	//	if err := stream.Send(&pb.MyGuest{
	//		Firstname: gofakeit.FirstName(),
	//		Lastname:  gofakeit.LastName(),
	//		Email:     gofakeit.Email(),
	//		RegTime:   timestamppb.Now(),
	//	}); err != nil {
	//		panic(err)
	//	}
	//}
	//
	//myGuestCh := make(chan *pb.MyGuest)
	//go func(client pb.GuestServices_StreamAddClient, ch chan *pb.MyGuest) {
	//	defer close(myGuestCh)
	//	for {
	//		stream, err := client.Recv()
	//		if err == io.EOF {
	//			break
	//		}
	//		if err != nil {
	//			break
	//		}
	//		ch <- stream
	//	}
	//}(stream, myGuestCh)
	//
	//for m := range myGuestCh {
	//	println(m.GetId())
	//	println(m.GetFirstname())
	//	println(m.GetLastname())
	//	println(m.GetEmail())
	//	println(m.GetRegTime().AsTime().String())
	//	println("===============")
	//}

	println("done")
}

func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "some-token",
	}
}

func MyGuestsStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("======= [Client Interceptor] ", method)

	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}

type wrappedStream struct {
	grpc.ClientStream
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Send a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Client Stream Interceptor] "+
		"Receive a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}

func MyGuestsUnaryClientInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Println("Method : " + method)
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println(reply)
	return err
}
