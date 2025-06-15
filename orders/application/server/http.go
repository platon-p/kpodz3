package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/platon-p/kpodz3/orders/application/services"
)

type HTTPServer struct {
	Port         int
	OrderService services.OrderService

	engine *gin.Engine
}

func NewHTTPServer(port int, orderService services.OrderService) *HTTPServer {
	return &HTTPServer{
		Port:         port,
		OrderService: orderService,

		engine: nil,
	}
}

func (s *HTTPServer) Setup() {
	s.engine = gin.Default()

	s.engine.Group("/").
		POST("", s.createOrder).
		GET("", s.getAllOrders).
		GET("/:name", s.getOrder)
}

func (s *HTTPServer) Run() error {
	return s.engine.Run(fmt.Sprintf(":%d", s.Port))
}

func (s *HTTPServer) getAllOrders(c *gin.Context) {
	orders, err := s.OrderService.GetAllOrders(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, orders)
}

func (s *HTTPServer) createOrder(c *gin.Context) {
	var r struct {
		UserId int    `json:"user_id" binding:"required"`
		Name   string `json:"name" binding:"required"`
		Amount int    `json:"amount" binding:"required"`
	}
	if err := c.BindJSON(&r); err != nil {
		c.AbortWithStatus(400)
		return
	}

	order, err := s.OrderService.CreateOrder(c.Request.Context(), r.UserId, r.Name, r.Amount)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.JSON(201, order)
}

func (s *HTTPServer) getOrder(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.AbortWithStatus(400)
		return
	}

	order, err := s.OrderService.GetOrder(c.Request.Context(), name)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, order)
}
