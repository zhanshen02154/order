package transaction

import "context"

type TransactionManager interface {
	// 执行事务
	Execute(ctx context.Context, fn func(txCtx context.Context) error) error

	// 执行包含子事务屏障的事务
	ExecuteWithBarrier(ctx context.Context, fn func(txCtx context.Context) error) error
}