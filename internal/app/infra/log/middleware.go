package log

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"github.com/typical-go/typical-rest-server/pkg/logruskit"
)

// Middleware log every request
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		start := time.Now()

		// generate request ID if not exist
		reqID := req.Header.Get(echo.HeaderXRequestID)
		if reqID == "" {
			reqID = generateRequestID()
		}
		res.Header().Set(echo.HeaderXRequestID, reqID)

		// put fields in context
		ctx := req.Context()
		logruskit.PutField(&ctx, "req_id", reqID)

		// update request with new context
		req = req.WithContext(ctx)
		c.SetRequest(req)

		// current handler
		if err := next(c); err != nil {
			c.Error(err)
		}

		stop := time.Now()
		if _debug {
			logrus.WithFields(logrus.Fields{
				"exec_time":   stop.Sub(start).String(),
				"req_id":      reqID,
				"resp_status": res.Status,
				"resp_size":   strconv.FormatInt(res.Size, 10),
				"req_size":    req.Header.Get(echo.HeaderContentLength),
			}).Debugf("%s %s", req.Method, req.RequestURI)
		}
		return nil
	}
}

func generateRequestID() string {
	return xid.New().String()
}
