import { browser } from '$app/environment';
import { getOwnerStatus } from '$lib/features/owner-status/api';
import type { OwnerStatus, OwnerStatusRealtimePayload } from '$lib/features/owner-status/types';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';

const defaultOwnerStatus: OwnerStatus = {
	ok: 0,
	process: '',
	extend: '',
	media: null,
	timestamp: 0,
	adminPanelOnline: false
};

class OwnerStatusStore {
	status = $state<OwnerStatus>({ ...defaultOwnerStatus });
	isConnected = $state(false);
	isLoading = $state(false);

	private started = false;
	private refreshTimer: ReturnType<typeof setInterval> | null = null;
	private unbindConnection: (() => void) | null = null;
	private unbindContent: (() => void) | null = null;

	start() {
		if (!browser || this.started) return;
		this.started = true;

		this.unbindConnection = realtimeWSCore.onConnection((connected) => {
			this.isConnected = connected;
		});

		this.unbindContent = realtimeWSCore.onContent((payload: unknown) => {
			if (!payload || typeof payload !== 'object') return;
			const event = payload as OwnerStatusRealtimePayload;
			if (event.type !== 'owner.status') return;
			this.applyStatus(event);
		});

		realtimeWSCore.start();
		void this.refreshNow();
		this.refreshTimer = setInterval(() => {
			void this.refreshNow();
		}, 30_000);
	}

	stop() {
		this.started = false;
		this.status = { ...defaultOwnerStatus };
		this.isConnected = false;
		this.isLoading = false;

		if (this.refreshTimer) {
			clearInterval(this.refreshTimer);
			this.refreshTimer = null;
		}

		this.unbindConnection?.();
		this.unbindConnection = null;
		this.unbindContent?.();
		this.unbindContent = null;
	}

	private async refreshNow() {
		if (!browser || !this.started) return;
		this.isLoading = true;
		try {
			const payload = await getOwnerStatus();
			this.applyStatus(payload);
		} catch {
			// ignore network errors; realtime updates and next poll will recover state
		} finally {
			this.isLoading = false;
		}
	}

	private applyStatus(payload: OwnerStatus) {
		const rawOK = Number(payload.ok);
		const ok = Number.isFinite(rawOK) && rawOK > 0 ? 1 : 0;
		const timestamp =
			typeof payload.timestamp === 'number' && Number.isFinite(payload.timestamp)
				? payload.timestamp
				: 0;

		this.status = {
			ok,
			process: typeof payload.process === 'string' ? payload.process : '',
			extend: typeof payload.extend === 'string' ? payload.extend : '',
			media: payload.media ?? null,
			timestamp,
			adminPanelOnline: payload.adminPanelOnline === true
		};
	}
}

export const ownerStatusStore = new OwnerStatusStore();
