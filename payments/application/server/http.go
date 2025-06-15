package server

import (
	"github.com/gin-gonic/gin"
	"github.com/platon-p/kpodz3/payments/application/services"
)

type HttpServer struct {
	Port           int
	AccountService services.AccountService
}

func NewHTTPServer(port int, accountService services.AccountService) *HttpServer {
	return &HttpServer{
		Port:           port,
		AccountService: accountService,
	}
}

func (s *HttpServer) Start() error {
	e := gin.Default()
	e.Group("/accounts/:userId").
		POST("/create", s.createAccount).
		POST("/topup", s.topUp).
		GET("/balance", s.getBalance)

	return nil
}

func (s *HttpServer) createAccount(c *gin.Context) {}
func (s *HttpServer) topUp(c *gin.Context)         {}
func (s *HttpServer) getBalance(c *gin.Context)    {}
