package wrapper

import (
	"context"
	"go-micro.dev/v4/broker"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 死信队列
type DeadLetterWrapper struct {
	b broker.Broker
}

// 包装器操作
func (w *DeadLetterWrapper) Wrapper() server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			err := next(ctx, msg)
			logger.Info("aaaaa")
			if err != nil {
				errStatus, ok := status.FromError(err)
				if !ok {
					logger.Errorf("failed to handler topic %v, error: %s; id: %s", msg.Topic(), err.Error(), msg.Header()["Micro-ID"])
					return err
				}
				switch errStatus.Code() {
				case codes.InvalidArgument:
					return err
				case codes.Aborted:
					return err
				case codes.DataLoss:
					return nil
				}

				logger.Error("failed to consume event: ", err)
				dlMsg := &broker.Message{
					Header: msg.Header(),
					Body:   msg.Body(),
				}
				dlMsg.Header["error"] = err.Error()
				topic := msg.Topic() + "-DLQ"
				if err := w.b.Publish(msg.Topic() + "-DLQ", dlMsg); err != nil {
					logger.Errorf("failed to publish to %s, error: %s", topic, err.Error())
					return nil
				}
			}
			return nil
		}
	}
}

func NewDeadLetterWrapper(b broker.Broker) *DeadLetterWrapper {
	return &DeadLetterWrapper{
		b:  b,
	}
}