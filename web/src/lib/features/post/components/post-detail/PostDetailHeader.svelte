<script lang="ts">
	import { resolve } from '$app/paths';
	import { postDetailCtx } from '$lib/features/post/context';
	import { sameMetrics } from './selector-equals';
	import { formatDateCN } from '$lib/shared/utils/date';
	import { ArrowLeft, Calendar, Clock } from 'lucide-svelte';
	import Icon from '@iconify/svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import Badge from '$lib/ui/primitives/badge/Badge.svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import TagList from '$lib/features/tag/components/TagList.svelte';
	import { buildCategoryPath } from '$lib/shared/utils/content-path';

	const titleStore = postDetailCtx.selectModelData((data) => data?.title ?? '');
	const postIdStore = postDetailCtx.selectModelData((data) => data?.id ?? 0);
	const createdAtStore = postDetailCtx.selectModelData((data) => data?.createdAt ?? '');
	const isHotStore = postDetailCtx.selectModelData((data) => data?.isHot ?? false);
	const metricsStore = postDetailCtx.selectModelData((data) => data?.metrics ?? null, {
		equals: sameMetrics
	});
	const tagsStore = postDetailCtx.selectModelData((data) => data?.tags ?? []);
	const categoryNameStore = postDetailCtx.selectModelData((data) => data?.categoryName ?? '');
	const categoryShortUrlStore = postDetailCtx.selectModelData(
		(data) => data?.categoryShortUrl ?? ''
	);
	const categoryLabelStore = $derived.by(() => {
		const categoryName = ($categoryNameStore || '').trim();
		return categoryName || '未分类';
	});

	function goBack() {
		history.back();
	}
</script>

{#snippet backContent()}
	<ArrowLeft size={14} class="group-hover:-translate-x-1 transition-transform" />
	<span>返回</span>
{/snippet}

<header class="max-w-4xl space-y-6">
	<div class="flex items-center gap-4">
		<Button
			variant="ghost"
			class="!h-auto !p-0 font-mono text-[10px] font-semibold tracking-[0.2em] text-ink-400 uppercase hover:!bg-transparent hover:text-ink-900 group"
			onclick={goBack}
			content={backContent}
		/>
		<div class="h-px w-6 bg-ink-200/50 dark:bg-ink-800/50"></div>
	</div>

	<div class="space-y-4">
		<div class="flex items-center gap-3">
			<Badge variant="soft">文章</Badge>
			{#if $categoryShortUrlStore}
				<a
					href={resolve(buildCategoryPath($categoryShortUrlStore))}
					class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
				>
					{categoryLabelStore}
				</a>
			{:else}
				<span class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase"
					>{categoryLabelStore}</span
				>
			{/if}
		</div>

		<h1
			class="font-serif text-2xl leading-[1.2] font-medium tracking-tight text-ink-950 md:text-3xl lg:text-4xl dark:text-ink-50"
		>
			{$titleStore}
		</h1>

		<div class="flex flex-col gap-4">
			<div
				class="flex flex-wrap items-center gap-5 font-mono text-[9px] tracking-widest text-ink-400 uppercase"
			>
				{#if $isHotStore}
					{#snippet hotIcon()}
						<Icon icon="ph:fire-fill" class="size-4 text-red-500" />
					{/snippet}
					<Badge
						variant="soft"
						class="!border-red-500/20 !bg-red-500/5 !text-red-600 dark:!text-red-400"
						icon={hotIcon}
					>
						热门
					</Badge>
				{/if}
				<span class="flex items-center gap-1.5">
					<Calendar size={12} />
					{formatDateCN($createdAtStore)}
				</span>
				<span class="flex items-center gap-1.5"><Clock size={12} /> 12 分钟阅读</span>
				<span class="flex items-center gap-1.5">浏览 {$metricsStore?.views ?? 0}</span>
				<span aria-hidden="true" class="opacity-40">·</span>
				<ContentLikeButton
					contentType="article"
					contentId={$postIdStore}
					likes={$metricsStore?.likes ?? 0}
					className="inline-flex items-center gap-1.5"
				/>
				<span aria-hidden="true" class="opacity-40">·</span>
				<span class="flex items-center gap-1.5">评论 {$metricsStore?.comments ?? 0}</span>
			</div>

			<TagList tags={$tagsStore} />
		</div>
	</div>
</header>
