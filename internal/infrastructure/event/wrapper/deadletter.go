package wrapper

import (
	"context"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

const deadLetterTopicKey = "DLQ"

// DeadLetterWrapper 死信队列
type DeadLetterWrapper struct {
	b broker.Broker
}

// Wrapper 包装器操作
func (w *DeadLetterWrapper) Wrapper() server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			err := next(ctx, msg)
			if err == nil {
				return nil
			}

			// 如果是死信队列则直接返回不再进入死信队列过程
			if strings.HasSuffix(msg.Topic(), deadLetterTopicKey) {
				return err
			}
			errStatus, ok := status.FromError(err)
			if !ok {
				logger.Errorf("failed to handler topic %v, error: %s; id: %s", msg.Topic(), err.Error(), msg.Header()["Micro-ID"])
				return err
			}
			switch errStatus.Code() {
			case codes.InvalidArgument:
				return nil
			case codes.DataLoss:
				return nil
			case codes.PermissionDenied:
				return nil
			case codes.Unauthenticated:
				return nil
			case codes.Aborted:
				return nil
			case codes.NotFound:
				return nil
			}

			header := make(map[string]string)
			header["error"] = err.Error()
			for k, v := range msg.Header() {
				header[k] = v
			}
			dlMsg := broker.Message{
				Header: header,
				Body:   msg.Body(),
			}
			topic := msg.Topic() + "DLQ"
			if err := w.b.Publish(topic, &dlMsg); err != nil {
				logger.Errorf("failed to publish to %s, error: %s", topic, err.Error())
			}

			// 一律返回nil让broker标记为成功
			return nil
		}
	}
}

func NewDeadLetterWrapper(b broker.Broker) *DeadLetterWrapper {
	return &DeadLetterWrapper{
		b: b,
	}
}
