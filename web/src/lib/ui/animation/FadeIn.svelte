<script lang="ts">
	import { browser } from '$app/environment';
	import { Spring } from 'svelte/motion';
	import { intersect } from '$lib/shared/actions/intersect';
	import type { Snippet } from 'svelte';

	let {
		children,
		y = 20,
		duration = 800,
		delay = 0,
		threshold = 0.1,
		rootMargin,
		spring: useSpring = true,
		class: className = ''
	} = $props<{
		children: Snippet;
		y?: number;
		duration?: number;
		delay?: number;
		threshold?: number;
		rootMargin?: string;
		spring?: boolean;
		class?: string;
	}>();

	// visible=true for SSR; CSS fallback path toggles this
	let visible = $state(true);
	let reducedMotion = $state(false);

	// Initialize at 1 so SSR renders fully visible content.
	// On client mount we snap to 0; on intersect we animate to 1.
	const progress = new Spring(1, { stiffness: 0.12, damping: 0.7 });

	$effect(() => {
		if (browser) {
			reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
			if (!reducedMotion) {
				if (useSpring) {
					// Snap to hidden instantly (no animation) — intersect will reverse this
					progress.set(0, { instant: true });
				} else {
					visible = false;
				}
			}
		}
	});

	function onEnter() {
		if (useSpring && !reducedMotion) {
			setTimeout(() => {
				progress.target = 1;
			}, delay);
		} else {
			visible = true;
		}
	}

	// Spring path: driven by progress (1 during SSR, 0 after mount, animates to 1)
	// CSS path: driven by visible flag
	const isSpringDriven = $derived(browser && useSpring && !reducedMotion);

	const currentOpacity = $derived(
		isSpringDriven ? Math.max(0, Math.min(1, progress.current)) : reducedMotion || visible ? 1 : 0
	);

	const currentTranslateY = $derived(
		isSpringDriven ? y * (1 - progress.current) : reducedMotion || visible ? 0 : y
	);

	const currentBlur = $derived(
		isSpringDriven
			? Math.max(0, 3 * (1 - progress.current))
			: reducedMotion || visible
				? 0
				: 3
	);
</script>

<div
	class={className}
	use:intersect={{ onEnter, threshold, rootMargin }}
	style:opacity={currentOpacity}
	style:transform="translateY({currentTranslateY}px)"
	style:filter="blur({currentBlur}px)"
	style:transition={!useSpring && !reducedMotion
		? `opacity ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${delay}ms, transform ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${delay}ms, filter ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${delay}ms`
		: 'none'}
>
	{@render children()}
</div>
