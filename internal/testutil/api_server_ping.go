package testutil

import (
	"context"
	"net/http"
)

func (s *ApiServer) handlePing(ctx context.Context, resp http.ResponseWriter, req *http.Request) error {
	resp.WriteHeader(http.StatusOK)
	_, err := resp.Write([]byte("OK"))

	return err
}
