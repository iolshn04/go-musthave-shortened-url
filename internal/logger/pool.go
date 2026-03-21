package logger

import "github.com/iolshn04/go-musthave-shortened-url/internal/pool"

var writerPool = pool.New(func() *responseWriter {
	return &responseWriter{}
})
