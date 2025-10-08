package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"path"
	"sync"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"

	"meeting-system/shared/config"
	"meeting-system/shared/logger"
)

// ServiceInfo 保存服务实例信息
type ServiceInfo struct {
	Name       string            `json:"name"`
	InstanceID string            `json:"instance_id"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Protocol   string            `json:"protocol"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Registered time.Time         `json:"registered_at"`
}

type registeredService struct {
	key     string
	leaseID clientv3.LeaseID
	cancel  context.CancelFunc
}

// ServiceRegistry 基于etcd的服务注册中心
type ServiceRegistry struct {
	client   *clientv3.Client
	basePath string
	ttl      int64

	mu         sync.Mutex
	registered map[string]*registeredService
}

// NewServiceRegistry 创建注册中心实例
func NewServiceRegistry(cfg config.EtcdConfig) (*ServiceRegistry, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints not configured")
	}

	dialTimeout := time.Duration(cfg.DialTimeout)
	if dialTimeout <= 0 {
		dialTimeout = 5
	}
	ttl := int64(cfg.TTL)
	if ttl <= 0 {
		ttl = 30
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: dialTimeout * time.Second,
		Username:    cfg.Username,
		Password:    cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect etcd: %w", err)
	}

	basePath := cfg.Namespace
	if basePath == "" {
		basePath = "/services"
	}

	return &ServiceRegistry{
		client:     client,
		basePath:   basePath,
		ttl:        ttl,
		registered: make(map[string]*registeredService),
	}, nil
}

// Close 关闭etcd客户端
func (sr *ServiceRegistry) Close() error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	for key, entry := range sr.registered {
		entry.cancel()
		_, _ = sr.client.Lease.Revoke(context.Background(), entry.leaseID)
		delete(sr.registered, key)
	}

	return sr.client.Close()
}

func (sr *ServiceRegistry) serviceKey(serviceName, instanceID string) string {
	return path.Join(sr.basePath, serviceName, instanceID)
}

// RegisterService 注册服务实例并保持心跳
func (sr *ServiceRegistry) RegisterService(service *ServiceInfo) (string, error) {
	if service == nil {
		return "", fmt.Errorf("service info is nil")
	}
	if service.Name == "" {
		return "", fmt.Errorf("service name is required")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service.InstanceID == "" {
		service.InstanceID = uuid.NewString()
	}

	leaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	leaseResp, err := sr.client.Grant(leaseCtx, sr.ttl)
	cancel()
	if err != nil {
		return "", fmt.Errorf("grant lease failed: %w", err)
	}

	service.Registered = time.Now().UTC()

	payload, err := json.Marshal(service)
	if err != nil {
		return "", fmt.Errorf("marshal service info failed: %w", err)
	}

	key := sr.serviceKey(service.Name, service.InstanceID)
	putCtx, cancelPut := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = sr.client.Put(putCtx, key, string(payload), clientv3.WithLease(leaseResp.ID))
	cancelPut()
	if err != nil {
		return "", fmt.Errorf("register service failed: %w", err)
	}

	keepCtx, keepCancel := context.WithCancel(context.Background())
	ch, err := sr.client.KeepAlive(keepCtx, leaseResp.ID)
	if err != nil {
		keepCancel()
		return "", fmt.Errorf("keepalive failed: %w", err)
	}

	go func(name string, instance string, ka <-chan *clientv3.LeaseKeepAliveResponse) {
		for range ka {
		}
		logger.Warn("Keepalive channel closed", logger.String("service", name), logger.String("instance", instance))
	}(service.Name, service.InstanceID, ch)

	sr.registered[key] = &registeredService{
		key:     key,
		leaseID: leaseResp.ID,
		cancel:  keepCancel,
	}

	logger.Info("Service registered", logger.String("service", service.Name), logger.String("instance", service.InstanceID), logger.String("address", fmt.Sprintf("%s:%d", service.Host, service.Port)))

	return service.InstanceID, nil
}

// DeregisterService 注销服务实例
func (sr *ServiceRegistry) DeregisterService(serviceName, instanceID string) error {
	if serviceName == "" || instanceID == "" {
		return fmt.Errorf("service name and instance id are required")
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	key := sr.serviceKey(serviceName, instanceID)
	if entry, ok := sr.registered[key]; ok {
		entry.cancel()
		_, _ = sr.client.Lease.Revoke(context.Background(), entry.leaseID)
		delete(sr.registered, key)
	}

	_, err := sr.client.Delete(context.Background(), key)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	logger.Info("Service deregistered", logger.String("service", serviceName), logger.String("instance", instanceID))
	return nil
}

// DiscoverServices 发现服务的所有实例
func (sr *ServiceRegistry) DiscoverServices(serviceName string) ([]*ServiceInfo, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	prefix := sr.serviceKey(serviceName, "")
	resp, err := sr.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to query etcd: %w", err)
	}

	instances := make([]*ServiceInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info ServiceInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			logger.Warn("Failed to decode service info", logger.String("service", serviceName), logger.String("key", string(kv.Key)), logger.Err(err))
			continue
		}
		info.InstanceID = path.Base(string(kv.Key))
		instances = append(instances, &info)
	}

	return instances, nil
}

// GetServiceAddress 返回一个可用的服务地址
func (sr *ServiceRegistry) GetServiceAddress(serviceName string) (string, error) {
	instances, err := sr.DiscoverServices(serviceName)
	if err != nil {
		return "", err
	}
	if len(instances) == 0 {
		return "", fmt.Errorf("no instances found for service %s", serviceName)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	instance := instances[r.Intn(len(instances))]
	return fmt.Sprintf("%s:%d", instance.Host, instance.Port), nil
}

// GetService 获取任意服务实例信息
func (sr *ServiceRegistry) GetService(serviceName string) (*ServiceInfo, error) {
	instances, err := sr.DiscoverServices(serviceName)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	return instances[0], nil
}

// GetAllServices 获取所有服务实例
func (sr *ServiceRegistry) GetAllServices() (map[string][]*ServiceInfo, error) {
	resp, err := sr.client.Get(context.Background(), sr.basePath, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to query etcd: %w", err)
	}

	result := make(map[string][]*ServiceInfo)
	for _, kv := range resp.Kvs {
		var info ServiceInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			logger.Warn("Failed to decode service info", logger.String("key", string(kv.Key)), logger.Err(err))
			continue
		}
		info.InstanceID = path.Base(string(kv.Key))

		serviceName := info.Name
		result[serviceName] = append(result[serviceName], &info)
	}

	return result, nil
}

// UpdateServiceMetadata 更新服务的额外信息
func (sr *ServiceRegistry) UpdateServiceMetadata(serviceName, instanceID string, metadata map[string]string) error {
	if serviceName == "" || instanceID == "" {
		return fmt.Errorf("service name and instance id are required")
	}

	key := sr.serviceKey(serviceName, instanceID)
	resp, err := sr.client.Get(context.Background(), key)
	if err != nil {
		return fmt.Errorf("failed to fetch service info: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return fmt.Errorf("service instance not found: %s/%s", serviceName, instanceID)
	}

	var info ServiceInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &info); err != nil {
		return fmt.Errorf("failed to decode service info: %w", err)
	}

	if info.Metadata == nil {
		info.Metadata = make(map[string]string)
	}
	for k, v := range metadata {
		info.Metadata[k] = v
	}

	payload, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	_, err = sr.client.Put(context.Background(), key, string(payload))
	if err != nil {
		return fmt.Errorf("failed to update service metadata: %w", err)
	}
	return nil
}

// ServiceResolver 简单的服务解析器
type ServiceResolver struct {
	registry *ServiceRegistry
}

// NewServiceResolver 创建解析器
func NewServiceResolver(registry *ServiceRegistry) *ServiceResolver {
	return &ServiceResolver{registry: registry}
}

// Resolve 查找服务地址
func (sr *ServiceResolver) Resolve(serviceName string) (string, error) {
	if sr.registry == nil {
		return "", fmt.Errorf("service registry not initialized")
	}
	return sr.registry.GetServiceAddress(serviceName)
}
