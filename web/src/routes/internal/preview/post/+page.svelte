<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import PostDetail from '$lib/features/post/components/PostDetail.svelte';
	import type { ContentExtInfo } from '$lib/shared/markdown/image-ext-info';
	import type { PostDetail as PostDetailModel, TOCNode, Tag } from '$lib/features/post/types';
	import { windowMessage } from '$lib/shared/actions/window-message';
	import { postDetailCtx } from '$lib/features/post/context';

	type PreviewPostPayload = {
		id?: number;
		title?: string;
		summary?: string;
		leadIn?: string | null;
		content?: string;
		contentHash?: string;
		shortUrl?: string;
		cover?: string | null;
		categoryId?: number | null;
		commentAreaId?: number | null;
		extInfo?: ContentExtInfo | null;
		toc?: TOCNode[];
		tags?: Tag[];
		metrics?: PostDetailModel['metrics'];
		isPublished?: boolean;
		isTop?: boolean;
		isHot?: boolean;
		isOriginal?: boolean;
		createdAt?: string;
		updatedAt?: string;
		authorId?: number;
	};

	postDetailCtx.mountModelData(() => null);
	const { updateModelData } = postDetailCtx.useModelActions();

	let hasPayload = $state(false);

	const applyPayload = (payload: PreviewPostPayload) => {
		const nowIso = new Date().toISOString();
		updateModelData(() => ({
			id: payload.id ?? 0,
			title: payload.title ?? '未命名',
			summary: payload.summary ?? '',
			aiSummary: null,
			content: payload.content ?? '',
			contentHash: payload.contentHash ?? '',
			leadIn: payload.leadIn ?? null,
			toc: payload.toc,
			authorId: payload.authorId ?? 0,
			shortUrl: payload.shortUrl ?? '',
			cover: payload.cover ?? undefined,
			categoryId: payload.categoryId ?? null,
			commentAreaId: payload.commentAreaId ?? null,
			extInfo: payload.extInfo ?? null,
			isPublished: payload.isPublished ?? false,
			tags: payload.tags ?? [],
			metrics: payload.metrics,
			isTop: payload.isTop ?? false,
			isHot: payload.isHot ?? false,
			isOriginal: payload.isOriginal ?? true,
			createdAt: payload.createdAt ?? nowIso,
			updatedAt: payload.updatedAt ?? nowIso
		}));
		hasPayload = true;
	};

	const handleMessage = (event: MessageEvent) => {
		if (!browser) return;
		if (event.source !== window.parent) return;
		const data = event.data as { type?: string; payload?: PreviewPostPayload } | null;
		if (!data || data.type !== 'grtblog-preview:post') return;
		applyPayload(data.payload ?? {});
	};

	onMount(() => {
		if (!browser) return;
		window.parent?.postMessage({ type: 'grtblog-preview:ready' }, '*');
	});
</script>

<div class="min-h-screen" use:windowMessage={{ handler: handleMessage }}>
	{#if hasPayload}
		<PostDetail />
	{:else}
		<div class="flex min-h-screen items-center justify-center text-sm text-ink-400">
			等待预览数据...
		</div>
	{/if}
</div>
