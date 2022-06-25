package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "learningGRPC/my_guest_client/grpc/cdp/v1/my_guest"
)

func MyGuestAdd(c *pb.GuestServicesClient, ctx *context.Context) error {
	boolValue, err := (*c).Add(*ctx, &pb.MyGuest{
		Firstname: gofakeit.FirstName(),
		Lastname:  gofakeit.LastName(),
		Email:     gofakeit.Email(),
		RegTime:   timestamppb.Now(),
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		errorCode := status.Code(err)
		if errorCode == codes.InvalidArgument {
			errorStatus := status.Convert(err)
			for _, d := range errorStatus.Details() {
				switch info := d.(type) {
				case *errdetails.BadRequest_FieldViolation:
					return errors.New(fmt.Sprintf("Request Field Invalid: %s", info))
				default:
					return errors.New(fmt.Sprintf("Unexpected error type: %s", info))
				}
			}
		} else {
			return errors.New(fmt.Sprintf("Unhandled error : %s ", errorCode))
		}
		return err
	}
	if !boolValue.Value {
		return errors.New("Insert record failed")
	}
	return nil
}
