package handlers

import (
	"gorm.io/gorm"
)

type Option func(*Handler)

type Handler struct {
	DB *gorm.DB
	// logger *Logger
}

func WithDB(db *gorm.DB) Option {
	return func(h *Handler) {
		h.DB = db
	}
}

// func WithLogger(logger *Logger) Option {
// 	return func(h *Handler) {
// 		h.logger = logger
// 	}
// }

func NewHandler(opts ...Option) *Handler {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
