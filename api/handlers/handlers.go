package handlers

import (
	"order-food-api/core/loader"

	"gorm.io/gorm"
)

type InfoOption struct {
	BasePath    string
	CouponCache *loader.Loader
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
