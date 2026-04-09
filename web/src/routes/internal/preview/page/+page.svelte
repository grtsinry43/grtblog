<script lang="ts">
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import PageDetail from '$lib/features/page/components/PageDetail.svelte';
	import type { PageDetail as PageDetailModel, TOCNode } from '$lib/features/page/types';
	import { windowMessage } from '$lib/shared/actions/window-message';

	type PreviewPagePayload = {
		id?: number;
		title?: string;
		description?: string | null;
		aiSummary?: string | null;
		toc?: TOCNode[];
		content?: string;
		contentHash?: string;
		commentAreaId?: number | null;
		shortUrl?: string;
		isEnabled?: boolean;
		isBuiltin?: boolean;
		metrics?: PageDetailModel['metrics'];
		createdAt?: string;
		updatedAt?: string;
	};

	let pageModel = $state<PageDetailModel | null>(null);

	const applyPayload = (payload: PreviewPagePayload) => {
		const nowIso = new Date().toISOString();
		pageModel = {
			id: payload.id ?? 0,
			title: payload.title ?? '未命名页面',
			description: payload.description ?? null,
			aiSummary: payload.aiSummary ?? null,
			toc: payload.toc,
			content: payload.content ?? '',
			contentHash: payload.contentHash ?? '',
			commentAreaId: payload.commentAreaId ?? null,
			shortUrl: payload.shortUrl ?? '',
			isEnabled: payload.isEnabled ?? true,
			isBuiltin: payload.isBuiltin ?? false,
			metrics: payload.metrics,
			contentUpdatedAt: payload.updatedAt ?? nowIso,
			createdAt: payload.createdAt ?? nowIso,
			updatedAt: payload.updatedAt ?? nowIso
		};
	};

	const handleMessage = (event: MessageEvent) => {
		if (!browser) return;
		if (event.source !== window.parent) return;
		const data = event.data as { type?: string; payload?: PreviewPagePayload } | null;
		if (!data || data.type !== 'grtblog-preview:page') return;
		applyPayload(data.payload ?? {});
	};

	onMount(() => {
		if (!browser) return;
		window.parent?.postMessage({ type: 'grtblog-preview:ready' }, '*');
	});
</script>

<div class="min-h-screen" use:windowMessage={{ handler: handleMessage }}>
	{#if pageModel}
		<PageDetail page={pageModel} />
	{:else}
		<div class="flex min-h-screen items-center justify-center text-sm text-ink-400">
			等待预览数据...
		</div>
	{/if}
</div>
