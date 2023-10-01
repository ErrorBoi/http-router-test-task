package main

import "time"

type proxyInitParams struct {
	readTimeout         time.Duration
	writeTimeout        time.Duration
	maxIdleConnDuration time.Duration
}
