import { browser } from '$app/environment';
import { checkPostLatest } from '$lib/features/post/api';
import type { PostContentPayload, PostDetail } from '$lib/features/post/types';
import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
import { toast } from 'svelte-sonner';

export type PostLiveUpdateCallbacks = {
	getId: () => number | null;
	getContentHash: () => string | null;
	updatePost: (updater: (prev: PostDetail | null) => PostDetail | null) => void;
};

/**
 * Creates a live-update controller for a post detail page.
 * Handles: WebSocket connection for real-time pushes + polling via checkPostLatest.
 */
export function createPostLiveUpdate(callbacks: PostLiveUpdateCallbacks) {
	if (!browser) return { start() {}, destroy() {} };

	let unsubscribeContent: (() => void) | null = null;

	const triggerUpdateHint = () => {
		toast.success('作者修改了内容，已为您自动更新了呀！', { duration: 3000 });
	};

	const applyPayload = (payload: PostContentPayload) => {
		callbacks.updatePost((prev) => {
			if (!prev) return prev;
			return {
				...prev,
				title: payload.title ?? prev.title,
				leadIn: payload.leadIn ?? prev.leadIn,
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
			const latest = await checkPostLatest(undefined, id, contentHash);
			if (!latest || latest.latest) return;

			applyPayload({
				contentHash: latest.contentHash,
				title: latest.title,
				leadIn: latest.leadIn,
				toc: latest.toc,
				content: latest.content
			});
		} catch {
			toast.error('检查文章更新时出错了，请检查您的网络连接', { duration: 5000 });
		}
	};

	const subscribe = (postId: number) => {
		realtimeWSCore.start();
		realtimeWSCore.setContentSubscription({ contentType: 'article', contentId: postId });

		unsubscribeContent?.();
		unsubscribeContent = realtimeWSCore.onContent((data: unknown) => {
			if (!data || typeof data !== 'object') {
				return;
			}

			const payload = data as PostContentPayload;
			if (!payload?.contentHash) return;

			const currentHash = callbacks.getContentHash();
			if (currentHash && currentHash === payload.contentHash) return;

			applyPayload(payload);
		});
	};

	return {
		start(postId: number) {
			subscribe(postId);
			void refreshIfNeeded();
		},
		destroy() {
			unsubscribeContent?.();
			unsubscribeContent = null;
			realtimeWSCore.setContentSubscription(null);
		}
	};
}
