package log_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrOffestOutOfRange struct {
	Offset uint64
}

func (e ErrOffestOutOfRange) GRPCStatus() *status.Status {
	st := status.New(404, fmt.Sprintf("offset out of range: %d", e.Offset))
	msg := fmt.Sprintf("The requested offset is outside of th elogs range: %d", e.Offset)
	details := &errdetails.LocalizedMessage{
		Locale:  "en-GB",
		Message: msg,
	}
	std, err := st.WithDetails(details)
	if err != nil {
		return st
	}
	return std
}
func (e ErrOffestOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}
