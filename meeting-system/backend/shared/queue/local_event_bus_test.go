package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalEventBus_EmitAndHandle(t *testing.T) {
	bus := NewLocalEventBus(100, 2)
	bus.Start()
	defer bus.Stop()

	received := make(chan *LocalEvent, 10)
	bus.On("test_event", func(ctx context.Context, event *LocalEvent) error {
		received <- event
		return nil
	})

	// 触发事件
	err := bus.Emit("test_event", map[string]interface{}{
		"message": "hello",
	}, "test_source")
	require.NoError(t, err)

	// 验证事件被处理
	select {
	case event := <-received:
		assert.Equal(t, "test_event", event.Type)
		assert.Equal(t, "hello", event.Payload["message"])
		assert.Equal(t, "test_source", event.Source)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestLocalEventBus_MultipleHandlers(t *testing.T) {
	bus := NewLocalEventBus(100, 2)
	bus.Start()
	defer bus.Stop()

	var count1, count2 int32

	// 注册多个处理器
	bus.On("multi_handler_event", func(ctx context.Context, event *LocalEvent) error {
		atomic.AddInt32(&count1, 1)
		return nil
	})

	bus.On("multi_handler_event", func(ctx context.Context, event *LocalEvent) error {
		atomic.AddInt32(&count2, 1)
		return nil
	})

	// 触发事件
	err := bus.Emit("multi_handler_event", map[string]interface{}{}, "test")
	require.NoError(t, err)

	// 等待处理完成
	time.Sleep(500 * time.Millisecond)

	// 验证两个处理器都被调用
	assert.Equal(t, int32(1), atomic.LoadInt32(&count1))
	assert.Equal(t, int32(2), atomic.LoadInt32(&count2))
}

func TestLocalEventBus_ConcurrentEmit(t *testing.T) {
	bus := NewLocalEventBus(1000, 4)
	bus.Start()
	defer bus.Stop()

	var processedCount int32
	bus.On("concurrent_event", func(ctx context.Context, event *LocalEvent) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	})

	// 并发触发事件
	const numEvents = 100
	var wg sync.WaitGroup
	wg.Add(numEvents)

	for i := 0; i < numEvents; i++ {
		go func(index int) {
			defer wg.Done()
			bus.Emit("concurrent_event", map[string]interface{}{
				"index": index,
			}, "test")
		}(i)
	}

	wg.Wait()
	time.Sleep(1 * time.Second)

	// 验证所有事件都被处理
	assert.Equal(t, int32(numEvents), atomic.LoadInt32(&processedCount))
}

func TestLocalEventBus_EmitSync(t *testing.T) {
	bus := NewLocalEventBus(10, 2)
	bus.Start()
	defer bus.Stop()

	received := make(chan bool, 1)
	bus.On("sync_event", func(ctx context.Context, event *LocalEvent) error {
		received <- true
		return nil
	})

	ctx := context.Background()
	err := bus.EmitSync(ctx, "sync_event", map[string]interface{}{}, "test")
	require.NoError(t, err)

	select {
	case <-received:
		// 成功
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for sync event")
	}
}

func TestLocalEventBus_BufferFull(t *testing.T) {
	// 创建小缓冲区的事件总线
	bus := NewLocalEventBus(2, 1)
	bus.Start()
	defer bus.Stop()

	// 注册一个慢处理器
	bus.On("slow_event", func(ctx context.Context, event *LocalEvent) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// 快速发送多个事件，超过缓冲区大小
	successCount := 0
	for i := 0; i < 10; i++ {
		err := bus.Emit("slow_event", map[string]interface{}{}, "test")
		if err == nil {
			successCount++
		}
	}

	// 应该有一些事件因为缓冲区满而被丢弃
	assert.Less(t, successCount, 10)
}

func TestLocalEventBus_GetStats(t *testing.T) {
	bus := NewLocalEventBus(100, 2)
	bus.Start()
	defer bus.Stop()

	bus.On("stats_event", func(ctx context.Context, event *LocalEvent) error {
		return nil
	})

	// 触发一些事件
	for i := 0; i < 5; i++ {
		bus.Emit("stats_event", map[string]interface{}{}, "test")
	}

	time.Sleep(500 * time.Millisecond)

	// 获取统计信息
	stats := bus.GetStats()
	assert.Equal(t, int64(5), stats["total_events"])
	assert.Equal(t, int64(5), stats["processed_events"])
	assert.Equal(t, int64(0), stats["failed_events"])
}

func TestPriorityLocalEventBus_Priority(t *testing.T) {
	bus := NewPriorityLocalEventBus(100, 1)
	bus.Start()
	defer bus.Stop()

	received := make(chan string, 10)
	bus.On("priority_event", func(ctx context.Context, event *LocalEvent) error {
		received <- event.Payload["priority"].(string)
		return nil
	})

	// 发送不同优先级的事件
	bus.EmitWithPriority(PriorityLow, "priority_event", map[string]interface{}{
		"priority": "low",
	}, "test")

	bus.EmitWithPriority(PriorityCritical, "priority_event", map[string]interface{}{
		"priority": "critical",
	}, "test")

	bus.EmitWithPriority(PriorityHigh, "priority_event", map[string]interface{}{
		"priority": "high",
	}, "test")

	bus.EmitWithPriority(PriorityNormal, "priority_event", map[string]interface{}{
		"priority": "normal",
	}, "test")

	// 验证处理顺序（Critical -> High -> Normal -> Low）
	expectedOrder := []string{"critical", "high", "normal", "low"}
	for i, expected := range expectedOrder {
		select {
		case priority := <-received:
			assert.Equal(t, expected, priority, "Event %d should be %s", i, expected)
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for event %d", i)
		}
	}
}

func TestLocalEventBus_HandlerPanic(t *testing.T) {
	bus := NewLocalEventBus(100, 2)
	bus.Start()
	defer bus.Stop()

	var normalHandlerCalled bool
	
	// 注册一个会panic的处理器
	bus.On("panic_event", func(ctx context.Context, event *LocalEvent) error {
		panic("test panic")
	})

	// 注册一个正常的处理器
	bus.On("panic_event", func(ctx context.Context, event *LocalEvent) error {
		normalHandlerCalled = true
		return nil
	})

	// 触发事件
	err := bus.Emit("panic_event", map[string]interface{}{}, "test")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// 验证正常处理器仍然被调用
	assert.True(t, normalHandlerCalled, "Normal handler should still be called")

	// 验证统计信息记录了失败
	stats := bus.GetStats()
	assert.Greater(t, stats["failed_events"], int64(0))
}

func TestLocalEventBus_ContextTimeout(t *testing.T) {
	bus := NewLocalEventBus(100, 2)
	bus.Start()
	defer bus.Stop()

	timeoutOccurred := make(chan bool, 1)
	
	bus.On("timeout_event", func(ctx context.Context, event *LocalEvent) error {
		select {
		case <-ctx.Done():
			timeoutOccurred <- true
			return ctx.Err()
		case <-time.After(15 * time.Second):
			return nil
		}
	})

	// 触发事件（处理器有10秒超时）
	err := bus.Emit("timeout_event", map[string]interface{}{}, "test")
	require.NoError(t, err)

	// 验证超时发生
	select {
	case <-timeoutOccurred:
		// 超时是预期的
	case <-time.After(12 * time.Second):
		t.Fatal("Handler should have timed out")
	}
}

func BenchmarkLocalEventBus_Emit(b *testing.B) {
	bus := NewLocalEventBus(10000, 4)
	bus.Start()
	defer bus.Stop()

	bus.On("bench_event", func(ctx context.Context, event *LocalEvent) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Emit("bench_event", map[string]interface{}{
			"index": i,
		}, "bench")
	}
}

func BenchmarkLocalEventBus_EmitParallel(b *testing.B) {
	bus := NewLocalEventBus(10000, 8)
	bus.Start()
	defer bus.Stop()

	bus.On("bench_parallel_event", func(ctx context.Context, event *LocalEvent) error {
		return nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			bus.Emit("bench_parallel_event", map[string]interface{}{
				"index": i,
			}, "bench")
			i++
		}
	})
}

