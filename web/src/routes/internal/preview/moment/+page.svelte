<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import MomentDetail from '$lib/features/moment/components/MomentDetail.svelte';
	import type { MomentDetail as MomentDetailModel, TOCNode, TopicTag } from '$lib/features/moment/types';
	import { windowMessage } from '$lib/shared/actions/window-message';
	import { momentDetailCtx } from '$lib/features/moment/context';

	type PreviewMomentPayload = {
		id?: number;
		title?: string;
		summary?: string;
		aiSummary?: string | null;
		content?: string;
		contentHash?: string;
		shortUrl?: string;
		image?: string[];
		columnId?: number | null;
		columnName?: string;
		columnShortUrl?: string;
		commentAreaId?: number | null;
		toc?: TOCNode[];
		topics?: TopicTag[];
		metrics?: MomentDetailModel['metrics'];
		isPublished?: boolean;
		isTop?: boolean;
		isHot?: boolean;
		isOriginal?: boolean;
		createdAt?: string;
		updatedAt?: string;
		authorId?: number;
	};

	momentDetailCtx.mountModelData(() => null);
	const { updateModelData } = momentDetailCtx.useModelActions();
	const momentStore = momentDetailCtx.selectModelData((d) => d);

	let hasPayload = $state(false);

	const applyPayload = (payload: PreviewMomentPayload) => {
		const nowIso = new Date().toISOString();
		updateModelData(() => ({
			id: payload.id ?? 0,
			title: payload.title ?? '未命名',
			summary: payload.summary ?? '',
			aiSummary: payload.aiSummary ?? null,
			content: payload.content ?? '',
			contentHash: payload.contentHash ?? '',
			toc: payload.toc,
			authorId: payload.authorId ?? 0,
			shortUrl: payload.shortUrl ?? '',
			image: payload.image ?? [],
			columnId: payload.columnId ?? null,
			columnName: payload.columnName ?? undefined,
			columnShortUrl: payload.columnShortUrl ?? undefined,
			commentAreaId: payload.commentAreaId ?? null,
			isPublished: payload.isPublished ?? false,
			topics: payload.topics ?? [],
			metrics: payload.metrics,
			isTop: payload.isTop ?? false,
			isHot: payload.isHot ?? false,
			isOriginal: payload.isOriginal ?? true,
			createdAt: payload.createdAt ?? nowIso,
			updatedAt: payload.updatedAt ?? nowIso,
			relatedPosts: []
		}));
		hasPayload = true;
	};

	const handleMessage = (event: MessageEvent) => {
		if (!browser) return;
		if (event.source !== window.parent) return;
		const data = event.data as { type?: string; payload?: PreviewMomentPayload } | null;
		if (!data || data.type !== 'grtblog-preview:moment') return;
		applyPayload(data.payload ?? {});
	};

	onMount(() => {
		if (!browser) return;
		window.parent?.postMessage({ type: 'grtblog-preview:ready' }, '*');
	});
</script>

<div class="min-h-screen" use:windowMessage={{ handler: handleMessage }}>
	{#if hasPayload && $momentStore}
		<MomentDetail moment={$momentStore} />
	{:else}
		<div class="flex min-h-screen items-center justify-center text-sm text-ink-400">
			等待预览数据...
		</div>
	{/if}
</div>
