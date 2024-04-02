package main

import "errors"

type Options struct {
	port *int
}

type Option func(options *Options) error

// DefaultPort 默认端口
var DefaultPort = 6666

func WithPort(port int) Option {
	return func(options *Options) error {
		if port <= 0 {
			return errors.New("port should be positive")
		}
		options.port = &port
		return nil
	}
}
