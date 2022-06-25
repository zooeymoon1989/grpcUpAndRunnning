package core

import (
	"context"
	"github.com/brianvoe/gofakeit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "learningGRPC/my_guest_server/grpc/cdp/v1/my_guest"
	"log"
	"net"
	"testing"
	"time"
)

// TestMessageServer_Add
// grpc MessageServer Add unit test
func TestMessageServer_Add(t *testing.T) {

	test := &pb.MyGuest{
		Firstname: gofakeit.FirstName(),
		Lastname:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		RegTime:   timestamppb.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	dial, err := grpc.DialContext(
		ctx,
		"localhost:54321",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer()),
	)
	defer dial.Close()
	if err != nil {
		t.Error(err.Error())
	}

	c := pb.NewGuestServicesClient(dial)

	boolValue, err := c.Add(ctx, test, grpc.UseCompressor(gzip.Name))

	if err != nil {
		t.Error(err)
	}

	if !boolValue.Value {
		t.Error("insert record failed")
	}
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterGuestServicesServer(server, NewService())
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
