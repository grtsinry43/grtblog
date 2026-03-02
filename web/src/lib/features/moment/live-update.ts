import { browser } from '$app/environment';
import { checkMomentLatest } from '$lib/features/moment/api';
import type { MomentContentPayload, MomentDetail } from '$lib/features/moment/types';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
import { toast } from 'svelte-sonner';

export type MomentLiveUpdateCallbacks = {
	getId: () => number | null;
	getContentHash: () => string | null;
	updateMoment: (updater: (prev: MomentDetail | null) => MomentDetail | null) => void;
};

export function createMomentLiveUpdate(callbacks: MomentLiveUpdateCallbacks) {
	if (!browser) return { start() {}, destroy() {} };

	let unsubscribeContent: (() => void) | null = null;

	const triggerUpdateHint = () => {
		toast.success('手记已更新，内容已自动刷新。', { duration: 3000 });
	};

	const applyPayload = (payload: MomentContentPayload) => {
		callbacks.updateMoment((prev) => {
			if (!prev) return prev;
			return {
				...prev,
				title: payload.title ?? prev.title,
				summary: payload.summary ?? prev.summary,
				toc: payload.toc ?? prev.toc,
				content: payload.content ?? prev.content,
				contentHash: payload.contentHash || prev.contentHash
			};
		});
		triggerUpdateHint();
	};

	const refreshIfNeeded = async () => {
		const id = callbacks.getId();
		const contentHash = callbacks.getContentHash();
		if (!id || !contentHash) return;

		try {
			const latest = await checkMomentLatest(undefined, id, contentHash);
			if (!latest || latest.latest) return;

			applyPayload({
				contentHash: latest.contentHash,
				title: latest.title,
				summary: latest.summary,
				toc: latest.toc,
				content: latest.content
			});
		} catch {
			toast.error('检查手记更新时失败，请稍后重试', { duration: 5000 });
		}
	};

	const subscribe = (momentId: number) => {
		realtimeWSCore.start();
		realtimeWSCore.setContentSubscription({ contentType: 'moment', contentId: momentId });

		unsubscribeContent?.();
		unsubscribeContent = realtimeWSCore.onContent((data: unknown) => {
			if (!data || typeof data !== 'object') return;

			const payload = data as MomentContentPayload;
			if (!payload?.contentHash) return;

			const currentHash = callbacks.getContentHash();
			if (currentHash && currentHash === payload.contentHash) return;

			applyPayload(payload);
		});
	};

	return {
		start(momentId: number) {
			subscribe(momentId);
			void refreshIfNeeded();
		},
		destroy() {
			unsubscribeContent?.();
			unsubscribeContent = null;
			realtimeWSCore.setContentSubscription(null);
		}
	};
}
