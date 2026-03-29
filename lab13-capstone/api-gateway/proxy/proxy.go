package proxy

import (
	"api-gateway/middleware"
	"io"
	"log"
	"net/http"
	"time"
)

func Handler(targetBase string) http.Handler {

	client := &http.Client{Timeout: 10 * time.Second}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetURL := targetBase + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}
		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, "proxy error", http.StatusBadGateway)
			return
		}

		for key, vals := range r.Header {
			for _, val := range vals {
				outReq.Header.Add(key, val)
			}
		}
		if reqID, ok := r.Context().Value(middleware.RequestIDKey).(string); ok {
			outReq.Header.Set("X-Request-ID", reqID)
		}
		resp, err := client.Do(outReq)
		if err != nil {
			log.Printf("[PROXY ERROR] %s → %s: %v", r.URL.Path, targetURL, err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"upstream service unavailable"}`, http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for key, vals := range resp.Header {
			for _, val := range vals {
				w.Header().Add(key, val)
			}
		}
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)

	})

}
