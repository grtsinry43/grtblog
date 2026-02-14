import { browser } from '$app/environment';
import type {
	PresenceClientReport,
	PresencePageItem,
	PresenceSnapshotPayload
} from '$lib/features/presence/types';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';

class PresenceStore {
	online = $state(0);
	pages = $state<PresencePageItem[]>([]);
	isConnected = $state(false);

	private started = false;
	private unbindConnection: (() => void) | null = null;
	private unbindPresence: (() => void) | null = null;
	private lastReportKey = '';

	start() {
		if (!browser || this.started) return;
		this.started = true;
		this.unbindConnection = realtimeWSCore.onConnection((connected) => {
			this.isConnected = connected;
		});
		this.unbindPresence = realtimeWSCore.onPresenceSnapshot((payload: PresenceSnapshotPayload) => {
			this.online = Number.isFinite(payload.online) ? Math.max(0, payload.online) : 0;
			this.pages = Array.isArray(payload.pages) ? payload.pages : [];
		});
		realtimeWSCore.start();
	}

	reportView(report: PresenceClientReport | null) {
		if (!report) return;

		const normalized: PresenceClientReport = {
			contentType: report.contentType,
			url: report.url || '/'
		};
		const reportKey = `${normalized.contentType}|${normalized.url}`;
		if (reportKey === this.lastReportKey) return;

		this.lastReportKey = reportKey;
		realtimeWSCore.setPresenceReport(normalized);
	}

	stop() {
		this.started = false;
		this.online = 0;
		this.pages = [];
		this.isConnected = false;
		this.lastReportKey = '';

		this.unbindConnection?.();
		this.unbindConnection = null;
		this.unbindPresence?.();
		this.unbindPresence = null;

		realtimeWSCore.setPresenceReport(null);
		realtimeWSCore.stop();
	}
}

export const presenceStore = new PresenceStore();
