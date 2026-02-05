package ws

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/websocket/v2"
)

type Manager struct {
	mu              sync.RWMutex
	rooms           map[string]*room
	cacheSize       int
	roomTTL         time.Duration
	cleanupInterval time.Duration
	done            chan struct{}
	currentOnline   int64
	joinTotal       int64
	leaveTotal      int64
	broadcastTotal  int64
	broadcastErrors int64
	broadcastFanout int64
	statsMu         sync.Mutex
	latencyBins     []int64
}

type Config struct {
	CacheSize       int
	RoomTTL         time.Duration
	CleanupInterval time.Duration
}

type room struct {
	mu           sync.Mutex
	clients      map[*Client]struct{}
	cache        []cachedMessage
	lastActivity time.Time
}

type cachedMessage struct {
	payload []byte
	at      time.Time
}

type Client struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

type Snapshot struct {
	CurrentOnline      int64            `json:"currentOnline"`
	Rooms              int              `json:"rooms"`
	JoinTotal          int64            `json:"joinTotal"`
	LeaveTotal         int64            `json:"leaveTotal"`
	BroadcastTotal     int64            `json:"broadcastTotal"`
	BroadcastErrors    int64            `json:"broadcastErrors"`
	BroadcastFanout    int64            `json:"broadcastFanout"`
	BroadcastP95MS     float64          `json:"broadcastP95Ms"`
	ByRoomType         map[string]int64 `json:"byRoomType"`
	AvgRecipients      float64          `json:"avgRecipients"`
	BroadcastErrorRate float64          `json:"broadcastErrorRate"`
}

func NewManager(cfg Config) *Manager {
	cacheSize := cfg.CacheSize
	if cacheSize <= 0 {
		cacheSize = 3
	}
	roomTTL := cfg.RoomTTL
	if roomTTL <= 0 {
		roomTTL = 30 * time.Second
	}
	cleanupInterval := cfg.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Second
	}
	manager := &Manager{
		rooms:           make(map[string]*room),
		cacheSize:       cacheSize,
		roomTTL:         roomTTL,
		cleanupInterval: cleanupInterval,
		done:            make(chan struct{}),
		latencyBins:     make([]int64, len(latencyBoundariesMS)+1),
	}
	go manager.cleanupLoop()
	return manager
}

func (m *Manager) Join(roomKey string, conn *websocket.Conn) (*Client, [][]byte) {
	if roomKey == "" || conn == nil {
		return nil, nil
	}

	m.mu.Lock()
	rm := m.rooms[roomKey]
	if rm == nil {
		rm = &room{
			clients:      make(map[*Client]struct{}),
			cache:        []cachedMessage{},
			lastActivity: time.Now(),
		}
		m.rooms[roomKey] = rm
	}
	cl := &Client{conn: conn}
	rm.mu.Lock()
	rm.clients[cl] = struct{}{}
	rm.lastActivity = time.Now()
	cached := make([][]byte, len(rm.cache))
	for i, msg := range rm.cache {
		cached[i] = append([]byte(nil), msg.payload...)
	}
	rm.mu.Unlock()
	m.currentOnline++
	m.joinTotal++
	m.mu.Unlock()

	return cl, cached
}

func (m *Manager) Leave(roomKey string, cl *Client) {
	if roomKey == "" || cl == nil {
		return
	}
	m.mu.RLock()
	rm := m.rooms[roomKey]
	m.mu.RUnlock()
	if rm == nil {
		return
	}
	rm.mu.Lock()
	_, existed := rm.clients[cl]
	delete(rm.clients, cl)
	rm.lastActivity = time.Now()
	rm.mu.Unlock()
	if existed {
		m.mu.Lock()
		if m.currentOnline > 0 {
			m.currentOnline--
		}
		m.leaveTotal++
		m.mu.Unlock()
	}
}

func (m *Manager) Broadcast(roomKey string, payload []byte) {
	if roomKey == "" || len(payload) == 0 {
		return
	}
	m.mu.RLock()
	rm := m.rooms[roomKey]
	m.mu.RUnlock()
	if rm == nil {
		return
	}

	rm.mu.Lock()
	rm.cache = append(rm.cache, cachedMessage{payload: append([]byte(nil), payload...), at: time.Now()})
	if len(rm.cache) > m.cacheSize {
		rm.cache = rm.cache[len(rm.cache)-m.cacheSize:]
	}
	rm.lastActivity = time.Now()
	clients := make([]*Client, 0, len(rm.clients))
	for cl := range rm.clients {
		clients = append(clients, cl)
	}
	rm.mu.Unlock()

	start := time.Now()
	var writeErrs int64
	for _, cl := range clients {
		if err := cl.Write(payload); err != nil {
			writeErrs++
			m.Leave(roomKey, cl)
		}
	}
	atomic.AddInt64(&m.broadcastTotal, 1)
	atomic.AddInt64(&m.broadcastFanout, int64(len(clients)))
	if writeErrs > 0 {
		atomic.AddInt64(&m.broadcastErrors, writeErrs)
	}
	m.recordBroadcastLatency(time.Since(start))
}

func (m *Manager) Close() {
	close(m.done)
}

func (m *Manager) CurrentConnections() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentOnline
}

func (m *Manager) Snapshot() Snapshot {
	m.mu.RLock()
	rooms := len(m.rooms)
	byType := make(map[string]int64, 4)
	for roomKey, rm := range m.rooms {
		rm.mu.Lock()
		size := len(rm.clients)
		rm.mu.Unlock()
		switch {
		case len(roomKey) >= 8 && roomKey[:8] == "article:":
			byType["article"] += int64(size)
		case len(roomKey) >= 7 && roomKey[:7] == "moment:":
			byType["moment"] += int64(size)
		case len(roomKey) >= 5 && roomKey[:5] == "page:":
			byType["page"] += int64(size)
		case len(roomKey) >= 11 && roomKey[:11] == "notif:user:":
			byType["notification"] += int64(size)
		default:
			byType["other"] += int64(size)
		}
	}
	current := m.currentOnline
	m.mu.RUnlock()

	broadcastTotal := atomic.LoadInt64(&m.broadcastTotal)
	broadcastErrors := atomic.LoadInt64(&m.broadcastErrors)
	broadcastFanout := atomic.LoadInt64(&m.broadcastFanout)
	avgRecipients := 0.0
	if broadcastTotal > 0 {
		avgRecipients = float64(broadcastFanout) / float64(broadcastTotal)
	}
	errorRate := 0.0
	if broadcastFanout > 0 {
		errorRate = float64(broadcastErrors) / float64(broadcastFanout)
	}

	return Snapshot{
		CurrentOnline:      current,
		Rooms:              rooms,
		JoinTotal:          atomic.LoadInt64(&m.joinTotal),
		LeaveTotal:         atomic.LoadInt64(&m.leaveTotal),
		BroadcastTotal:     broadcastTotal,
		BroadcastErrors:    broadcastErrors,
		BroadcastFanout:    broadcastFanout,
		BroadcastP95MS:     m.broadcastP95MS(),
		ByRoomType:         byType,
		AvgRecipients:      avgRecipients,
		BroadcastErrorRate: errorRate,
	}
}

func (c *Client) Write(payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupRooms()
		case <-m.done:
			return
		}
	}
}

func (m *Manager) cleanupRooms() {
	now := time.Now()
	m.mu.Lock()
	for key, rm := range m.rooms {
		rm.mu.Lock()
		expired := len(rm.clients) == 0 && now.Sub(rm.lastActivity) > m.roomTTL
		rm.mu.Unlock()
		if expired {
			delete(m.rooms, key)
		}
	}
	m.mu.Unlock()
}

var latencyBoundariesMS = []int64{5, 10, 25, 50, 100, 200, 500, 1000, 2000}

func (m *Manager) recordBroadcastLatency(d time.Duration) {
	ms := d.Milliseconds()
	idx := len(latencyBoundariesMS)
	for i, bound := range latencyBoundariesMS {
		if ms <= bound {
			idx = i
			break
		}
	}
	m.statsMu.Lock()
	m.latencyBins[idx]++
	m.statsMu.Unlock()
}

func (m *Manager) broadcastP95MS() float64 {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()
	var total int64
	for _, count := range m.latencyBins {
		total += count
	}
	if total == 0 {
		return 0
	}
	target := int64(float64(total) * 0.95)
	if target <= 0 {
		target = 1
	}
	var seen int64
	for i, count := range m.latencyBins {
		seen += count
		if seen >= target {
			if i >= len(latencyBoundariesMS) {
				return float64(latencyBoundariesMS[len(latencyBoundariesMS)-1])
			}
			return float64(latencyBoundariesMS[i])
		}
	}
	return float64(latencyBoundariesMS[len(latencyBoundariesMS)-1])
}
