package server

import (
	"net/http"

	"github.com/generalledger/response"
)

func (s *Server) ping() http.HandlerFunc {
	type pingResponse struct {
		Version string `json:"version"`
		DbConn  string `json:"database_connection"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w)
		defer resp.Output()
		responseStatus := http.StatusOK
		pingResp := pingResponse{
			Version: s.Version,
			DbConn:  "OK",
		}
		err := s.DB.Ping()
		if err != nil {
			responseStatus = http.StatusInternalServerError
			pingResp.DbConn = err.Error()
		}
		resp.SetResult(responseStatus, pingResp)
	}
}
