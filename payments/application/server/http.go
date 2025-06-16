package server

import (
	"errors"
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

func (s *HttpServer) Setup() {
	s.engine = gin.Default()
	s.engine.Group("/:userId").
		GET("/", s.getBalance).
		POST("/", s.createAccount).
		PUT("/", s.topUp)
}

func (s *HttpServer) Serve() error {
	return s.engine.Run(fmt.Sprintf(":%d", s.Port))
}

func (s *HttpServer) createAccount(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid user ID")
		return
	}
	err = s.AccountService.CreateAccount(c.Request.Context(), userId)
	if errors.Is(err, services.ErrAccountAlreadyExists) {
		c.JSON(http.StatusConflict, "account already exists")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *HttpServer) topUp(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid user ID")
		return
	}

	amountStr := c.Query("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid amount")
		return
	}

	err = s.AccountService.TopUp(c.Request.Context(), userId, amount)
	if errors.Is(err, services.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (s *HttpServer) getBalance(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid user ID")
		return
	}

	balance, err := s.AccountService.GetBalance(c.Request.Context(), userId)
	if errors.Is(err, services.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, balance)
}

func getUserId(c *gin.Context) (int, error) {
	userIdStr := c.Param("userId")
	return strconv.Atoi(userIdStr)
}
