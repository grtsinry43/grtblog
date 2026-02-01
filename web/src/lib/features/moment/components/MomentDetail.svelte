<script lang="ts">
	import type { MomentDetail } from '$lib/features/moment/types';
	import { renderMarkdown } from '$lib/shared/markdown/markdown';
	import { markdownComponents } from '$lib/shared/actions/markdown-components';
	import { ArrowLeft, Sun } from 'lucide-svelte';
	import { page } from '$app/stores';
	import '$lib/ui/markdown/register';
	import Loading from '$lib/ui/common/Loading.svelte';
	import QueryRoot from '$lib/ui/common/QueryRoot.svelte';
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import { tick, onDestroy } from 'svelte';

	interface Props {
		moment: MomentDetail;
	}

	let { moment } = $props<Props>();

	const siteOrigin = $derived($page.url.origin);
	const contentHtml = $derived(
		renderMarkdown(moment.content ?? '', [], {
			origin: siteOrigin
		})
	);

	// Date formatting
	const dateObj = new Date(moment.createdAt);
	const dateStr = `${dateObj.getFullYear()}.${String(dateObj.getMonth() + 1).padStart(2, '0')}.${String(dateObj.getDate()).padStart(2, '0')}`;
	const dateNo = `${String(dateObj.getMonth() + 1).padStart(2, '0')}${String(dateObj.getDate()).padStart(2, '0')}`;

	// Season deriver (same as list)
	function getSeason(date: Date) {
		const month = date.getMonth() + 1;
		if (month >= 3 && month <= 5) return '春';
		if (month >= 6 && month <= 8) return '夏';
		if (month >= 9 && month <= 11) return '秋';
		return '冬';
	}
	const season = getSeason(dateObj);
	const column = moment.topics?.[0]?.name || '手记';

	// Mock Weather
	const weather = 'Sun';

	function goBack() {
		history.back();
	}

	// --- TOC Logic (Reused from PostDetail) ---
	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor = $state<string | null>(null);
	let observer: IntersectionObserver | null = null;

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
		// Re-run observer when content changes (or on mount logic if derived)
		void refreshObserver();
	});

	onDestroy(() => {
		observer?.disconnect();
	});
</script>

<div
	class="relative z-10 grid gap-10 lg:grid-cols-[1fr_220px] lg:gap-16 max-w-[1200px] mx-auto animate-sheet-enter origin-right pb-24"
>
	<!-- Main Paper Content -->
	<article class="flex-1 w-full relative">
		<!-- Visual Anchor (Bookmark) -->
		<div
			class="absolute -top-4 right-6 md:right-12 z-20 flex flex-col items-center animate-settle"
			style="animation-delay: 0.3s"
		>
			<div
				class="w-10 md:w-12 h-20 bg-ink-50 dark:bg-ink-800 shadow-lg rounded-b-sm border-x border-b border-ink-200 dark:border-ink-200/20 border-t-4 border-t-ink-800/10 flex flex-col items-center pt-3 pb-2 justify-between"
			>
				<div class="w-1.5 h-1.5 rounded-full bg-ink-300 dark:bg-ink-800/50 shadow-inner"></div>
				<span
					class="[writing-mode:vertical-rl] text-[11px] font-serif font-bold text-cinnabar-500 tracking-[0.3em] opacity-80"
				>
					{season}
				</span>
				<div class="w-full h-0.5 bg-cinnabar-500/20"></div>
			</div>
		</div>

		<!-- Navigation -->
		<div
			class="mb-6 px-4 md:px-0 opacity-60 hover:opacity-100 transition-opacity flex justify-between items-end"
		>
			<button
				onclick={goBack}
				class="group flex items-center gap-2 text-xs font-serif text-ink-800 dark:text-ink-200 hover:text-cinnabar-500 transition-colors"
			>
				<ArrowLeft size={14} class="transition-transform group-hover:-translate-x-1" />
				<span class="tracking-widest">收起这一页</span>
			</button>
		</div>

		<StickyHeader title={moment.title} />

		<!-- Paper Sheet -->
		<div
			class="
			bg-ink-50 md:bg-[#fbf9f4] dark:bg-ink-900 dark:md:bg-ink-900
			shadow-[0_4px_30px_-8px_rgba(0,0,0,0.06)] dark:shadow-none
			border border-ink-200/80 dark:border-ink-200/10
			px-8 py-12 md:p-20 rounded-sm relative overflow-hidden min-h-[80vh]
			transition-colors duration-500
		"
			style:view-transition-name={`moment-${moment.id}`}
		>
			<!-- Paper Texture & Watermark -->
			<div class="absolute inset-0 bg-noise opacity-30 pointer-events-none"></div>

			<!-- Content Container -->
			<div class="relative z-10">
				<!-- Header Area -->
				<header class="mb-12 flex flex-col gap-6">
					<div class="flex items-center justify-between border-b border-ink-800/10 pb-4">
						<div
							class="flex items-center gap-3 text-xs font-mono text-ink-800/40 dark:text-ink-200/40"
						>
							<span>NO. {dateNo}</span>
							<span>—</span>
							<span class="font-serif text-cinnabar-500">{column}</span>
						</div>
						<div class="text-ink-800/40 dark:text-ink-200/40">
							<Sun size={18} stroke-width={1.5} />
						</div>
					</div>

					<h1
						class="text-3xl md:text-5xl font-serif font-bold text-ink-900 dark:text-ink-50 leading-[1.2]"
					>
						{moment.title}
					</h1>
				</header>

				<!-- Body Text -->
				<div
					class="markdown-preview prose prose-stone dark:prose-invert prose-lg max-w-none text-ink-900/80 dark:text-ink-200/90 font-serif text-justify text-[15px]"
					bind:this={contentRoot}
					use:markdownComponents
				>
					{@html contentHtml}
				</div>

				<!-- Footer Stamp -->
				<div class="mt-24 flex justify-center opacity-40">
					<div
						class="w-24 h-24 border-2 border-dashed border-ink-800 dark:border-ink-200 rounded-full flex items-center justify-center rotate-12"
					>
						<div class="text-center text-ink-800 dark:text-ink-200">
							<div class="text-[9px] uppercase tracking-widest mb-1">审阅</div>
							<div class="font-serif font-bold text-lg">阅</div>
							<div class="text-[9px] mt-1">{dateStr}</div>
						</div>
					</div>
				</div>

				<!-- Comments Section -->
				{#snippet commentFallback()}
					<div class="flex justify-center py-20">
						<Loading size="w-6 h-6" duration={1000} text="Loading comments..." />
					</div>
				{/snippet}
				<div class="mt-16 pt-10 border-t border-ink-200/50 dark:border-ink-700/30">
					<QueryRoot
						loader={() => import('$lib/features/comment/components/CommentAreaClient.svelte')}
						loaderProps={{
							areaId: moment.commentAreaId,
							commentsCount: moment.metrics?.comments ?? 0
						}}
						fallback={commentFallback}
					/>
				</div>
			</div>
		</div>
	</article>

	<!-- TOC Sidebar (Desktop) -->
	{#if moment.toc?.length}
		<aside class="hidden lg:block pt-24 h-full relative slide-in-right">
			<div class="sticky top-24 space-y-10">
				<div class="space-y-5">
					<span
						class="block border-b border-ink-800/10 pb-2 font-mono text-[8px] font-bold tracking-[0.4em] text-ink-400 uppercase"
					>
						目录
					</span>
					<ul class="space-y-3 font-sans">
						{#each moment.toc as item}
							<li class="space-y-2">
								<a
									class={`block text-[12px] text-ink-500 transition-all hover:translate-x-0.5 hover:text-cinnabar-600 dark:text-ink-400 dark:hover:text-cinnabar-400 ${
										activeAnchor === item.anchor
											? 'font-bold text-cinnabar-700 dark:text-cinnabar-400'
											: ''
									}`}
									href={'#' + item.anchor}
									onclick={(event) => scrollToAnchor(item.anchor, event)}
								>
									{item.name}
								</a>
								{#if item.children?.length}
									<ul class="space-y-1.5 border-l border-ink-200 pl-3 dark:border-ink-800/30">
										{#each item.children as child}
											<li>
												<a
													class={`block text-[11px] text-ink-400 transition-all hover:translate-x-0.5 hover:text-cinnabar-500 dark:text-ink-500 ${
														activeAnchor === child.anchor
															? 'font-bold text-cinnabar-600 dark:text-cinnabar-300'
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
			</div>
		</aside>
	{/if}
</div>
