package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"meeting-system/shared/logger"
)

type sessionKey struct {
	meetingID string
	modelType string
}

func (k sessionKey) String() string {
	return fmt.Sprintf("%s|%s", k.meetingID, k.modelType)
}

type managedSession struct {
	key sessionKey

	mu       sync.Mutex
	session  *InferenceSession
	lastUsed time.Time
	created  time.Time
}

type SessionManager struct {
	client *EdgeLLMClient

	mu       sync.Mutex
	sessions map[sessionKey]*managedSession

	createGroup      singleflight.Group
	idleTTL          time.Duration
	cleanupInterval  time.Duration
	stopCleanup      chan struct{}
	cleanupStoppedWG sync.WaitGroup
}

func NewSessionManager(client *EdgeLLMClient, idleTTL, cleanupInterval time.Duration) *SessionManager {
	if idleTTL <= 0 {
		idleTTL = 15 * time.Minute
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 1 * time.Minute
	}

	m := &SessionManager{
		client:          client,
		sessions:        make(map[sessionKey]*managedSession),
		idleTTL:         idleTTL,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}

	m.cleanupStoppedWG.Add(1)
	go m.cleanupLoop()

	return m
}

func (m *SessionManager) Close() {
	close(m.stopCleanup)
	m.cleanupStoppedWG.Wait()

	m.mu.Lock()
	sessions := make([]*managedSession, 0, len(m.sessions))
	for _, entry := range m.sessions {
		sessions = append(sessions, entry)
	}
	m.sessions = make(map[sessionKey]*managedSession)
	m.mu.Unlock()

	for _, entry := range sessions {
		entry.mu.Lock()
		session := entry.session
		entry.session = nil
		entry.mu.Unlock()
		if session != nil {
			_ = m.client.Exit(context.Background(), session)
		}
	}
}

func (m *SessionManager) Acquire(ctx context.Context, meetingID string, modelType string) (*InferenceSession, func(), error) {
	if m == nil || m.client == nil {
		return nil, nil, fmt.Errorf("session manager not initialized")
	}
	if modelType == "" {
		return nil, nil, fmt.Errorf("modelType is required")
	}
	if meetingID == "" {
		meetingID = "global"
	}

	key := sessionKey{meetingID: meetingID, modelType: modelType}

	entryAny, err, _ := m.createGroup.Do(key.String(), func() (interface{}, error) {
		return m.getOrCreate(ctx, key)
	})
	if err != nil {
		return nil, nil, err
	}

	entry, ok := entryAny.(*managedSession)
	if !ok || entry == nil {
		return nil, nil, fmt.Errorf("failed to acquire session: invalid entry type")
	}

	entry.mu.Lock()
	if entry.session == nil || entry.session.Conn == nil {
		entry.mu.Unlock()
		return nil, nil, fmt.Errorf("failed to acquire session: session not ready")
	}
	entry.lastUsed = time.Now()

	released := false
	release := func() {
		if released {
			return
		}
		released = true
		entry.lastUsed = time.Now()
		entry.mu.Unlock()
	}

	return entry.session, release, nil
}

func (m *SessionManager) Invalidate(ctx context.Context, meetingID string, modelType string) {
	if m == nil || m.client == nil {
		return
	}
	if meetingID == "" {
		meetingID = "global"
	}
	if modelType == "" {
		return
	}

	key := sessionKey{meetingID: meetingID, modelType: modelType}

	m.mu.Lock()
	entry := m.sessions[key]
	delete(m.sessions, key)
	m.mu.Unlock()

	if entry == nil {
		return
	}

	entry.mu.Lock()
	session := entry.session
	entry.session = nil
	entry.mu.Unlock()

	if session != nil {
		if err := m.client.Exit(ctx, session); err != nil {
			logger.Warn("Failed to exit invalidated session", logger.Err(err),
				logger.String("meeting_id", meetingID),
				logger.String("model", modelType))
		}
	}
}

func (m *SessionManager) getOrCreate(ctx context.Context, key sessionKey) (*managedSession, error) {
	// fast path
	m.mu.Lock()
	entry := m.sessions[key]
	m.mu.Unlock()

	if entry != nil {
		entry.mu.Lock()
		ready := entry.session != nil && entry.session.Conn != nil
		entry.mu.Unlock()
		if ready {
			return entry, nil
		}
	}

	// slow path: create / recreate under lock to avoid duplicate setup for the same key
	m.mu.Lock()
	entry = m.sessions[key]
	if entry == nil {
		entry = &managedSession{
			key:      key,
			lastUsed: time.Now(),
		}
		m.sessions[key] = entry
	}
	m.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()

	// another goroutine may have created it while we were waiting for the entry lock
	if entry.session != nil && entry.session.Conn != nil {
		entry.lastUsed = time.Now()
		return entry, nil
	}

	session, err := m.client.Setup(ctx, key.modelType)
	if err != nil {
		// cleanup entry on setup failure
		m.mu.Lock()
		if current := m.sessions[key]; current == entry {
			delete(m.sessions, key)
		}
		m.mu.Unlock()

		return nil, err
	}

	now := time.Now()
	entry.session = session
	entry.created = now
	entry.lastUsed = now

	logger.Info("AI session setup reused",
		logger.String("meeting_id", key.meetingID),
		logger.String("model", key.modelType),
		logger.String("work_id", session.WorkID),
	)

	return entry, nil
}

func (m *SessionManager) cleanupLoop() {
	defer m.cleanupStoppedWG.Done()

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCleanup:
			return
		case <-ticker.C:
			m.cleanupOnce()
		}
	}
}

func (m *SessionManager) cleanupOnce() {
	if m == nil || m.client == nil {
		return
	}

	now := time.Now()

	m.mu.Lock()
	entries := make([]*managedSession, 0, len(m.sessions))
	for _, entry := range m.sessions {
		entries = append(entries, entry)
	}
	m.mu.Unlock()

	for _, entry := range entries {
		entry.mu.Lock()
		idle := entry.session != nil && entry.session.Conn != nil && now.Sub(entry.lastUsed) > m.idleTTL
		session := entry.session
		key := entry.key
		if idle {
			entry.session = nil
		}
		entry.mu.Unlock()

		if !idle || session == nil {
			continue
		}

		logger.Info("Closing idle AI session",
			logger.String("meeting_id", key.meetingID),
			logger.String("model", key.modelType),
			logger.String("work_id", session.WorkID),
		)

		_ = m.client.Exit(context.Background(), session)

		m.mu.Lock()
		// remove entry only if it wasn't recreated in the meantime
		if current := m.sessions[key]; current == entry {
			delete(m.sessions, key)
		}
		m.mu.Unlock()
	}
}

