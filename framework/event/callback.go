package event

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"time"
)

/*
	TODO use logger from another common package
*/

const (
	defaultCallbackTimeout = 60 * time.Second
	defaultMaxRetries      = 3 // initial attempt + 3 retries
	defaultRetryInterval   = 2 * time.Second
)

// HandleCallback handles the success and failure callback logic using the EventHandler's http.Client
func (eh *EventHandler) HandleCallback(err error, callback CallbackInfo) {
	if err != nil {
		if callback.FailURL != "" {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), eh.callbackConfig.callbackTimeout)
				defer cancel()
				eh.sendCallback(ctx, callback.FailURL)
			}()
		}
	} else {
		if callback.SuccessURL != "" {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), eh.callbackConfig.callbackTimeout)
				defer cancel()
				eh.sendCallback(ctx, callback.SuccessURL)
			}()
		}
	}
}

// sendCallback sends a callback using the EventHandler's http.Client with retry logic
func (eh *EventHandler) sendCallback(ctx context.Context, url string) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= eh.callbackConfig.maxRetries; attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			log.Printf("Failed to create request for callback to: %s, error: %v\n", url, reqErr)
			return
		}

		resp, err = eh.httpClient.Do(req)
		if err != nil {
			log.Printf("Attempt %d: Failed to send callback to: %s, error: %v\n", attempt+1, url, err)
		} else {
			// Read and discard the response body
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				log.Printf("Callback succeeded with status: %s\n", resp.Status)
				return // success, exit the function
			} else if resp.StatusCode >= 500 && resp.StatusCode < 600 {
				// Server error, can retry
				log.Printf("Attempt %d: Server error for callback to %s: %s\n", attempt+1, url, resp.Status)
			} else {
				// Client error or other non-retryable status
				log.Printf("Callback failed with status: %s\n", resp.Status)
				return
			}
		}

		if attempt < eh.callbackConfig.maxRetries {
			// Exponential backoff
			sleepDuration := eh.callbackConfig.retryInterval * (1 << attempt)
			jitter := time.Duration(float64(sleepDuration) * 0.1 * (0.5 - rand.Float64()))
			sleepDuration += jitter

			select {
			case <-ctx.Done():
				log.Printf("Context canceled, aborting retries for callback to: %s\n", url)
				return
			case <-time.After(sleepDuration):
				// Proceed to next attempt
			}
		}
	}

	log.Printf("All retries failed for callback to: %s\n", url)
}
