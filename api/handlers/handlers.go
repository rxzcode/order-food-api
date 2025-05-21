package handlers

import (
	"gorm.io/gorm"
)

type Cache interface {
	AppearsInAtLeastN(code string, n int) bool
}

type InfoOption struct {
	BasePath    string
	CouponCache Cache
}

type Option func(*Handler)

type Handler struct {
	DB   *gorm.DB
	Info InfoOption
}

func WithDB(db *gorm.DB) Option {
	return func(h *Handler) {
		h.DB = db
	}
}

func WithInfo(info InfoOption) Option {
	return func(h *Handler) {
		h.Info = info
	}
}

func NewHandler(opts ...Option) *Handler {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
