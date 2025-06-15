package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/platon-p/kpodz3/payments/application/services"
)

type HttpServer struct {
	Port           int
	AccountService services.AccountService

	engine *gin.Engine
}

func NewHTTPServer(port int, accountService services.AccountService) *HttpServer {
	return &HttpServer{
		Port:           port,
		AccountService: accountService,

		engine: nil,
	}
}

func (s *HttpServer) Setup() error {
	s.engine = gin.Default()
	s.engine.Group("/accounts/:userId").
		POST("/create", s.createAccount).
		POST("/topup", s.topUp).
		GET("/balance", s.getBalance)

	return nil
}

func (s *HttpServer) Serve() error {
	return s.engine.Run(fmt.Sprintf(":%d", s.Port))
}

func (s *HttpServer) createAccount(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = s.AccountService.CreateAccount(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (s *HttpServer) topUp(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	amountStr := c.Query("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = s.AccountService.TopUp(c.Request.Context(), userId, amount)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)

}
func (s *HttpServer) getBalance(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	balance, err := s.AccountService.GetBalance(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, balance)
}

func getUserId(c *gin.Context) (int, error) {
	userIdStr := c.Param("userId")
	return strconv.Atoi(userIdStr)
}
