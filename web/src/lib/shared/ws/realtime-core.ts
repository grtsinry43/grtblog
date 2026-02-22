import { browser } from '$app/environment';
import type { PresenceClientReport, PresenceSnapshotPayload } from '$lib/features/presence/types';
import { getOrCreateVisitorId } from '$lib/shared/visitor/visitor-id';

export type ContentSubscription = {
	contentType: 'article' | 'moment' | 'page';
	contentId: number;
};

type ConnectionListener = (connected: boolean) => void;
type PresenceListener = (snapshot: PresenceSnapshotPayload) => void;
type ContentListener = (payload: unknown) => void;

class RealtimeWSCore {
	private socket: WebSocket | null = null;
	private started = false;
	private reconnectAttempts = 0;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private connected = false;

	private presenceReport: PresenceClientReport | null = null;
	private contentSubscription: ContentSubscription | null = null;
	private sentPresenceKey = '';
	private sentContentKey = '';
	private sentVisitorId = '';

	private connectionListeners = new Set<ConnectionListener>();
	private presenceListeners = new Set<PresenceListener>();
	private contentListeners = new Set<ContentListener>();

	start() {
		if (!browser || this.started) return;
		this.started = true;
		this.connect();
	}

	stop() {
		this.started = false;
		this.reconnectAttempts = 0;
		this.sentPresenceKey = '';
		this.sentContentKey = '';
		this.sentVisitorId = '';

		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}

		if (this.socket) {
			const active = this.socket;
			this.socket = null;
			active.close();
		}

		this.setConnected(false);
	}

	setPresenceReport(report: PresenceClientReport | null) {
		this.presenceReport = report;
		this.flushPresenceReport();
	}

	setContentSubscription(subscription: ContentSubscription | null) {
		this.contentSubscription = subscription;
		this.flushContentSubscription();
	}

	onConnection(listener: ConnectionListener): () => void {
		this.connectionListeners.add(listener);
		listener(this.connected);
		return () => {
			this.connectionListeners.delete(listener);
		};
	}

	onPresenceSnapshot(listener: PresenceListener): () => void {
		this.presenceListeners.add(listener);
		return () => {
			this.presenceListeners.delete(listener);
		};
	}

	onContent(listener: ContentListener): () => void {
		this.contentListeners.add(listener);
		return () => {
			this.contentListeners.delete(listener);
		};
	}

	private connect() {
		if (!browser || !this.started) return;

		const wsUrl = new URL('/api/v2/ws/realtime', window.location.origin);
		wsUrl.protocol = wsUrl.protocol === 'https:' ? 'wss:' : 'ws:';

		const socket = new WebSocket(wsUrl.toString());
		this.socket = socket;

		socket.onopen = () => {
			if (this.socket !== socket) return;
			this.reconnectAttempts = 0;
			this.sentPresenceKey = '';
			this.sentContentKey = '';
			this.sentVisitorId = '';
			this.setConnected(true);
			this.flushPresenceIdentify();
			this.flushPresenceReport();
			this.flushContentSubscription();
		};

		socket.onmessage = (event) => {
			if (this.socket !== socket) return;

			let payload: unknown;
			try {
				payload = JSON.parse(event.data);
			} catch {
				return;
			}

			if (
				payload &&
				typeof payload === 'object' &&
				'type' in payload &&
				(payload as { type?: string }).type === 'presence.snapshot'
			) {
				const snapshot = payload as PresenceSnapshotPayload;
				for (const listener of this.presenceListeners) {
					listener(snapshot);
				}
				return;
			}

			for (const listener of this.contentListeners) {
				listener(payload);
			}
		};

		socket.onerror = () => {
			socket.close();
		};

		socket.onclose = () => {
			if (this.socket !== socket) return;
			this.socket = null;
			this.setConnected(false);
			if (!this.started) return;
			this.scheduleReconnect();
		};
	}

	private scheduleReconnect() {
		if (this.reconnectTimer || !this.started) return;

		const delay = Math.min(1000 * 2 ** Math.min(this.reconnectAttempts, 4), 15000);
		this.reconnectAttempts += 1;
		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null;
			this.connect();
		}, delay);
	}

	private setConnected(connected: boolean) {
		if (this.connected === connected) return;
		this.connected = connected;
		for (const listener of this.connectionListeners) {
			listener(connected);
		}
	}

	private flushPresenceReport() {
		if (!this.presenceReport) return;
		if (!this.isSocketOpen()) return;
		this.flushPresenceIdentify();

		const key = `${this.presenceReport.contentType}|${this.presenceReport.url}`;
		if (key === this.sentPresenceKey) return;
		const visitorId = this.getVisitorId();

		this.socket?.send(
			JSON.stringify({
				type: 'presence.report',
				contentType: this.presenceReport.contentType,
				url: this.presenceReport.url,
				visitorId: visitorId || undefined
			})
		);
		this.sentPresenceKey = key;
	}

	private flushPresenceIdentify() {
		if (!this.isSocketOpen()) return;

		const visitorId = this.getVisitorId();
		if (!visitorId || visitorId === this.sentVisitorId) return;

		this.socket?.send(
			JSON.stringify({
				type: 'presence.identify',
				visitorId
			})
		);
		this.sentVisitorId = visitorId;
	}

	private getVisitorId(): string {
		return getOrCreateVisitorId().trim().slice(0, 255);
	}

	private flushContentSubscription() {
		if (!this.isSocketOpen()) return;

		if (!this.contentSubscription) {
			if (!this.sentContentKey) return;
			this.socket?.send(JSON.stringify({ type: 'content.unsubscribe' }));
			this.sentContentKey = '';
			return;
		}

		const key = `${this.contentSubscription.contentType}:${this.contentSubscription.contentId}`;
		if (key === this.sentContentKey) return;

		this.socket?.send(
			JSON.stringify({
				type: 'content.subscribe',
				contentType: this.contentSubscription.contentType,
				contentId: this.contentSubscription.contentId
			})
		);
		this.sentContentKey = key;
	}

	private isSocketOpen(): boolean {
		return !!this.socket && this.socket.readyState === WebSocket.OPEN;
	}
}

export const realtimeWSCore = new RealtimeWSCore();
