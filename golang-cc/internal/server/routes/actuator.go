package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	actuatorrepo "github.com/rafa/golang-cc/internal/repositories/actuator"
	servermw "github.com/rafa/golang-cc/internal/server/middleware"
	actuatorlib "github.com/sinhashubham95/go-actuator"
)

func RegisterActuator(engine *gin.Engine, repository *actuatorrepo.HealthRepository, metrics *servermw.Metrics, authorization gin.HandlerFunc) {
	handler := actuatorlib.GetActuatorHandler(&actuatorlib.Config{
		Endpoints: []int{actuatorlib.Health},
		Health: &actuatorlib.HealthConfig{
			CacheDuration: time.Second,
			Timeout:       3 * time.Second,
			Checkers: []actuatorlib.HealthChecker{{
				Key: "postgresql", Func: repository.Check, IsMandatory: true,
			}},
		},
	})
	group := engine.Group("/actuator", authorization)
	group.GET("/health", gin.WrapH(handler))
	group.GET("/prometheus", gin.WrapH(promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})))
}
