<script lang="ts">
	import { browser } from '$app/environment';
	import { intersect } from '$lib/shared/actions/intersect';
	import type { MobileTimelineEntry } from '../types';
	import MobileTimelineItem from './MobileTimelineItem.svelte';

	let { entry, delay = 0 } = $props<{
		entry: MobileTimelineEntry;
		delay?: number;
	}>();

	let revealed = $state(true);
	let reducedMotion = $state(false);

	const nodeAt = $derived(entry.side === 'left' ? '82%' : '18%');
	const cardTransform = $derived(
		revealed || reducedMotion
			? 'translate3d(0, 0, 0) rotate(0deg) scale(1)'
			: entry.side === 'left'
				? 'translate3d(-2.75rem, 0.9rem, 0) rotate(-2.2deg) scale(0.965)'
				: 'translate3d(2.75rem, 0.9rem, 0) rotate(2.2deg) scale(0.965)'
	);

	$effect(() => {
		if (!browser) return;

		reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		if (!reducedMotion) revealed = false;
	});

	function reveal() {
		revealed = true;
	}
</script>

<li
	class="timeline-entry relative pb-16 last:pb-8"
	class:is-revealed={revealed || reducedMotion}
	style:--reveal-delay={`${delay}ms`}
	use:intersect={{ onEnter: reveal, threshold: 0.16, rootMargin: '0px 0px -7% 0px' }}
>
	<div
		class="timeline-stem pointer-events-none absolute top-0 h-8 w-px bg-ink-300 dark:bg-ink-700"
		style:left={nodeAt}
		aria-hidden="true"
	></div>

	<svg
		class="pointer-events-none absolute inset-x-0 bottom-0 top-8 h-[calc(100%_-_2rem)] w-full overflow-visible"
		viewBox="0 0 100 100"
		preserveAspectRatio="none"
		aria-hidden="true"
	>
		<path
			d={entry.side === 'left' ? 'M 82 0 C 82 42, 18 58, 18 100' : 'M 18 0 C 18 42, 82 58, 82 100'}
			pathLength="1"
			fill="none"
			stroke="currentColor"
			stroke-width="0.7"
			vector-effect="non-scaling-stroke"
			class="timeline-curve text-ink-300 dark:text-ink-700"
		/>
	</svg>

	<div
		class="timeline-node pointer-events-none absolute top-8 z-20 h-2.5 w-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-ink-50 bg-jade-600 shadow-[0_0_0_3px_rgba(20,184,166,0.12)] dark:border-ink-950 dark:bg-jade-400"
		style:left={nodeAt}
		aria-hidden="true"
	></div>

	<div
		class="timeline-connector pointer-events-none absolute top-8 h-px bg-ink-300 dark:bg-ink-700 {entry.side ===
		'left'
			? 'left-[72%] w-[10%] origin-right'
			: 'left-[18%] w-[10%] origin-left'}"
		aria-hidden="true"
	></div>

	<div
		class="timeline-card relative z-10 w-[72%] pt-4 {entry.side === 'left' ? 'mr-auto' : 'ml-auto'}"
		style:transform={cardTransform}
	>
		<MobileTimelineItem item={entry.item} side={entry.side} />
	</div>
</li>

<style>
	.timeline-card {
		opacity: 0;
		filter: blur(3px);
		will-change: transform, opacity, filter;
		transition:
			opacity 520ms cubic-bezier(0.22, 1, 0.36, 1) calc(var(--reveal-delay) + 100ms),
			transform 720ms cubic-bezier(0.16, 1, 0.3, 1) calc(var(--reveal-delay) + 100ms),
			filter 600ms ease-out calc(var(--reveal-delay) + 100ms);
	}

	.timeline-node {
		opacity: 0;
		transform: translate(-50%, -50%) scale(0.25);
		transition:
			opacity 180ms ease-out var(--reveal-delay),
			transform 480ms cubic-bezier(0.34, 1.56, 0.64, 1) var(--reveal-delay);
	}

	.timeline-stem,
	.timeline-connector {
		opacity: 0;
		transform: scaleX(0);
		transition:
			opacity 180ms ease-out calc(var(--reveal-delay) + 45ms),
			transform 360ms cubic-bezier(0.22, 1, 0.36, 1) calc(var(--reveal-delay) + 45ms);
	}

	.timeline-stem {
		transform: scaleY(0);
		transform-origin: bottom;
	}

	.timeline-curve {
		stroke-dasharray: 1;
		stroke-dashoffset: 1;
		transition: stroke-dashoffset 900ms cubic-bezier(0.45, 0, 0.25, 1)
			calc(var(--reveal-delay) + 390ms);
	}

	.timeline-entry.is-revealed .timeline-card {
		opacity: 1;
		filter: blur(0);
	}

	.timeline-entry.is-revealed .timeline-node {
		opacity: 1;
		transform: translate(-50%, -50%) scale(1);
	}

	.timeline-entry.is-revealed .timeline-stem,
	.timeline-entry.is-revealed .timeline-connector {
		opacity: 1;
		transform: scale(1);
	}

	.timeline-entry.is-revealed .timeline-curve {
		stroke-dashoffset: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.timeline-card,
		.timeline-node,
		.timeline-stem,
		.timeline-connector,
		.timeline-curve {
			transition: none;
		}
	}
</style>
