package health

import (
	"context"
	"time"

	"github.com/relay/backend/internal/logger"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

type Check struct {
	Name     string        `json:"name"`
	Status   Status        `json:"status"`
	Duration time.Duration `json:"duration_ms"`
	Message  string        `json:"message,omitempty"`
}

type Response struct {
	Status Status    `json:"status"`
	Checks []Check   `json:"checks"`
	Time   time.Time `json:"time"`
	Uptime int64     `json:"uptime_seconds"`
}

type Checker struct {
	log       *logger.Logger
	startTime time.Time
	checks    []healthCheck
}

type healthCheck func(ctx context.Context) Check

func New(log *logger.Logger) *Checker {
	return &Checker{
		log:       log,
		startTime: time.Now(),
		checks:    make([]healthCheck, 0),
	}
}

func (c *Checker) RegisterCheck(check healthCheck) {
	c.checks = append(c.checks, check)
}

func (c *Checker) Check(ctx context.Context) Response {
	checks := make([]Check, 0, len(c.checks))

	for _, check := range c.checks {
		checks = append(checks, check(ctx))
	}

	overallStatus := StatusHealthy
	for _, check := range checks {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return Response{
		Status: overallStatus,
		Checks: checks,
		Time:   time.Now().UTC(),
		Uptime: int64(time.Since(c.startTime).Seconds()),
	}
}

func (c *Checker) LivenessCheck() bool {
	return true
}

func (c *Checker) ReadinessCheck() bool {
	response := c.Check(context.Background())
	return response.Status == StatusHealthy
}
