package main

import (
	"runtime"
	"sync"
)

var (
	renderFuncs chan func()
	purge       chan bool
	initOnce    sync.Once
)

func init() {
	renderFuncs = make(chan func(), 1000)
	purge = make(chan bool)
}

func Queue(f func()) {
	renderFuncs <- f
}

func PurgeQueue() {
	purge <- true
	<-purge
}

func InitQueue() {
	initOnce.Do(func() {
		go func() {
			runtime.LockOSThread()
			for {
				select {
				case f := <-renderFuncs:
					f()
				case <-purge:
					for {
						select {
						case f := <-renderFuncs:
							f()
						default:
							goto purged
						}
					}
				purged:
					purge <- true
				}
			}
		}()
	})
}
