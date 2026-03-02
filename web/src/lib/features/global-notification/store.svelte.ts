import { browser } from '$app/environment';
import { fetchActiveGlobalNotifications } from '$lib/features/global-notification/api';
import type {
	GlobalNotificationItem,
	GlobalNotificationRealtimePayload,
	GlobalNotificationRealtimeUpsertPayload
} from '$lib/features/global-notification/types';
import GlobalNotificationToastBody from '$lib/features/global-notification/components/GlobalNotificationToastBody.svelte';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
import { toast } from 'svelte-sonner';

const DISMISSED_STORAGE_KEY = 'grtblog:global-notification:dismissed:v1';

class GlobalNotificationStore {
	items = $state<GlobalNotificationItem[]>([]);
	current = $state<GlobalNotificationItem | null>(null);
	isConnected = $state(false);
	isLoading = $state(false);

	private started = false;
	private unbindConnection: (() => void) | null = null;
	private unbindContent: (() => void) | null = null;
	private tickTimer: ReturnType<typeof setInterval> | null = null;
	private refreshDebounceTimer: ReturnType<typeof setTimeout> | null = null;
	private activeToastId: string | number | null = null;
	private activeToastKey = '';
	private suppressedDismissToastIDs: Record<string, true> = {};
	private transientDismissedVersions: Record<string, string> = {};

	start() {
		if (!browser || this.started) return;
		this.started = true;

		this.unbindConnection = realtimeWSCore.onConnection((connected) => {
			const wasConnected = this.isConnected;
			this.isConnected = connected;
			if (connected && !wasConnected) {
				this.debouncedRefresh();
			}
		});

		this.unbindContent = realtimeWSCore.onContent((payload: unknown) => {
			this.handleRealtimePayload(payload);
		});

		realtimeWSCore.start();
		void this.refreshNow();
		this.tickTimer = setInterval(() => {
			this.syncCurrent();
		}, 1000);
	}

	stop() {
		this.started = false;
		this.items = [];
		this.current = null;
		this.isConnected = false;
		this.isLoading = false;
		this.transientDismissedVersions = {};

		if (this.tickTimer) {
			clearInterval(this.tickTimer);
			this.tickTimer = null;
		}
		if (this.refreshDebounceTimer) {
			clearTimeout(this.refreshDebounceTimer);
			this.refreshDebounceTimer = null;
		}

		this.unbindConnection?.();
		this.unbindConnection = null;
		this.unbindContent?.();
		this.unbindContent = null;
		this.clearToast();
	}

	dismissCurrent() {
		if (!browser || !this.current) return;
		this.dismissTransient(this.current);
		this.syncCurrent();
	}

	private debouncedRefresh() {
		if (this.refreshDebounceTimer) clearTimeout(this.refreshDebounceTimer);
		this.refreshDebounceTimer = setTimeout(() => {
			this.refreshDebounceTimer = null;
			void this.refreshNow();
		}, 2000);
	}

	private async refreshNow() {
		if (!browser || !this.started) return;
		this.isLoading = true;
		try {
			const list = await fetchActiveGlobalNotifications();
			this.items = this.normalizeItems(list);
			this.syncCurrent();
		} catch {
			// ignore network errors; realtime or next reconnect refresh will recover.
		} finally {
			this.isLoading = false;
		}
	}

	private handleRealtimePayload(payload: unknown) {
		if (!payload || typeof payload !== 'object') return;
		const event = payload as GlobalNotificationRealtimePayload;
		if (
			event.type !== 'global.notification.created' &&
			event.type !== 'global.notification.updated' &&
			event.type !== 'global.notification.deleted'
		) {
			return;
		}

		if (event.type === 'global.notification.deleted') {
			this.items = this.items.filter((item) => item.id !== event.id);
			this.syncCurrent();
			return;
		}

		const normalized = this.normalizeRealtimeUpsert(event);
		if (!normalized) return;

		if (!this.isActive(normalized)) {
			this.items = this.items.filter((item) => item.id !== normalized.id);
			this.syncCurrent();
			return;
		}

		const index = this.items.findIndex((item) => item.id === normalized.id);
		if (index >= 0) {
			const next = this.items.slice();
			next[index] = normalized;
			this.items = this.normalizeItems(next);
		} else {
			this.items = this.normalizeItems([...this.items, normalized]);
		}
		this.syncCurrent();
	}

	private normalizeRealtimeUpsert(
		payload: GlobalNotificationRealtimeUpsertPayload
	): GlobalNotificationItem | null {
		const id = Number(payload.id);
		if (!Number.isFinite(id) || id <= 0) return null;
		const content = typeof payload.content === 'string' ? payload.content.trim() : '';
		if (!content) return null;

		const publishAt = typeof payload.publishAt === 'string' ? payload.publishAt : '';
		const expireAt = typeof payload.expireAt === 'string' ? payload.expireAt : '';
		if (!publishAt || !expireAt) return null;

		const at = typeof payload.at === 'string' ? payload.at : '';

		return {
			id,
			content,
			publishAt,
			expireAt,
			allowClose: payload.allowClose === true,
			createdAt: at,
			updatedAt: at
		};
	}

	private normalizeItems(items: GlobalNotificationItem[]): GlobalNotificationItem[] {
		const now = Date.now();
		const dedup: Record<string, GlobalNotificationItem> = {};

		for (const raw of items) {
			const id = Number(raw.id);
			if (!Number.isFinite(id) || id <= 0) continue;

			const content = typeof raw.content === 'string' ? raw.content.trim() : '';
			if (!content) continue;

			const publishAt = typeof raw.publishAt === 'string' ? raw.publishAt : '';
			const expireAt = typeof raw.expireAt === 'string' ? raw.expireAt : '';
			if (!publishAt || !expireAt) continue;

			const item: GlobalNotificationItem = {
				id,
				content,
				publishAt,
				expireAt,
				allowClose: raw.allowClose === true,
				createdAt: typeof raw.createdAt === 'string' ? raw.createdAt : '',
				updatedAt: typeof raw.updatedAt === 'string' ? raw.updatedAt : ''
			};

			if (!this.isActive(item, now)) continue;
			dedup[String(id)] = item;
		}

		return Object.values(dedup).sort((a, b) => {
			const diff = this.parseTime(b.publishAt) - this.parseTime(a.publishAt);
			if (diff !== 0) return diff;
			return b.id - a.id;
		});
	}

	private syncCurrent() {
		if (!browser) {
			this.current = null;
			return;
		}

		const now = Date.now();
		this.items = this.items.filter((item) => this.isActive(item, now));
		this.current = this.items.find((item) => !this.isDismissed(item)) ?? null;
		this.syncToast();
	}

	private syncToast() {
		if (!browser) return;

		if (!this.current) {
			this.clearToast();
			return;
		}

		const item = this.current;
		const toastKey = `${item.id}:${this.dismissVersion(item)}`;
		if (toastKey === this.activeToastKey) {
			return;
		}

		this.clearToast();
		this.activeToastKey = toastKey;
		this.activeToastId = toast('全站通知', {
			description: GlobalNotificationToastBody,
			componentProps: {
				content: item.content,
				allowNeverShow: item.allowClose,
				onNeverShow: () => {
					this.dismissPersistent(item);
					this.dismissTransient(item);
					this.syncCurrent();
				}
			},
			duration: Infinity,
			closeButton: true,
			class: 'min-w-[18rem] max-w-[30rem]',
			classes: {
				toast: 'items-start',
				description: 'pr-3 whitespace-pre-wrap'
			},
			onDismiss: (dismissedToast) => {
				const dismissedID = this.toastIDKey(dismissedToast.id);
				if (this.suppressedDismissToastIDs[dismissedID]) {
					delete this.suppressedDismissToastIDs[dismissedID];
					return;
				}
				this.dismissTransient(item);
				this.syncCurrent();
			}
		});
	}

	private clearToast() {
		if (this.activeToastId !== null) {
			this.suppressedDismissToastIDs[this.toastIDKey(this.activeToastId)] = true;
			toast.dismiss(this.activeToastId);
			this.activeToastId = null;
		}
		this.activeToastKey = '';
	}

	private toastIDKey(id: string | number | undefined): string {
		return typeof id === 'number' || typeof id === 'string' ? String(id) : '';
	}

	private isActive(item: GlobalNotificationItem, now = Date.now()): boolean {
		const publishAt = this.parseTime(item.publishAt);
		const expireAt = this.parseTime(item.expireAt);
		if (publishAt <= 0 || expireAt <= 0) return false;
		if (publishAt > expireAt) return false;
		return publishAt <= now && expireAt >= now;
	}

	private parseTime(value: string): number {
		const timestamp = Date.parse(value);
		return Number.isFinite(timestamp) ? timestamp : 0;
	}

	private dismissPersistent(item: GlobalNotificationItem) {
		const map = this.readDismissed();
		map[String(item.id)] = this.dismissVersion(item);
		this.writeDismissed(map);
	}

	private dismissTransient(item: GlobalNotificationItem) {
		this.transientDismissedVersions[String(item.id)] = this.dismissVersion(item);
	}

	private isDismissed(item: GlobalNotificationItem): boolean {
		const transientVersion = this.transientDismissedVersions[String(item.id)];
		if (transientVersion && transientVersion === this.dismissVersion(item)) {
			return true;
		}

		const map = this.readDismissed();
		const dismissedVersion = map[String(item.id)];
		if (!dismissedVersion) return false;
		return dismissedVersion === this.dismissVersion(item);
	}

	private dismissVersion(item: GlobalNotificationItem): string {
		const version = (item.updatedAt || '').trim();
		if (version) return version;
		return `${item.publishAt}|${item.expireAt}|${item.content}`;
	}

	private readDismissed(): Record<string, string> {
		if (!browser) return {};
		try {
			const raw = localStorage.getItem(DISMISSED_STORAGE_KEY);
			if (!raw) return {};
			const parsed = JSON.parse(raw);
			if (!parsed || typeof parsed !== 'object') return {};
			const out: Record<string, string> = {};
			for (const [key, value] of Object.entries(parsed)) {
				if (typeof value !== 'string') continue;
				const id = Number(key);
				if (!Number.isFinite(id) || id <= 0) continue;
				out[String(id)] = value;
			}
			return out;
		} catch {
			return {};
		}
	}

	private writeDismissed(map: Record<string, string>) {
		if (!browser) return;
		try {
			localStorage.setItem(DISMISSED_STORAGE_KEY, JSON.stringify(map));
		} catch {
			// ignore persistence failures
		}
	}
}

export const globalNotificationStore = new GlobalNotificationStore();
