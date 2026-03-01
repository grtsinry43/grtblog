import { browser } from '$app/environment';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';

export type SystemMode = 'healthy' | 'maintenance' | 'degraded' | 'critical' | 'outage';

export type HealthSSRData = {
	maintenance: boolean;
	healthMode: string;
	isDev: boolean;
};

type HealthWSPayload = {
	type: 'system.health.state';
	healthBits: number;
	maintenance: boolean;
	mode: SystemMode;
	components: Record<string, boolean>;
	isDev: boolean;
	timestamp: string;
};

class SiteHealthStore {
	mode = $state<SystemMode>('healthy');
	maintenance = $state(false);
	isDev = $state(false);
	healthBits = $state(63);

	private started = false;
	private unbindContent: (() => void) | null = null;

	initFromSSR(data: HealthSSRData) {
		this.maintenance = data.maintenance ?? false;
		this.isDev = data.isDev ?? false;
		const rawMode = data.healthMode as SystemMode;
		if (rawMode && ['healthy', 'maintenance', 'degraded', 'critical', 'outage'].includes(rawMode)) {
			this.mode = rawMode;
		}
	}

	handleWSMessage(payload: unknown) {
		if (!payload || typeof payload !== 'object') return;
		const msg = payload as HealthWSPayload;
		if (msg.type !== 'system.health.state') return;

		this.healthBits = typeof msg.healthBits === 'number' ? msg.healthBits : 63;
		this.maintenance = msg.maintenance === true;
		this.isDev = msg.isDev === true;

		const rawMode = msg.mode as SystemMode;
		if (rawMode && ['healthy', 'maintenance', 'degraded', 'critical', 'outage'].includes(rawMode)) {
			this.mode = rawMode;
		}
	}

	start() {
		if (!browser || this.started) return;
		this.started = true;

		this.unbindContent = realtimeWSCore.onContent((payload: unknown) => {
			this.handleWSMessage(payload);
		});

		realtimeWSCore.start();
	}

	stop() {
		this.started = false;
		this.unbindContent?.();
		this.unbindContent = null;
	}

	get showBanner(): boolean {
		if (this.isDev) return false;
		return this.mode !== 'healthy';
	}
}

export const siteHealthStore = new SiteHealthStore();
