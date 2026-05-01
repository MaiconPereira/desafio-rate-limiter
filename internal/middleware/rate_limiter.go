package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/MaiconPereira/desafio-rate-limit/internal/limiter"
)

const blockedMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"

type RateLimiterMiddleware struct {
	limiter *limiter.RateLimiter
}

func NewRateLimiter(limiter *limiter.RateLimiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{limiter: limiter}
}

func (m *RateLimiterMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		subject := subjectFromRequest(r)

		allowed, err := m.limiter.Allow(r.Context(), subject)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(blockedMessage))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func subjectFromRequest(r *http.Request) limiter.Subject {
	token := strings.TrimSpace(r.Header.Get("API_KEY"))
	if token != "" {
		return limiter.Subject{
			Type:  limiter.SubjectToken,
			Value: token,
		}
	}

	return limiter.Subject{
		Type:  limiter.SubjectIP,
		Value: clientIP(r),
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
