// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package singleflight provides a duplicate function call suppression
// mechanism.
// singleflight包提供了重复函数调用抑制机制。
package singleflight // import "golang.org/x/sync/singleflight"

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
)

// errGoexit indicates the runtime.Goexit was called in
// the user given function.
// errGoexit 表示 runtime.Goexit 被用户的函数调用了
var errGoexit = errors.New("runtime.Goexit was called")

// A panicError is an arbitrary value recovered from a panic
// panicError 是从panic中 恢复的任意值
// with the stack trace during the execution of given function.
// 执行给定函数期间的堆栈跟踪
type panicError struct {
	value interface{}
	stack []byte
}

// Error implements error interface.
// Error 实现错误接口
func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.value, p.stack)
}

func newPanicError(v interface{}) error {
	stack := debug.Stack()

	// The first line of the stack trace is of the form "goroutine N [status]:"
	// 堆栈跟踪的第一行的形式为“goroutine N [status]:”
	// but by the time the panic reaches Do the goroutine may no longer exist
	// 但当panic达到 Do 时，goroutine 可能不再存在
	// and its status will have changed. Trim out the misleading line.
	// 并且它的状态将会改变。修剪掉误导性的线条。
	if line := bytes.IndexByte(stack[:], '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{value: v, stack: stack}
}

// call is an in-flight or completed singleflight.Do call
// call 是正在进行的或已完成的 singleflight.Do() 调用
type call struct {
	wg sync.WaitGroup

	// These fields are written once before the WaitGroup is done
	// 这些字段在 WaitGroup 完成之前写入一次
	// and are only read after the WaitGroup is done.
	// 并且仅在 WaitGroup 完成后才读取。
	val interface{}
	err error

	// These fields are read and written with the singleflight
	// 这些字段是用 singleflight mutex  读写的
	// mutex held before the WaitGroup is done, and are read but
	//  在 WaitGroup完成前。
	// not written after the WaitGroup is done.
	// 并且 只读不写，在WaitGroup完成后。
	dups  int
	chans []chan<- Result
}

// Group represents a class of work and forms a namespace in
// Group 代表一个工作类，并在其中形成一个命名空间
// which units of work can be executed with duplicate suppression.
// 哪些工作单元可以通过重复抑制来执行。
type Group struct {
	mu sync.Mutex       // protects m 用来保护m，并发安全
	m  map[string]*call // lazily initialized  延迟初始化
}

// Result holds the results of Do, so they can be passed
// Result保存了Do的结果，因此可以传递
// on a channel.
// 在通道上
type Result struct {
	Val    interface{}
	Err    error
	Shared bool
}

// Do executes and returns the results of the given function,
// Do 执行并返回给定函数的结果
// making sure that only one execution is in-flight for a given key at a time.
// 确保在某一时刻对于给定的键只有一次正在执行
// If a duplicate comes in, the duplicate caller waits for the original
// 如果有重复的调用者进入，则重复的调用者将等待最初者
// to complete and receives the same results.
// 完成并收到相同的结果。
// The return value shared indicates whether v was given to multiple callers.
// 返回值shared表示v是否被给予多个调用者。
func (g *Group) Do(key string, fn func() (interface{}, error)) (v interface{}, err error, shared bool) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait()

		if e, ok := c.err.(*panicError); ok {
			panic(e)
		} else if c.err == errGoexit {
			runtime.Goexit()
		}
		return c.val, c.err, true
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	g.doCall(c, key, fn)
	return c.val, c.err, c.dups > 0
}

// DoChan is like Do but returns a channel that will receive the
// results when they are ready.
// DoChan 与 Do 类似，但返回一个chanel通道 接收准备好后的结果。
//
// The returned channel will not be closed.
// 返回的channel通道不会被关闭。
func (g *Group) DoChan(key string, fn func() (interface{}, error)) <-chan Result {
	ch := make(chan Result, 1)
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		c.dups++
		c.chans = append(c.chans, ch)
		g.mu.Unlock()
		return ch
	}
	c := &call{chans: []chan<- Result{ch}}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	go g.doCall(c, key, fn)

	return ch
}

// doCall handles the single call for a key.
// doCall 处理对key的单个调用。
func (g *Group) doCall(c *call, key string, fn func() (interface{}, error)) {
	normalReturn := false
	recovered := false

	// use double-defer to distinguish panic from runtime.Goexit,
	// 使用双重延迟 来区分panic和runtime.Goexit,
	// more details see https://golang.org/cl/134395
	// 更多详情参见 https://golang.org/cl/134395
	defer func() {
		// the given function invoked runtime.Goexit
		// 调用给定函数runtime.Goexit
		if !normalReturn && !recovered {
			c.err = errGoexit
		}

		g.mu.Lock()
		defer g.mu.Unlock()
		c.wg.Done()
		if g.m[key] == c {
			delete(g.m, key)
		}

		if e, ok := c.err.(*panicError); ok {
			// In order to prevent the waiting channels from being blocked forever,
			// 为了防止等待通道永远被阻塞，
			// needs to ensure that this panic cannot be recovered.
			// 需要确保这种panic恐慌无法恢复。
			if len(c.chans) > 0 {
				go panic(e)
				select {} // Keep this goroutine around so that it will appear in the crash dump.
				// 保留此 goroutine，以便它出现在故障转储中。
			} else {
				panic(e)
			}
		} else if c.err == errGoexit {
			// Already in the process of goexit, no need to call again
			// 已经在goexit过程中，无需再次调用
		} else {
			// Normal return
			// 正常返回
			for _, ch := range c.chans {
				ch <- Result{c.val, c.err, c.dups > 0}
			}
		}
	}()

	func() {
		defer func() {
			if !normalReturn {
				// Ideally, we would wait to take a stack trace until we've determined
				// 理想情况下，我们会等待获取堆栈跟踪，直到我们确定
				// whether this is a panic or a runtime.Goexit.
				// 这是恐慌还是runtime.Goexit。
				//
				// Unfortunately, the only way we can distinguish the two is to see
				// 不幸的是，我们区分两者的唯一方法就是看
				// whether the recover stopped the goroutine from terminating, and by
				// 恢复是否阻止 goroutine 终止，并且通过
				// the time we know that, the part of the stack trace relevant to the
				// 当我们知道时，堆栈跟踪中与
				// panic has been discarded.
				// 恐慌已被丢弃。
				if r := recover(); r != nil {
					c.err = newPanicError(r)
				}
			}
		}()

		c.val, c.err = fn()
		normalReturn = true
	}()

	if !normalReturn {
		recovered = true
	}
}

// Forget tells the singleflight to forget about a key.  Future calls
// Forget 告诉 singleflight 忘记某个键。未来的calls调用
// to Do for this key will call the function rather than waiting for
// 为此键执行的操作将调用该函数而不是等待
// an earlier call to complete.
// 较早的调用完成。
func (g *Group) Forget(key string) {
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()
}
