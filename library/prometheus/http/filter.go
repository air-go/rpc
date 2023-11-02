package http

import "github.com/gin-gonic/gin"

// Filter if hit filter return false, don't incr metrics's statistics
type Filter func(*gin.Context) bool
