package core

import (
	"github.com/golang/mock/gomock"
	"testing"
)

func TestMyGuestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
}
