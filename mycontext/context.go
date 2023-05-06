package mycontext

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var (
	// 检查是否实现了标准库
	_ context.Context = (*myContextImpl)(nil)
	// 检查是否实现了MyContext
	_ MyContext = (*myContextImpl)(nil)
)

// MyContext extend standard library myContextImpl package, make it better
// It should have name, location, environment and logger
type MyContext interface {
	// composite
	context.Context
	// 实现标准库的行为
	WithCancel() (MyContext, func())
	WithDeadline(time.Time) (MyContext, func())
	WithTimeout(time.Duration) (MyContext, func())
	WithValue(key, value interface{}) MyContext

	// Fork return a copied myContextImpl, if there is a new goroutine generated,
	// it will use my name with a sequential number suffix started from 1
	Fork() MyContext
	// At return a copied myContextImpl, specify the current location where it is in,
	// it should chain all locations start from root
	At(location string) MyContext
	ForkAt(location string) MyContext
	// Reborn will use stdcontext.Background() instead of internal myContextImpl,
	// it used for escaping internal myContextImpl's cancel request
	Reborn() MyContext
	// RebornWith will use specified myContextImpl instead of internal myContextImpl,
	// it used for escaping internal myContextImpl's cancel request
	RebornWith(context.Context) MyContext
	// Name returns my logger's name
	Name() string
	// Location returns my logger's location
	Location() string

	// Env return my env
	// WARN: env value and official myContextImpl value are two diffrent things
	Env() Env

	// Logger return my logger
	Logger() Logger
	// shortcut methods of my logger
	Debug(msg string, kvs ...interface{})
	Info(msg string, kvs ...interface{})
	Warn(msg string, kvs ...interface{})
	Error(msg string, kvs ...interface{})
	Panic(msg string, kvs ...interface{})
	Fatal(msg string, kvs ...interface{})
}

type myContextImpl struct {
	stdCtx  context.Context
	tracker *uint64

	env    Env
	logger Logger
}

// Generator 定义了一个Context的生成函数，每次调用都应当返回一个新的Context
type Generator = func() MyContext

// Simple return a basic MyContext, without name, without location,
// and use S() as internal logger
func Simple() MyContext {
	return New(
		context.Background(),
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// Session 返回一个简单的Context，命名部分使用自动生成的uuid。
func Session() MyContext {
	return New(
		context.Background(),
		NewEnv(),
		NewLogger(uuid.New().String(), "", _S, nil, nil),
	)
}

// Named 返回一个简单的Context，Context的名称可以通过参数name实现自定义。
func Named(name string) MyContext {
	return New(
		context.Background(),
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
	)
}

// ToSimple 将给定的标准库Context对象作为Context的内部基础对象。
func ToSimple(stdCtx context.Context) MyContext {
	return New(
		stdCtx,
		NewEnv(),
		NewLogger("", "", _S, nil, nil),
	)
}

// ToSession 将给定的标准库Context对象对位Context的内部基础对象，并自动生成uuid作为name。
func ToSession(stdCtx context.Context) MyContext {
	return New(
		stdCtx,
		NewEnv(),
		NewLogger(uuid.New().String(), "", _S, nil, nil),
	)
}

// ToNamed 使用给定的标准库Context和名称列表构建一个Context。
func ToNamed(stdCtx context.Context, name string) MyContext {
	return New(
		stdCtx,
		NewEnv(),
		NewLogger(name, "", _S, nil, nil),
	)
}

// New use a standard library Context, an Env and a Logger to generate a new myContextImpl.
// It will use the default value if not given.
func New(stdCtx context.Context, env Env, logger Logger) MyContext {
	if stdCtx == nil {
		stdCtx = context.Background()
	}
	if env == nil {
		env = NewEnv()
	}
	if logger == nil {
		logger = NewLogger("", "", nil, nil, nil)
	}

	var tracker uint64
	return &myContextImpl{
		stdCtx:  stdCtx,
		tracker: &tracker,

		env:    env,
		logger: logger,
	}
}

func (myctx *myContextImpl) fork(name, location string) *myContextImpl {
	return &myContextImpl{
		stdCtx:  myctx.stdCtx,
		tracker: myctx.tracker,

		env:    myctx.env.Fork(),
		logger: myctx.logger.Fork(name, location),
	}
}

func (myctx *myContextImpl) Deadline() (deadline time.Time, ok bool) {
	return myctx.stdCtx.Deadline()
}

func (myctx *myContextImpl) Done() <-chan struct{} {
	return myctx.stdCtx.Done()
}

func (myctx *myContextImpl) Err() error {
	return myctx.stdCtx.Err()
}

func (myctx *myContextImpl) Value(key interface{}) interface{} {
	return myctx.stdCtx.Value(key)
}

func (myctx *myContextImpl) WithCancel() (MyContext, func()) {
	var newCtx = myctx.fork("", "")
	var newStdCtx, cancel = context.WithCancel(newCtx.stdCtx)
	newCtx.stdCtx = newStdCtx
	return newCtx, cancel
}

func (myctx *myContextImpl) WithDeadline(d time.Time) (MyContext, func()) {
	var newCtx = myctx.fork("", "")
	var newStdCtx, cancel = context.WithDeadline(newCtx.stdCtx, d)
	newCtx.stdCtx = newStdCtx
	return newCtx, cancel
}

func (myctx *myContextImpl) WithTimeout(timeout time.Duration) (MyContext, func()) {
	var newCtx = myctx.fork("", "")
	var newStdCtx, cancel = context.WithTimeout(newCtx.stdCtx, timeout)
	newCtx.stdCtx = newStdCtx
	return newCtx, cancel
}

func (myctx *myContextImpl) WithValue(key, value interface{}) MyContext {
	var newCtx = myctx.fork("", "")
	newCtx.stdCtx = context.WithValue(newCtx.stdCtx, key, value)
	return newCtx
}

func (myctx *myContextImpl) Fork() MyContext {
	return myctx.ForkAt("")
}

func (myctx *myContextImpl) At(location string) MyContext {
	return myctx.fork("", location)
}

func (myctx *myContextImpl) ForkAt(location string) MyContext {
	var seq = atomic.AddUint64(myctx.tracker, 1)
	var newCtx = myctx.fork(strconv.FormatUint(seq, 10), location)
	var tracker uint64
	newCtx.tracker = &tracker
	return newCtx
}

func (myctx *myContextImpl) Reborn() MyContext {
	return myctx.RebornWith(context.Background())
}

func (myctx *myContextImpl) RebornWith(stdCtx context.Context) MyContext {
	if stdCtx == nil {
		stdCtx = context.Background()
	}
	var newCtx = myctx.fork("", "")
	newCtx.stdCtx = stdCtx
	return newCtx
}

func (myctx *myContextImpl) Name() string {
	return myctx.logger.Name()
}

func (myctx *myContextImpl) Location() string {
	return myctx.logger.Location()
}

func (myctx *myContextImpl) Env() Env {
	return myctx.env
}

func (myctx *myContextImpl) Logger() Logger {
	return myctx.logger
}

func (myctx *myContextImpl) Debug(msg string, kvs ...interface{}) {
	myctx.logger.Debug(msg, kvs...)
}

func (myctx *myContextImpl) Info(msg string, kvs ...interface{}) {
	myctx.logger.Info(msg, kvs...)
}

func (myctx *myContextImpl) Warn(msg string, kvs ...interface{}) {
	myctx.logger.Warn(msg, kvs...)
}

func (myctx *myContextImpl) Error(msg string, kvs ...interface{}) {
	myctx.logger.Error(msg, kvs...)
}

func (myctx *myContextImpl) Panic(msg string, kvs ...interface{}) {
	myctx.logger.Panic(msg, kvs...)
}

func (myctx *myContextImpl) Fatal(msg string, kvs ...interface{}) {
	myctx.logger.Fatal(msg, kvs...)
}
