package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// 测试配置
type StressTestConfig struct {
	UserServiceURL    string
	MeetingServiceURL string
	ConcurrentUsers   []int // 并发用户数级别
	TestDuration      time.Duration
	RequestTimeout    time.Duration
}

// 性能指标
type PerformanceMetrics struct {
	TotalRequests     int64
	SuccessRequests   int64
	FailedRequests    int64
	TotalResponseTime int64
	MinResponseTime   int64
	MaxResponseTime   int64
	StartTime         time.Time
	EndTime           time.Time
}

// 测试结果
type TestResult struct {
	TestName        string
	ConcurrentUsers int
	Duration        time.Duration
	Metrics         PerformanceMetrics
	ThroughputRPS   float64
	AvgResponseTime float64
	ErrorRate       float64
	MemoryUsage     runtime.MemStats
}

// 用户注册请求 - 匹配服务端UserCreateRequest结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
}

// 用户登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
}

// 会议创建请求
type CreateMeetingRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Type        int       `json:"type"`
	Password    string    `json:"password"`
}

// 压力测试器
type StressTester struct {
	config  StressTestConfig
	client  *http.Client
	results []TestResult
	mu      sync.Mutex
}

// 创建新的压力测试器
func NewStressTester(config StressTestConfig) *StressTester {
	return &StressTester{
		config: config,
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
		results: make([]TestResult, 0),
	}
}

// 发送HTTP请求
func (st *StressTester) sendRequest(method, url string, body interface{}, token string) (*http.Response, time.Duration, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	start := time.Now()
	resp, err := st.client.Do(req)
	duration := time.Since(start)

	return resp, duration, err
}

// 用户注册测试
func (st *StressTester) testUserRegistration(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
	defer wg.Done()

	req := RegisterRequest{
		Username: fmt.Sprintf("testuser%d", userID),
		Email:    fmt.Sprintf("testuser%d@example.com", userID),
		Password: "password123",
		Nickname: fmt.Sprintf("Test User %d", userID),
		Phone:    "", // 可选字段
	}

	resp, duration, err := st.sendRequest("POST", st.config.UserServiceURL+"/api/v1/auth/register", req, "")

	atomic.AddInt64(&metrics.TotalRequests, 1)
	atomic.AddInt64(&metrics.TotalResponseTime, duration.Nanoseconds())

	if err != nil || resp.StatusCode != http.StatusOK {
		atomic.AddInt64(&metrics.FailedRequests, 1)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	atomic.AddInt64(&metrics.SuccessRequests, 1)
	resp.Body.Close()

	// 更新最小/最大响应时间
	for {
		min := atomic.LoadInt64(&metrics.MinResponseTime)
		if min == 0 || duration.Nanoseconds() < min {
			if atomic.CompareAndSwapInt64(&metrics.MinResponseTime, min, duration.Nanoseconds()) {
				break
			}
		} else {
			break
		}
	}

	for {
		max := atomic.LoadInt64(&metrics.MaxResponseTime)
		if duration.Nanoseconds() > max {
			if atomic.CompareAndSwapInt64(&metrics.MaxResponseTime, max, duration.Nanoseconds()) {
				break
			}
		} else {
			break
		}
	}
}

// 用户登录测试
func (st *StressTester) testUserLogin(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) (*string, error) {
	defer wg.Done()

	req := LoginRequest{
		Username: fmt.Sprintf("testuser%d", userID),
		Password: "password123",
	}

	resp, duration, err := st.sendRequest("POST", st.config.UserServiceURL+"/api/v1/auth/login", req, "")

	atomic.AddInt64(&metrics.TotalRequests, 1)
	atomic.AddInt64(&metrics.TotalResponseTime, duration.Nanoseconds())

	if err != nil || resp.StatusCode != http.StatusOK {
		atomic.AddInt64(&metrics.FailedRequests, 1)
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	}

	atomic.AddInt64(&metrics.SuccessRequests, 1)

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	return &loginResp.Token, nil
}

// 会议创建测试
func (st *StressTester) testMeetingCreation(userID int, token string, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
	defer wg.Done()

	req := CreateMeetingRequest{
		Title:       fmt.Sprintf("Test Meeting %d", userID),
		Description: fmt.Sprintf("This is a test meeting created by user %d", userID),
		StartTime:   time.Now().Add(time.Hour),
		EndTime:     time.Now().Add(2 * time.Hour),
		Type:        1,
		Password:    "meeting123",
	}

	resp, duration, err := st.sendRequest("POST", st.config.MeetingServiceURL+"/api/v1/meetings", req, token)

	atomic.AddInt64(&metrics.TotalRequests, 1)
	atomic.AddInt64(&metrics.TotalResponseTime, duration.Nanoseconds())

	if err != nil || resp.StatusCode != http.StatusCreated {
		atomic.AddInt64(&metrics.FailedRequests, 1)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	atomic.AddInt64(&metrics.SuccessRequests, 1)
	resp.Body.Close()
}

// 运行并发测试
func (st *StressTester) runConcurrentTest(testName string, concurrentUsers int, testFunc func(int, *PerformanceMetrics, *sync.WaitGroup)) TestResult {
	fmt.Printf("🚀 开始 %s 测试 (并发用户: %d)\n", testName, concurrentUsers)

	metrics := PerformanceMetrics{
		StartTime: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	// 启动并发测试
	for i := 0; i < concurrentUsers; i++ {
		go testFunc(i, &metrics, &wg)
	}

	wg.Wait()
	metrics.EndTime = time.Now()

	// 收集内存使用情况
	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	// 计算性能指标
	duration := metrics.EndTime.Sub(metrics.StartTime)
	throughput := float64(metrics.SuccessRequests) / duration.Seconds()
	avgResponseTime := float64(metrics.TotalResponseTime) / float64(metrics.TotalRequests) / 1e6 // 转换为毫秒
	errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100

	result := TestResult{
		TestName:        testName,
		ConcurrentUsers: concurrentUsers,
		Duration:        duration,
		Metrics:         metrics,
		ThroughputRPS:   throughput,
		AvgResponseTime: avgResponseTime,
		ErrorRate:       errorRate,
		MemoryUsage:     memStats,
	}

	st.mu.Lock()
	st.results = append(st.results, result)
	st.mu.Unlock()

	fmt.Printf("✅ %s 测试完成\n", testName)
	fmt.Printf("   - 吞吐量: %.2f RPS\n", throughput)
	fmt.Printf("   - 平均响应时间: %.2f ms\n", avgResponseTime)
	fmt.Printf("   - 错误率: %.2f%%\n", errorRate)
	fmt.Printf("   - 内存使用: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Println()

	return result
}

// 运行所有压力测试
func (st *StressTester) RunAllTests() {
	fmt.Println("🎯 开始会议系统压力测试")
	fmt.Println("========================================")

	for _, concurrentUsers := range st.config.ConcurrentUsers {
		fmt.Printf("📊 测试负载级别: %d 并发用户\n", concurrentUsers)
		fmt.Println("----------------------------------------")

		// 1. 用户注册压力测试
		st.runConcurrentTest("用户注册", concurrentUsers, st.testUserRegistration)

		// 等待一段时间让系统恢复
		time.Sleep(2 * time.Second)

		// 2. 用户登录压力测试
		st.runConcurrentTest("用户登录", concurrentUsers, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
			st.testUserLogin(userID, metrics, wg)
		})

		time.Sleep(2 * time.Second)

		// 3. 会议创建压力测试（需要先登录获取token）
		// 这里简化处理，使用固定token进行测试
		st.runConcurrentTest("会议创建", concurrentUsers, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
			// 在实际测试中，这里应该先登录获取真实token
			st.testMeetingCreation(userID, "test-token", metrics, wg)
		})

		time.Sleep(5 * time.Second) // 更长的恢复时间
	}
}

// 生成测试报告
func (st *StressTester) GenerateReport() {
	fmt.Println("📋 压力测试报告")
	fmt.Println("========================================")

	for _, result := range st.results {
		fmt.Printf("🔍 %s (并发用户: %d)\n", result.TestName, result.ConcurrentUsers)
		fmt.Printf("   - 测试时长: %v\n", result.Duration)
		fmt.Printf("   - 总请求数: %d\n", result.Metrics.TotalRequests)
		fmt.Printf("   - 成功请求: %d\n", result.Metrics.SuccessRequests)
		fmt.Printf("   - 失败请求: %d\n", result.Metrics.FailedRequests)
		fmt.Printf("   - 吞吐量: %.2f RPS\n", result.ThroughputRPS)
		fmt.Printf("   - 平均响应时间: %.2f ms\n", result.AvgResponseTime)
		fmt.Printf("   - 最小响应时间: %.2f ms\n", float64(result.Metrics.MinResponseTime)/1e6)
		fmt.Printf("   - 最大响应时间: %.2f ms\n", float64(result.Metrics.MaxResponseTime)/1e6)
		fmt.Printf("   - 错误率: %.2f%%\n", result.ErrorRate)
		fmt.Printf("   - 内存使用: %.2f MB\n", float64(result.MemoryUsage.Alloc)/1024/1024)
		fmt.Println()
	}

	// 性能评级
	st.generatePerformanceRating()
}

// 生成性能评级
func (st *StressTester) generatePerformanceRating() {
	fmt.Println("🏆 性能评级")
	fmt.Println("----------------------------------------")

	totalTests := len(st.results)
	excellentTests := 0
	goodTests := 0
	averageTests := 0
	poorTests := 0

	for _, result := range st.results {
		if result.ErrorRate < 1.0 && result.AvgResponseTime < 100 && result.ThroughputRPS > 100 {
			excellentTests++
		} else if result.ErrorRate < 5.0 && result.AvgResponseTime < 500 && result.ThroughputRPS > 50 {
			goodTests++
		} else if result.ErrorRate < 10.0 && result.AvgResponseTime < 1000 && result.ThroughputRPS > 20 {
			averageTests++
		} else {
			poorTests++
		}
	}

	fmt.Printf("🌟 优秀 (错误率<1%%, 响应时间<100ms, 吞吐量>100RPS): %d/%d\n", excellentTests, totalTests)
	fmt.Printf("👍 良好 (错误率<5%%, 响应时间<500ms, 吞吐量>50RPS): %d/%d\n", goodTests, totalTests)
	fmt.Printf("📊 一般 (错误率<10%%, 响应时间<1000ms, 吞吐量>20RPS): %d/%d\n", averageTests, totalTests)
	fmt.Printf("⚠️  较差: %d/%d\n", poorTests, totalTests)

	overallRating := "较差"
	if excellentTests > totalTests/2 {
		overallRating = "优秀"
	} else if goodTests > totalTests/3 {
		overallRating = "良好"
	} else if averageTests > totalTests/4 {
		overallRating = "一般"
	}

	fmt.Printf("\n🎯 总体评级: %s\n", overallRating)
}

// 运行完整的压力测试套件
func (st *StressTester) RunCompleteStressTest() {
	fmt.Println("🎯 开始完整压力测试套件")
	fmt.Println("========================================")

	// 1. 基准性能测试
	st.runBaselineTests()

	// 2. 渐进式负载测试
	st.runProgressiveLoadTests()

	// 3. 峰值负载测试
	st.runPeakLoadTests()

	// 4. 稳定性测试
	st.runStabilityTests()

	// 5. 混合场景测试
	st.runMixedScenarioTests()
}

// 基准性能测试
func (st *StressTester) runBaselineTests() {
	fmt.Println("📊 1. 基准性能测试")
	fmt.Println("----------------------------------------")

	// 单用户性能测试
	st.runConcurrentTest("基准-用户注册", 1, st.testUserRegistration)
	time.Sleep(1 * time.Second)
	st.runConcurrentTest("基准-用户登录", 1, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
		st.testUserLogin(userID, metrics, wg)
	})
	time.Sleep(1 * time.Second)
}

// 渐进式负载测试
func (st *StressTester) runProgressiveLoadTests() {
	fmt.Println("📈 2. 渐进式负载测试")
	fmt.Println("----------------------------------------")

	for _, concurrentUsers := range st.config.ConcurrentUsers {
		fmt.Printf("🔥 负载级别: %d 并发用户\n", concurrentUsers)

		// 用户注册测试
		st.runConcurrentTest(fmt.Sprintf("渐进-用户注册-%d", concurrentUsers), concurrentUsers, st.testUserRegistration)
		time.Sleep(2 * time.Second)

		// 用户登录测试
		st.runConcurrentTest(fmt.Sprintf("渐进-用户登录-%d", concurrentUsers), concurrentUsers, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
			st.testUserLogin(userID, metrics, wg)
		})
		time.Sleep(2 * time.Second)

		// 会议创建测试
		st.runConcurrentTest(fmt.Sprintf("渐进-会议创建-%d", concurrentUsers), concurrentUsers, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
			st.testMeetingCreation(userID, "test-token", metrics, wg)
		})
		time.Sleep(3 * time.Second)
	}
}

// 峰值负载测试
func (st *StressTester) runPeakLoadTests() {
	fmt.Println("🚀 3. 峰值负载测试")
	fmt.Println("----------------------------------------")

	peakUsers := []int{1000, 1500, 2000}
	for _, users := range peakUsers {
		fmt.Printf("⚡ 峰值测试: %d 并发用户\n", users)

		// 用户登录峰值测试
		st.runConcurrentTest(fmt.Sprintf("峰值-用户登录-%d", users), users, func(userID int, metrics *PerformanceMetrics, wg *sync.WaitGroup) {
			st.testUserLogin(userID%1000, metrics, wg) // 重复使用已注册用户
		})
		time.Sleep(5 * time.Second)
	}
}

// 稳定性测试
func (st *StressTester) runStabilityTests() {
	fmt.Println("⏱️ 4. 稳定性测试")
	fmt.Println("----------------------------------------")

	// 长时间稳定性测试
	st.runLongRunningTest("稳定性-持续负载", 100, 60*time.Second)
}

// 混合场景测试
func (st *StressTester) runMixedScenarioTests() {
	fmt.Println("🎭 5. 混合场景测试")
	fmt.Println("----------------------------------------")

	// 模拟真实用户行为
	st.runMixedScenarioTest("混合场景-真实用户行为", 200)
}

// 长时间运行测试
func (st *StressTester) runLongRunningTest(testName string, concurrentUsers int, duration time.Duration) {
	fmt.Printf("🕐 开始 %s 测试 (并发用户: %d, 持续时间: %v)\n", testName, concurrentUsers, duration)

	metrics := PerformanceMetrics{
		StartTime: time.Now(),
	}

	var wg sync.WaitGroup
	stopChan := make(chan bool)

	// 启动并发测试
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					// 模拟用户行为：登录 -> 创建会议 -> 查询会议
					st.performUserWorkflow(userID, &metrics)
					time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond) // 随机间隔
				}
			}
		}(i)
	}

	// 运行指定时间
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	metrics.EndTime = time.Now()

	// 计算性能指标
	testDuration := metrics.EndTime.Sub(metrics.StartTime)
	throughput := float64(metrics.SuccessRequests) / testDuration.Seconds()
	avgResponseTime := float64(metrics.TotalResponseTime) / float64(metrics.TotalRequests) / 1e6
	errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100

	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	result := TestResult{
		TestName:        testName,
		ConcurrentUsers: concurrentUsers,
		Duration:        testDuration,
		Metrics:         metrics,
		ThroughputRPS:   throughput,
		AvgResponseTime: avgResponseTime,
		ErrorRate:       errorRate,
		MemoryUsage:     memStats,
	}

	st.mu.Lock()
	st.results = append(st.results, result)
	st.mu.Unlock()

	fmt.Printf("✅ %s 测试完成\n", testName)
	fmt.Printf("   - 总请求数: %d\n", metrics.TotalRequests)
	fmt.Printf("   - 吞吐量: %.2f RPS\n", throughput)
	fmt.Printf("   - 平均响应时间: %.2f ms\n", avgResponseTime)
	fmt.Printf("   - 错误率: %.2f%%\n", errorRate)
	fmt.Printf("   - 内存使用: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Println()
}

// 混合场景测试
func (st *StressTester) runMixedScenarioTest(testName string, concurrentUsers int) {
	fmt.Printf("🎭 开始 %s 测试 (并发用户: %d)\n", testName, concurrentUsers)

	metrics := PerformanceMetrics{
		StartTime: time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	// 启动混合场景测试
	for i := 0; i < concurrentUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			var localWg sync.WaitGroup

			// 随机选择用户行为
			scenario := rand.Intn(4)
			switch scenario {
			case 0:
				// 新用户注册
				localWg.Add(1)
				st.testUserRegistration(userID+10000, &metrics, &localWg)
			case 1:
				// 用户登录
				localWg.Add(1)
				st.testUserLogin(userID%1000, &metrics, &localWg)
			case 2:
				// 创建会议
				localWg.Add(1)
				st.testMeetingCreation(userID, "test-token", &metrics, &localWg)
			case 3:
				// 完整用户工作流
				st.performUserWorkflow(userID, &metrics)
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()

	// 计算性能指标
	testDuration := metrics.EndTime.Sub(metrics.StartTime)
	throughput := float64(metrics.SuccessRequests) / testDuration.Seconds()
	avgResponseTime := float64(metrics.TotalResponseTime) / float64(metrics.TotalRequests) / 1e6
	errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100

	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	result := TestResult{
		TestName:        testName,
		ConcurrentUsers: concurrentUsers,
		Duration:        testDuration,
		Metrics:         metrics,
		ThroughputRPS:   throughput,
		AvgResponseTime: avgResponseTime,
		ErrorRate:       errorRate,
		MemoryUsage:     memStats,
	}

	st.mu.Lock()
	st.results = append(st.results, result)
	st.mu.Unlock()

	fmt.Printf("✅ %s 测试完成\n", testName)
	fmt.Printf("   - 吞吐量: %.2f RPS\n", throughput)
	fmt.Printf("   - 平均响应时间: %.2f ms\n", avgResponseTime)
	fmt.Printf("   - 错误率: %.2f%%\n", errorRate)
	fmt.Printf("   - 内存使用: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
	fmt.Println()
}

// 执行用户工作流
func (st *StressTester) performUserWorkflow(userID int, metrics *PerformanceMetrics) {
	var wg sync.WaitGroup

	// 1. 用户登录
	wg.Add(1)
	token, err := st.testUserLogin(userID%1000, metrics, &wg)
	if err != nil || token == nil {
		return
	}

	// 2. 创建会议
	wg.Add(1)
	st.testMeetingCreation(userID, *token, metrics, &wg)

	// 3. 查询会议列表
	st.testMeetingList(*token, metrics)
}

// 会议列表查询测试
func (st *StressTester) testMeetingList(token string, metrics *PerformanceMetrics) {
	resp, duration, err := st.sendRequest("GET", st.config.MeetingServiceURL+"/api/v1/meetings", nil, token)

	atomic.AddInt64(&metrics.TotalRequests, 1)
	atomic.AddInt64(&metrics.TotalResponseTime, duration.Nanoseconds())

	if err != nil || resp.StatusCode != http.StatusOK {
		atomic.AddInt64(&metrics.FailedRequests, 1)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}

	atomic.AddInt64(&metrics.SuccessRequests, 1)
	resp.Body.Close()
}

// 检查服务可用性
func (st *StressTester) checkServiceAvailability() bool {
	userServiceOK := false
	meetingServiceOK := false

	// 检查用户服务
	fmt.Println("  检查用户服务...")
	resp, _, err := st.sendRequest("GET", st.config.UserServiceURL+"/health", nil, "")
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("  ❌ 用户服务不可用: %v\n", err)
		if resp != nil {
			resp.Body.Close()
		}
	} else {
		fmt.Println("  ✅ 用户服务可用")
		userServiceOK = true
		resp.Body.Close()
	}

	// 检查会议服务
	fmt.Println("  检查会议服务...")
	resp, _, err = st.sendRequest("GET", st.config.MeetingServiceURL+"/health", nil, "")
	if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("  ⚠️ 会议服务不可用: %v (将跳过会议相关测试)\n", err)
		if resp != nil {
			resp.Body.Close()
		}
	} else {
		fmt.Println("  ✅ 会议服务可用")
		meetingServiceOK = true
		resp.Body.Close()
	}

	// 至少需要用户服务可用
	_ = meetingServiceOK // 暂时不使用，但保留用于将来扩展
	return userServiceOK
}

// 生成详细报告
func (st *StressTester) generateDetailedReport() {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("stress-test-detailed-report-%s.md", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("❌ 无法创建报告文件: %v\n", err)
		return
	}
	defer file.Close()

	// 写入详细报告
	fmt.Fprintf(file, "# 会议系统完整压力测试详细报告\n\n")
	fmt.Fprintf(file, "**测试时间**: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**测试环境**: %s %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(file, "**Go版本**: %s\n\n", runtime.Version())

	fmt.Fprintf(file, "## 📊 测试概览\n\n")
	fmt.Fprintf(file, "| 测试名称 | 并发用户 | 总请求数 | 成功请求 | 失败请求 | 吞吐量(RPS) | 平均响应时间(ms) | 错误率(%%) |\n")
	fmt.Fprintf(file, "|----------|----------|----------|----------|----------|-------------|------------------|----------|\n")

	for _, result := range st.results {
		fmt.Fprintf(file, "| %s | %d | %d | %d | %d | %.2f | %.2f | %.2f |\n",
			result.TestName,
			result.ConcurrentUsers,
			result.Metrics.TotalRequests,
			result.Metrics.SuccessRequests,
			result.Metrics.FailedRequests,
			result.ThroughputRPS,
			result.AvgResponseTime,
			result.ErrorRate,
		)
	}

	fmt.Fprintf(file, "\n## 🎯 性能分析\n\n")
	fmt.Fprintf(file, "### 最佳性能指标\n")

	// 找出最佳性能指标
	var bestThroughput, bestResponseTime TestResult
	bestThroughput.ThroughputRPS = 0
	bestResponseTime.AvgResponseTime = float64(^uint(0) >> 1) // 最大值

	for _, result := range st.results {
		if result.ThroughputRPS > bestThroughput.ThroughputRPS {
			bestThroughput = result
		}
		if result.AvgResponseTime < bestResponseTime.AvgResponseTime && result.AvgResponseTime > 0 {
			bestResponseTime = result
		}
	}

	fmt.Fprintf(file, "- **最高吞吐量**: %.2f RPS (%s)\n", bestThroughput.ThroughputRPS, bestThroughput.TestName)
	fmt.Fprintf(file, "- **最低响应时间**: %.2f ms (%s)\n", bestResponseTime.AvgResponseTime, bestResponseTime.TestName)

	fmt.Fprintf(file, "\n---\n*报告生成时间: %s*\n", time.Now().Format("2006-01-02 15:04:05"))

	fmt.Printf("📋 详细报告已生成: %s\n", filename)
}

func main() {
	fmt.Println("🔥 会议系统完整压力测试")
	fmt.Println("========================================")
	fmt.Printf("开始时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	config := StressTestConfig{
		UserServiceURL:    "http://localhost:8081",
		MeetingServiceURL: "http://localhost:8082",
		ConcurrentUsers:   []int{10, 50, 100, 200, 500},
		TestDuration:      30 * time.Second,
		RequestTimeout:    10 * time.Second,
	}

	tester := NewStressTester(config)

	// 检查服务可用性
	fmt.Println("🔍 检查服务可用性...")
	if !tester.checkServiceAvailability() {
		fmt.Println("❌ 服务不可用，请先启动服务")
		return
	}
	fmt.Println("✅ 服务可用，开始压力测试")
	fmt.Println()

	// 运行完整压力测试
	tester.RunCompleteStressTest()

	// 生成详细报告
	tester.GenerateReport()
	tester.generateDetailedReport()

	fmt.Printf("结束时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
