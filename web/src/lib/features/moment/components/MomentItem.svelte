<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import type { MomentSummary } from '$lib/features/moment/types';
	import { ArrowRight, Pin } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { isDifferentDay } from '$lib/shared/utils/date';

	interface Props {
		moment: MomentSummary;
	}

	let { moment }: Props = $props();

	// Helpers to format date and derivation
	const dateObj = $derived.by(() => new Date(moment.createdAt));
	const formattedDate = $derived.by(
		() =>
			`${String(dateObj.getMonth() + 1).padStart(2, '0')}.${String(dateObj.getDate()).padStart(2, '0')}`
	);
	const contentUpdatedDateObj = $derived.by(() => new Date(moment.contentUpdatedAt));
	const formattedUpdatedDate = $derived.by(
		() =>
			`${String(contentUpdatedDateObj.getMonth() + 1).padStart(2, '0')}.${String(contentUpdatedDateObj.getDate()).padStart(2, '0')}`
	);
	const showUpdated = $derived(isDifferentDay(moment.createdAt, moment.contentUpdatedAt));
	const columnLabel = $derived.by(() => {
		const name = (moment.columnName || '').trim();
		return name || '未分类手记';
	});
	const hasImages = $derived(!!(moment.image && moment.image.length > 0));

	// Navigate to detail
	const handleClick = () => {
		goto(resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt)));
	};

	// Image blur-to-sharp load effect
	const handleImgLoad = (e: Event) => {
		const img = e.target as HTMLImageElement;
		img.classList.add('moment-img-loaded');
	};
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="group cursor-pointer relative moment-vt"
	style="view-transition-name: moment-{moment.id};"
	onclick={handleClick}
>
	<!-- Card Body -->
	<div
		class="
		relative w-full aspect-[3/4]
		bg-ink-50 dark:bg-ink-950/40
		border border-ink-200 dark:border-ink-200/10
		hover:border-jade-500/30 dark:hover:border-jade-500/40
		shadow-[0_2px_15px_-3px_rgba(0,0,0,0.05)]
		hover:shadow-[0_20px_40px_-15px_rgba(0,0,0,0.1)]
		hover:-translate-y-1.5
		transition-all duration-500 ease-[cubic-bezier(0.22,1,0.36,1)]
		flex flex-col overflow-hidden rounded-sm
		noise-surface
	"
	>
		<!-- Image strip -->
		{#if moment.image && moment.image.length > 0}
			<div class="moment-img-zone relative shrink-0">
				<div
					class="moment-img-strip flex overflow-x-auto overflow-y-hidden"
					onclick={(e) => e.stopPropagation()}
				>
					{#each moment.image as src, i (`${moment.id}-${i}-${src}`)}
						<img
							{src}
							alt="{moment.title} - {i + 1}"
							class="moment-img h-36 w-auto object-cover shrink-0"
							loading="lazy"
							draggable={false}
							onload={handleImgLoad}
						/>
					{/each}
				</div>
				<!-- Progressive blur transition band -->
				<div class="moment-blur-band" aria-hidden="true">
					<div class="moment-blur-layer moment-blur-1"></div>
					<div class="moment-blur-layer moment-blur-2"></div>
					<div class="moment-blur-layer moment-blur-3"></div>
					<div class="moment-blur-layer moment-blur-4"></div>
				</div>
			</div>
		{/if}

		<!-- Vertical Column Label — always visible, floats over images -->
		<div
			class="absolute top-0 right-6 h-16 w-8 bg-jade-500/5 dark:bg-jade-500/10 border-x border-jade-500/10 flex items-center justify-center pt-2 z-10"
		>
			<span
				class="[writing-mode:vertical-rl] text-[9px] font-serif font-bold text-jade-700 dark:text-jade-400 tracking-[0.2em] opacity-80 uppercase"
			>
				{columnLabel}
			</span>
		</div>

		<div class="relative z-[2] flex flex-col flex-1 min-h-0 p-6 {hasImages ? '-mt-3' : ''}">
			<!-- Top Meta -->
			<div class="flex items-start justify-between mb-4 shrink-0">
				<div class="flex flex-col gap-1">
					<div class="flex items-center gap-2">
						<span class="font-mono text-[10px] text-ink-400 dark:text-ink-500 tracking-wider">
							{formattedDate}{#if showUpdated}<span class="text-ink-300 dark:text-ink-600 ml-1"
									>（更新于 {formattedUpdatedDate}）</span
								>{/if}
						</span>
						{#if moment.isTop}
							<span
								class="inline-flex shrink-0 items-center gap-0.5 align-middle px-1 py-px text-[9px] font-mono font-normal tracking-wider text-jade-600 dark:text-jade-400"
							>
								<Pin size={9} strokeWidth={2} class="rotate-45" />
							</span>
						{/if}
					</div>
					<div class="h-px w-8 bg-jade-500/30"></div>
				</div>
			</div>

			<!-- Title & Content Preview -->
			<div class="flex-1 min-h-0 overflow-hidden flex flex-col gap-4 mt-2">
				<h3
					class="font-serif font-bold text-lg text-ink-900 dark:text-ink-100 leading-relaxed group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-300 line-clamp-2"
				>
					{moment.title}
				</h3>

				{#if moment.summary}
					<p
						class="text-[13px] text-ink-600 dark:text-ink-400 font-serif leading-loose {hasImages
							? 'line-clamp-2'
							: 'line-clamp-4'} opacity-80"
					>
						{moment.summary}
					</p>
				{/if}
			</div>

			<!-- Bottom Actions/Decor -->
			<div
				class="mt-auto pt-4 shrink-0 border-t border-ink-100 dark:border-ink-800/50 flex items-center justify-between"
			>
				<div class="flex items-center gap-3 text-[10px] font-mono text-ink-400">
					<span class="flex items-center gap-1">
						浏览 {moment.views}
					</span>
					<span class="opacity-30">/</span>
					<span class="flex items-center gap-1">
						评论 {moment.comments}
					</span>
				</div>

				<div
					class="opacity-0 group-hover:opacity-100 transition-all duration-500 transform translate-x-2 group-hover:translate-x-0"
				>
					<ArrowRight size={14} class="text-jade-600" />
				</div>
			</div>
		</div>
	</div>

	<!-- Aesthetic Shadow -->
	<div
		class="absolute -bottom-2 left-1/2 -translate-x-1/2 w-[85%] h-4 bg-black/5 blur-md rounded-full -z-10 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
	></div>
</div>

<style>
	/* Hide scrollbar */
	.moment-img-strip::-webkit-scrollbar {
		display: none;
	}
	.moment-img-strip {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}

	/*
	 * Progressive blur: 4 stacked layers, each with increasing
	 * backdrop-filter blur and a mask that only reveals its slice,
	 * creating a smooth sharp → blurred → solid color transition.
	 */
	.moment-blur-band {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 40%;
		pointer-events: none;
		z-index: 1;
	}

	.moment-blur-layer {
		position: absolute;
		inset: 0;
	}

	/* Layer 1: very slight blur at the top of the band */
	.moment-blur-1 {
		backdrop-filter: blur(2px);
		-webkit-backdrop-filter: blur(2px);
		-webkit-mask-image: linear-gradient(
			to bottom,
			transparent 0%,
			black 25%,
			black 25.1%,
			transparent 50%
		);
		mask-image: linear-gradient(to bottom, transparent 0%, black 25%, black 25.1%, transparent 50%);
	}

	/* Layer 2: medium blur in the middle */
	.moment-blur-2 {
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
		-webkit-mask-image: linear-gradient(
			to bottom,
			transparent 20%,
			black 50%,
			black 50.1%,
			transparent 75%
		);
		mask-image: linear-gradient(
			to bottom,
			transparent 20%,
			black 50%,
			black 50.1%,
			transparent 75%
		);
	}

	/* Layer 3: heavy blur toward the bottom */
	.moment-blur-3 {
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		-webkit-mask-image: linear-gradient(to bottom, transparent 45%, black 75%, black 100%);
		mask-image: linear-gradient(to bottom, transparent 45%, black 75%, black 100%);
	}

	/* Layer 4: color fill that dissolves into card bg */
	.moment-blur-4 {
		backdrop-filter: blur(20px) saturate(1.2);
		-webkit-backdrop-filter: blur(20px) saturate(1.2);
		background: linear-gradient(
			to bottom,
			transparent 30%,
			rgba(250, 250, 249, 0.4) 60%,
			rgba(250, 250, 249, 0.85) 100%
		);
		-webkit-mask-image: linear-gradient(to bottom, transparent 30%, black 100%);
		mask-image: linear-gradient(to bottom, transparent 30%, black 100%);
	}

	:global(.dark) .moment-blur-4 {
		background: linear-gradient(
			to bottom,
			transparent 30%,
			rgba(5, 5, 5, 0.4) 60%,
			rgba(5, 5, 5, 0.85) 100%
		);
	}

	/* Image: start blurry + slightly scaled, animate to sharp on load */
	.moment-img {
		filter: blur(12px);
		transform: scale(1.05);
		transition:
			filter 0.7s cubic-bezier(0.4, 0, 0.2, 1),
			transform 0.7s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.moment-img:global(.moment-img-loaded) {
		filter: blur(0);
		transform: scale(1);
	}
</style>
