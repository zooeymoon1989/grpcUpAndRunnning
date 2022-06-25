package core

import (
	"context"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/profiling/proto"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/gorm"
	"io"
	"learningGRPC/my_guest_server/bootstrap"
	pb "learningGRPC/my_guest_server/grpc/cdp/v1/my_guest"
	"learningGRPC/my_guest_server/models"
	"log"
)

type MessageServer struct {
	proto.UnimplementedProfilingServer
	pb.GuestServicesServer
	gormDB *gorm.DB
}

func (m *MessageServer) StreamAdd(stream pb.GuestServices_StreamAddServer) error {
	var results []models.MyGuest
	find := m.gormDB.Table("MyGuests").Find(&results)
	if find.Error != nil {
		return find.Error
	}
	for _, v := range results {
		err := stream.Send(&pb.MyGuest{
			Id:        v.Id,
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			RegTime:   timestamppb.New(v.RegDate),
		})
		if err != nil {
			return err
		}
	}

	for {
		mg, err := stream.Recv()
		if err == io.EOF {

		}
		if err != nil {
			return err
		}
		var md = &models.MyGuest{
			Firstname: mg.GetFirstname(),
			Lastname:  mg.GetLastname(),
			Email:     mg.GetEmail(),
			RegDate:   mg.GetRegTime().AsTime(),
		}
		rst := m.gormDB.Table("MyGuests").Create(md)
		if rst.Error != nil {
			return rst.Error
		}
		err = stream.Send(mg)
		if err != nil {
			return err
		}
	}
}

func (m *MessageServer) Get(_ *emptypb.Empty, stream pb.GuestServices_GetServer) error {
	var mgs []models.MyGuest
	rst := m.gormDB.Table("MyGuests").Find(&mgs)
	if rst.Error != nil {
		return rst.Error
	}
	for _, v := range mgs {
		err := stream.Send(&pb.MyGuest{
			Id:        v.Id,
			Firstname: v.Firstname,
			Lastname:  v.Lastname,
			Email:     v.Email,
			RegTime:   timestamppb.New(v.RegDate),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MessageServer) Add(ctx context.Context, guest *pb.MyGuest) (*wrapperspb.BoolValue, error) {
	model := models.MyGuest{
		Firstname: guest.GetFirstname(),
		Lastname:  guest.GetLastname(),
		Email:     guest.GetEmail(),
		RegDate:   guest.GetRegTime().AsTime(),
	}
	rst := m.gormDB.Table(model.TableName()).Create(&model)
	if rst.Error != nil {
		return nil, rst.Error
	}
	return wrapperspb.Bool(true), nil
}

func (m *MessageServer) Update(ctx context.Context, guest *pb.MyGuest) (*pb.MyGuest, error) {
	var model = &models.MyGuest{
		Firstname: guest.GetFirstname(),
		Lastname:  guest.GetLastname(),
		Email:     guest.GetEmail(),
		RegDate:   guest.GetRegTime().AsTime(),
	}
	rst := m.gormDB.Table(model.TableName()).Where("id = ?", guest.Id).Updates(model)
	if rst.Error != nil {
		return nil, rst.Error
	}
	return guest, nil
}

func (m *MessageServer) Delete(ctx context.Context, value *wrapperspb.Int64Value) (*wrapperspb.BoolValue, error) {

	if value.Value < 0 {
		log.Printf("Id is invalid! -> Received Id %d\n", value.Value)
		s := status.New(codes.InvalidArgument, "Invalid information received")
		ds, err := s.WithDetails(&errdetails.BadRequest_FieldViolation{
			Field:       "ID",
			Description: fmt.Sprintf("ID RECEIVED IS NOT VALID %d", value.Value),
		})
		if err != nil {
			return wrapperspb.Bool(false), s.Err()
		}
		return wrapperspb.Bool(false), ds.Err()
	}

	rst := m.gormDB.Table("my_guest").Delete(&models.MyGuest{}, value.Value)
	if rst.Error != nil {
		return wrapperspb.Bool(false), rst.Error
	}
	return wrapperspb.Bool(true), nil
}

func NewService() *MessageServer {
	return &MessageServer{
		gormDB: bootstrap.NewDatabase(),
	}
}
