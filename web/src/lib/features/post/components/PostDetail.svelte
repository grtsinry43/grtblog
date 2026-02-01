<script lang="ts">
	import { onDestroy, tick } from 'svelte';
	import type { PostDetail } from '$lib/features/post/types';
	import { renderMarkdown } from '$lib/shared/markdown/markdown';
	import type { TOCNode } from '$lib/features/post/types';
	import { markdownComponents } from '$lib/shared/actions/markdown-components';
	import { buildImageExtInfoState, imageExtInfoCtx } from '$lib/shared/markdown/image-ext-info';
	import { Calendar, Clock, Share2, ArrowLeft } from 'lucide-svelte';
	import { page } from '$app/stores';
	import Button from '$lib/ui/ui/button/Button.svelte';
	import Badge from '$lib/ui/ui/badge/Badge.svelte';
	import Tag from '$lib/ui/ui/tag/Tag.svelte';
	import Loading from '$lib/ui/common/Loading.svelte';
	import '$lib/ui/markdown/register';
	import { postDetailCtx } from '$routes/posts/[id]/post-detail-context';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';

	const postStore = postDetailCtx.selectModelData((data) => data as PostDetail | null);
	const imageExtInfoStore = imageExtInfoCtx.mountModelData(
		buildImageExtInfoState($postStore?.extInfo ?? null)
	);
	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor = $state<string | null>(null);
	let observer: IntersectionObserver | null = null;
	const siteOrigin = $derived($page.url.origin);

	const flattenTOC = (nodes?: TOCNode[]) => {
		if (!nodes?.length) return [];
		const anchors: string[] = [];
		const walk = (items: TOCNode[]) => {
			for (const item of items) {
				anchors.push(item.anchor);
				if (item.children?.length) walk(item.children);
			}
		};
		walk(nodes);
		return anchors;
	};

	let contentHtml = $derived(
		$postStore
			? renderMarkdown($postStore.content ?? '', flattenTOC($postStore.toc), {
					origin: siteOrigin
				})
			: ''
	);

	const setupObserver = () => {
		if (!contentRoot || typeof IntersectionObserver === 'undefined') return;
		observer?.disconnect();
		const headings = contentRoot.querySelectorAll('h1, h2, h3, h4, h5, h6');
		if (!headings.length) {
			activeAnchor = null;
			return;
		}
		observer = new IntersectionObserver(
			(entries) => {
				const visible = entries.filter((entry) => entry.isIntersecting);
				if (!visible.length) return;
				visible.sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
				const target = visible[0]?.target as HTMLElement | undefined;
				if (target?.id) activeAnchor = target.id;
			},
			{ rootMargin: '0px 0px -70% 0px', threshold: 0 }
		);
		headings.forEach((heading) => observer?.observe(heading));
	};

	const refreshObserver = async () => {
		await tick();
		setupObserver();
	};

	const scrollToAnchor = (anchor: string, event: MouseEvent) => {
		event.preventDefault();
		if (!contentRoot) return;
		const target = contentRoot.querySelector(`#${CSS.escape(anchor)}`) as HTMLElement | null;
		if (!target) return;
		target.scrollIntoView({ behavior: 'smooth', block: 'start' });
		activeAnchor = anchor;
		if (typeof history !== 'undefined') history.replaceState(null, '', `#${anchor}`);
	};

	$effect(() => {
		imageExtInfoCtx.syncModelData(
			imageExtInfoStore,
			buildImageExtInfoState($postStore?.extInfo ?? null)
		);
		void refreshObserver();
	});

	onDestroy(() => {
		observer?.disconnect();
	});

	const formatDate = (value?: string) => {
		if (!value) return '';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
	};
</script>

{#if $postStore}
	{#snippet backContent()}
		<ArrowLeft size={14} class="group-hover:-translate-x-1 transition-transform" />
		<span>返回</span>
	{/snippet}

	{#snippet shareContent()}
		<Share2 size={14} />
	{/snippet}

	{#snippet topContent()}
		返回顶部
	{/snippet}

	<StickyHeader title={$postStore.title} />

	<article class="article-enter space-y-10">
		<!-- Header -->
		<header class="max-w-4xl space-y-6">
			<div class="flex items-center gap-4">
				<Button
					variant="ghost"
					class="!h-auto !p-0 font-mono text-[10px] font-semibold tracking-[0.2em] text-ink-400 uppercase hover:!bg-transparent hover:text-ink-900 group"
					onclick={() => history.back()}
					content={backContent}
				/>
				<div class="h-px w-6 bg-ink-200/50 dark:bg-ink-800/50"></div>
			</div>

			<div class="space-y-4">
				<div class="flex items-center gap-3">
					<Badge variant="soft">专题</Badge>
					<span class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase"
						>技术与设计</span
					>
				</div>

				<h1
					class="font-serif text-2xl leading-[1.2] font-medium tracking-tight text-ink-950 md:text-3xl lg:text-4xl dark:text-ink-50"
				>
					{$postStore.title}
				</h1>

				<div
					class="flex flex-wrap items-center gap-5 font-mono text-[9px] tracking-widest text-ink-400 uppercase"
				>
					<span class="flex items-center gap-1.5">
						<Calendar size={12} />
						{formatDate($postStore.createdAt)}
					</span>
					<span class="flex items-center gap-1.5"><Clock size={12} /> 12 分钟阅读</span>
					<span class="flex items-center gap-1.5">
						浏览 {$postStore.metrics?.views ?? 0} · 喜欢 {$postStore.metrics?.likes ?? 0} · 评论
						{$postStore.metrics?.comments ?? 0}
					</span>
					{#snippet fallback()}
						<div class="">
							<Loading size="w-3 h-3" duration={1000} />
						</div>
					{/snippet}
					<QueryRoot
						loader={() => import('$lib/features/post/components/PostMetricsClient.svelte')}
						{fallback}
					/>
				</div>
			</div>

			{#if $postStore.leadIn}
				<p
					class="border-l-[1px] border-jade-500/20 py-0.5 pl-5 font-serif text-base leading-relaxed font-normal text-ink-600 italic opacity-90 md:text-lg dark:text-ink-400"
				>
					{$postStore.leadIn}
				</p>
			{/if}
		</header>

		<!-- Content Grid -->
		<div class="grid gap-10 lg:grid-cols-[1fr_220px] lg:gap-16">
			<!-- Main Content -->
			<main class="min-w-0">
				<div
					class="markdown-preview markdown-body prose prose-ink dark:prose-invert max-w-none text-[15px] leading-[1.8] font-normal text-ink-800 md:text-base dark:text-ink-200 prose-headings:mt-10 prose-headings:mb-4 prose-headings:font-serif prose-headings:font-medium prose-headings:tracking-tight prose-headings:text-ink-950 dark:prose-headings:text-ink-50 prose-h2:text-xl md:prose-h2:text-2xl prose-h3:text-lg md:prose-h3:text-xl prose-p:mb-6 prose-blockquote:my-8 prose-blockquote:border-l-[1px] prose-blockquote:border-jade-500/40 prose-blockquote:py-0.5 prose-blockquote:pl-5 prose-blockquote:text-[0.95em] prose-blockquote:text-ink-600 prose-blockquote:italic prose-blockquote:opacity-90 dark:prose-blockquote:text-ink-400"
					bind:this={contentRoot}
					use:markdownComponents
				>
					{@html contentHtml}
				</div>

				<footer
					class="mt-16 flex flex-col items-start justify-between gap-4 border-t border-ink-50 pt-8 md:flex-row md:items-center dark:border-ink-800/30"
				>
					<div class="flex items-center gap-3">
						<span class="font-mono text-[9px] tracking-widest text-ink-400 uppercase">分享此文</span
						>
						<Button
							variant="ghost"
							class="h-auto p-1.5 text-ink-400 hover:text-jade-600"
							title="分享"
							content={shareContent}
						/>
					</div>
					<Button
						variant="ghost"
						class="!h-auto !p-0 font-mono text-[9px] tracking-[0.2em] text-ink-400 uppercase hover:!bg-transparent hover:text-ink-900"
						onclick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
						content={topContent}
					/>
				</footer>

				<!-- Comments Section -->
				{#snippet commentFallback()}
					<div class="flex justify-center py-40">
						<Loading size="w-8 h-8" duration={1000} text="评论区在赶来的路上..." />
					</div>
				{/snippet}
				<QueryRoot
					loader={() => import('$lib/features/comment/components/CommentAreaClient.svelte')}
					loaderProps={{
						areaId: $postStore.commentAreaId,
						commentsCount: $postStore.metrics?.comments ?? 0
					}}
					fallback={commentFallback}
				/>
			</main>

			<!-- Sidebar / TOC -->
			{#if $postStore.toc?.length}
				<aside class="hidden slide-in-right lg:block">
					<div class="sticky top-20 space-y-10">
						<div class="space-y-5">
							<span
								class="block border-b border-ink-50 pb-2 font-mono text-[8px] font-bold tracking-[0.4em] text-ink-300 uppercase dark:border-ink-800/30"
							>
								本页目录
							</span>
							<ul class="space-y-3 font-sans">
								{#each $postStore.toc as item}
									<li class="space-y-2">
										<a
											class={`block text-[12px] text-ink-500 transition-all hover:translate-x-0.5 hover:text-jade-600 dark:text-ink-400 dark:hover:text-jade-400 ${
												activeAnchor === item.anchor
													? 'font-bold text-jade-700 dark:text-jade-400'
													: ''
											}`}
											href={'#' + item.anchor}
											onclick={(event) => scrollToAnchor(item.anchor, event)}
										>
											{item.name}
										</a>
										{#if item.children?.length}
											<ul class="space-y-1.5 border-l border-ink-50 pl-3 dark:border-ink-800/30">
												{#each item.children as child}
													<li>
														<a
															class={`block text-[11px] text-ink-400 transition-all hover:translate-x-0.5 hover:text-jade-500 dark:text-ink-500 ${
																activeAnchor === child.anchor
																	? 'font-bold text-jade-600 dark:text-jade-300'
																	: ''
															}`}
															href={'#' + child.anchor}
															onclick={(event) => scrollToAnchor(child.anchor, event)}
														>
															{child.name}
														</a>
													</li>
												{/each}
											</ul>
										{/if}
									</li>
								{/each}
							</ul>
						</div>

						<div class="space-y-2.5 rounded-lg border border-jade-500/10 bg-jade-500/5 p-5">
							<Tag variant="jade" class="border-none px-0">感悟</Tag>
							<p class="text-[10px] leading-relaxed font-normal text-ink-500 dark:text-ink-400">
								每一篇文章都是漫长探索中的一小步。如果这些文字能引起共鸣，欢迎留下你的思考。
							</p>
						</div>
					</div>
				</aside>
			{/if}
		</div>
	</article>
{:else}
	<div class="py-24 text-center font-serif text-sm text-ink-400 italic">
		<p>请求的内容未能呈现。</p>
	</div>
{/if}
