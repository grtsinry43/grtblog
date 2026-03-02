import { browser } from '$app/environment';
import { checkPageLatest } from '$lib/features/page/api';
import type { PageContentPayload, PageDetail } from '$lib/features/page/types';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
import { toast } from 'svelte-sonner';

export type PageLiveUpdateCallbacks = {
	getId: () => number | null;
	getContentHash: () => string | null;
	updatePage: (updater: (prev: PageDetail | null) => PageDetail | null) => void;
};

export function createPageLiveUpdate(callbacks: PageLiveUpdateCallbacks) {
	if (!browser) return { start() {}, destroy() {} };

	let unsubscribeContent: (() => void) | null = null;

	const triggerUpdateHint = () => {
		toast.success('页面内容有更新，已自动同步。', { duration: 3000 });
	};

	const applyPayload = (payload: PageContentPayload) => {
		callbacks.updatePage((prev) => {
			if (!prev) return prev;
			return {
				...prev,
				title: payload.title ?? prev.title,
				description: payload.description ?? prev.description,
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
			const latest = await checkPageLatest(undefined, id, contentHash);
			if (!latest || latest.latest) return;

			applyPayload({
				contentHash: latest.contentHash,
				title: latest.title,
				description: latest.description,
				toc: latest.toc,
				content: latest.content
			});
		} catch {
			toast.error('检查页面更新时失败，请稍后重试', { duration: 5000 });
		}
	};

	const subscribe = (pageId: number) => {
		realtimeWSCore.start();
		realtimeWSCore.setContentSubscription({ contentType: 'page', contentId: pageId });

		unsubscribeContent?.();
		unsubscribeContent = realtimeWSCore.onContent((data: unknown) => {
			if (!data || typeof data !== 'object') return;

			const payload = data as PageContentPayload;
			if (!payload?.contentHash) return;

			const currentHash = callbacks.getContentHash();
			if (currentHash && currentHash === payload.contentHash) return;

			applyPayload(payload);
		});
	};

	return {
		start(pageId: number) {
			subscribe(pageId);
			void refreshIfNeeded();
		},
		destroy() {
			unsubscribeContent?.();
			unsubscribeContent = null;
			realtimeWSCore.setContentSubscription(null);
		}
	};
}
