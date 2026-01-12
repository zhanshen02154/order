package event

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"go-micro.dev/v4"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/metadata"
	"strings"
	"sync"
)

const (
	partitionKey   = "Pkey"
	traceparentKey = "Traceparent"
)

// 事件侦听器
// 异步发送事件
type microListener struct {
	mu             sync.RWMutex
	eventPublisher map[string]micro.Event
	c              client.Client
	successChan    chan *sarama.ProducerMessage
	errorChan      chan *sarama.ProducerError
	wg             sync.WaitGroup
	quitChan       chan struct{}
	// started 用于防止重复 Start
	started bool
	opts    *options
}

// Publish 发布
func (l *microListener) Publish(ctx context.Context, topic string, msg interface{}, key string, opts ...client.PublishOption) error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		return fmt.Errorf("topic: %s event not registerd", topic)
	}
	// 将key放到metadata
	if key != "" {
		if _, ok := metadata.Get(ctx, partitionKey); !ok {
			ctx = metadata.Set(ctx, partitionKey, key)
		}
	}
	err := l.eventPublisher[topic].Publish(ctx, msg, opts...)
	return err
}

// Register 注册
func (l *microListener) Register(topic string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.eventPublisher[topic]; !ok {
		l.eventPublisher[topic] = micro.NewEvent(topic, l.c)
	}
	logger.Info("event ", topic, " was registered")
	return true
}

// UnRegister 取消注册
func (l *microListener) UnRegister(topic string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.eventPublisher == nil {
		return true
	}
	if len(l.eventPublisher) == 0 {
		return true
	}
	if _, ok := l.eventPublisher[topic]; ok {
		delete(l.eventPublisher, topic)
		logger.Info("event: ", topic, " unregistered")
	}
	return true
}

// Close 关闭
func (l *microListener) Close() {
	l.mu.Lock()
	// 如果已经关闭或未初始化，直接返回
	if l.eventPublisher == nil && l.quitChan == nil {
		l.mu.Unlock()
		return
	}
	if l.eventPublisher != nil {
		for k := range l.eventPublisher {
			delete(l.eventPublisher, k)
			logger.Info("event: ", k, " unregistered")
		}
		l.eventPublisher = nil
	}
	if l.quitChan != nil {
		close(l.quitChan)
		l.quitChan = nil
	}
	l.mu.Unlock()

	l.wg.Wait()
}

// Start 启动
func (l *microListener) Start() {
	l.mu.Lock()
	if l.started {
		l.mu.Unlock()
		return
	}
	l.started = true
	l.mu.Unlock()

	l.watchKafkaPipeline()
}

// 监听管道
func (l *microListener) watchKafkaPipeline() {
	l.wg.Add(2)
	go l.handleSuccess()
	go l.handleErrors()

}

// handleSuccess 处理发布成功的逻辑
func (l *microListener) handleSuccess() {
	defer l.wg.Done()
	for {
		select {
		case success, ok := <-l.successChan:
			if !ok {
				return
			}
			l.handleCallback(success, nil)
		case <-l.quitChan:
			logger.Info("Successes handler received stop signal.")
			return
		}
	}
}

// handleErrors 处理发布失败的逻辑
func (l *microListener) handleErrors() {
	defer l.wg.Done()
	for {
		select {
		case errMsg, ok := <-l.errorChan:
			if !ok {
				return
			}
			if errMsg != nil {
				l.handleCallback(errMsg.Msg, errMsg.Err)
			}
		case <-l.quitChan:
			logger.Info("Errors handler received stop signal.")
			return
		}
	}
}

// 处理回调信息
func (l *microListener) handleCallback(sg *sarama.ProducerMessage, err error) {
	if sg == nil || sg.Metadata == nil {
		return
	}
	msg, ok := sg.Metadata.(*broker.Message)
	if !ok || msg == nil {
		return
	}
	if msg.Header == nil {
		return
	}
	if v, ok := msg.Header[traceparentKey]; ok {
		msg.Header[strings.ToLower(traceparentKey)] = v
	}
	ctx := metadata.NewContext(context.Background(), msg.Header)
	ctx = context.WithValue(ctx, partitionContextKey{}, sg.Partition)
	ctx = context.WithValue(ctx, offsetKey{}, sg.Offset)

	fn := func(ctx context.Context, msg *broker.Message, err error) {
		return
	}
	for i := len(l.opts.wrappers); i > 0; i-- {
		fn = l.opts.wrappers[i-1](fn)
	}
	fn(ctx, msg, err)
}

// NewListener 新建侦听器
func NewListener(opts ...Option) Listener {
	listener := microListener{
		mu:             sync.RWMutex{},
		eventPublisher: make(map[string]micro.Event),
		wg:             sync.WaitGroup{},
		quitChan:       make(chan struct{}),
		opts: &options{
			wrappers: make([]PublishCallbackWrapper, 0, 30),
		},
	}
	for _, opt := range opts {
		opt(&listener)
	}
	return &listener
}

// Successes 返回内部 success 通道（用于将 producer.Successes() 转发至此）
func (l *microListener) Successes() chan *sarama.ProducerMessage {
	return l.successChan
}

// Errors 返回内部 error 通道（用于将 producer.Errors() 转发至此）
func (l *microListener) Errors() chan *sarama.ProducerError {
	return l.errorChan
}
