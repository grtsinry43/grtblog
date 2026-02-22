const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '/api/v2').replace(/\/$/, '')

type ConnectionListener = (connected: boolean) => void
type MessageListener = (payload: unknown) => void

class AdminRealtimeWSCore {
  private socket: WebSocket | null = null
  private started = false
  private connected = false
  private reconnectAttempts = 0
  private reconnectTimer: number | null = null
  private panelPingTimer: number | null = null

  private jwtToken: string | null = null
  private panelHeartbeatEnabled = false

  private connectionListeners = new Set<ConnectionListener>()
  private messageListeners = new Set<MessageListener>()

  start() {
    if (this.started) return
    this.started = true
    this.connect()
  }

  stop() {
    this.started = false
    this.reconnectAttempts = 0
    this.clearReconnectTimer()
    this.clearPanelPingTimer()
    if (this.socket) {
      const active = this.socket
      this.socket = null
      active.close(1000, 'stop')
    }
    this.setConnected(false)
  }

  updateToken(token: string | null | undefined) {
    const normalized = token?.trim() || null
    if (this.jwtToken === normalized) return
    this.jwtToken = normalized
    if (!this.started) return
    this.refreshConnection()
  }

  setPanelHeartbeat(enabled: boolean) {
    if (this.panelHeartbeatEnabled === enabled) return
    this.panelHeartbeatEnabled = enabled
    if (!enabled) {
      this.clearPanelPingTimer()
      return
    }
    this.sendPanelPing()
    this.ensurePanelPingTimer()
  }

  onConnection(listener: ConnectionListener): () => void {
    this.connectionListeners.add(listener)
    listener(this.connected)
    return () => {
      this.connectionListeners.delete(listener)
    }
  }

  onMessage(listener: MessageListener): () => void {
    this.messageListeners.add(listener)
    return () => {
      this.messageListeners.delete(listener)
    }
  }

  private connect() {
    if (!this.started) return
    const socket = this.createSocket()
    this.socket = socket

    socket.onopen = () => {
      if (this.socket !== socket) return
      this.reconnectAttempts = 0
      this.clearReconnectTimer()
      this.setConnected(true)
      this.sendPanelPing()
      this.ensurePanelPingTimer()
    }

    socket.onmessage = (event) => {
      if (this.socket !== socket) return
      let payload: unknown
      try {
        payload = JSON.parse(event.data)
      } catch {
        return
      }
      for (const listener of this.messageListeners) {
        listener(payload)
      }
    }

    socket.onerror = () => {
      socket.close()
    }

    socket.onclose = () => {
      if (this.socket !== socket) return
      this.socket = null
      this.clearPanelPingTimer()
      this.setConnected(false)
      if (!this.started) return
      this.scheduleReconnect()
    }
  }

  private refreshConnection() {
    this.clearReconnectTimer()
    this.reconnectAttempts = 0
    if (this.socket) {
      const active = this.socket
      this.socket = null
      active.close(1000, 'refresh')
    }
    this.clearPanelPingTimer()
    this.setConnected(false)
    this.connect()
  }

  private createSocket(): WebSocket {
    const wsUrl = new URL(API_BASE_URL, window.location.origin)
    wsUrl.protocol = wsUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    wsUrl.pathname = `${wsUrl.pathname.replace(/\/$/, '')}/ws/realtime`
    wsUrl.search = ''

    if (this.jwtToken) {
      return new WebSocket(wsUrl.toString(), ['grtblog.jwt', this.jwtToken])
    }
    return new WebSocket(wsUrl.toString())
  }

  private ensurePanelPingTimer() {
    if (!this.panelHeartbeatEnabled || !this.isSocketOpen()) return
    this.clearPanelPingTimer()
    this.panelPingTimer = window.setInterval(() => {
      this.sendPanelPing()
    }, 20_000)
  }

  private sendPanelPing() {
    if (!this.panelHeartbeatEnabled || !this.isSocketOpen()) return
    this.socket?.send(
      JSON.stringify({
        type: 'owner.panel.ping',
      }),
    )
  }

  private scheduleReconnect() {
    if (this.reconnectTimer != null || !this.started) return
    const delay = Math.min(1000 * 2 ** Math.min(this.reconnectAttempts, 4), 15_000)
    this.reconnectAttempts += 1
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer == null) return
    window.clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
  }

  private clearPanelPingTimer() {
    if (this.panelPingTimer == null) return
    window.clearInterval(this.panelPingTimer)
    this.panelPingTimer = null
  }

  private isSocketOpen(): boolean {
    return !!this.socket && this.socket.readyState === WebSocket.OPEN
  }

  private setConnected(connected: boolean) {
    if (this.connected === connected) return
    this.connected = connected
    for (const listener of this.connectionListeners) {
      listener(connected)
    }
  }
}

export const adminRealtimeWSCore = new AdminRealtimeWSCore()
