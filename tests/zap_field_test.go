package tests

import (
	"go.uber.org/zap"
	"sync"
	"testing"
)

var (
	fieldPool = sync.Pool{
		New: func() interface{} {
			return make([]zap.Field, 0, 12) // 预分配容量
		},
	}
)

// 使用对象池的方案
func BenchmarkLoggingWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fields := fieldPool.Get().([]zap.Field)
		fields = fields[:0] // 重置长度
		// ... 填充 fields 数据 ...
		fields = append(fields,
			zap.String("service", "order"),
			zap.String("version", "5.0.0"),
		)
		fields = fields[:0] // 再次清空，准备归还
		fieldPool.Put(fields)
	}
}

// 直接创建新切片的方案
func BenchmarkLoggingWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fields := make([]zap.Field, 0, 12) // 在堆栈上直接创建
		fields = append(fields,
			zap.String("service", "order"),
			zap.String("version", "5.0.0"),
		)
	}
}
