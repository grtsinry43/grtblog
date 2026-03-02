<script lang="ts">
	import { browser } from '$app/environment';
	import { intersect } from '$lib/shared/actions/intersect';
	import type { Snippet } from 'svelte';

	let {
		children,
		staggerDelay = 80,
		y = 16,
		duration = 500,
		threshold = 0.1,
		rootMargin,
		class: className = '',
		itemSelector = ':scope > *'
	} = $props<{
		children: Snippet;
		staggerDelay?: number;
		y?: number;
		duration?: number;
		threshold?: number;
		rootMargin?: string;
		class?: string;
		itemSelector?: string;
	}>();

	let wrapper: HTMLElement | undefined = $state();
	let revealed = $state(false);
	let reducedMotion = $state(false);

	$effect(() => {
		if (browser) {
			reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		}
	});

	// Set initial hidden state on items
	$effect(() => {
		if (!browser || !wrapper || reducedMotion) return;

		const items = wrapper.querySelectorAll<HTMLElement>(itemSelector);
		for (const item of items) {
			item.style.opacity = '0';
		}
	});

	// On reveal, apply staggered CSS animations via a class + custom property
	$effect(() => {
		if (!revealed || !wrapper || reducedMotion) return;

		const items = wrapper.querySelectorAll<HTMLElement>(itemSelector);
		items.forEach((item, index) => {
			item.style.setProperty('--stagger-index', String(index));
			item.classList.add('stagger-reveal');
		});
	});

	function onEnter() {
		revealed = true;
	}
</script>

<div
	bind:this={wrapper}
	class={className}
	use:intersect={{ onEnter, threshold, rootMargin }}
	style:--stagger-delay="{staggerDelay}ms"
	style:--stagger-duration="{duration}ms"
	style:--stagger-y="{y}px"
>
	{@render children()}
</div>

<style>
	div :global(.stagger-reveal) {
		animation: stagger-enter var(--stagger-duration, 500ms) cubic-bezier(0.23, 1, 0.32, 1) both;
		animation-delay: calc(var(--stagger-index, 0) * var(--stagger-delay, 80ms));
	}

	@keyframes stagger-enter {
		from {
			opacity: 0;
			transform: translateY(var(--stagger-y, 16px));
			filter: blur(2px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
			filter: blur(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		div :global(.stagger-reveal) {
			animation: none;
			opacity: 1;
		}
	}
</style>
