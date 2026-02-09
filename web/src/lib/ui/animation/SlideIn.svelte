<script lang="ts">
	import { browser } from '$app/environment';
	import { Spring } from 'svelte/motion';
	import { intersect } from '$lib/shared/actions/intersect';
	import type { Snippet } from 'svelte';

	let {
		children,
		direction = 'left',
		offset = 30,
		duration = 800,
		delay = 0,
		threshold = 0.1,
		rootMargin,
		spring: useSpring = true,
		class: className = ''
	} = $props<{
		children: Snippet;
		direction?: 'left' | 'right' | 'up' | 'down';
		offset?: number;
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
	const progress = new Spring(1, { stiffness: 0.12, damping: 0.7 });

	// Direction multipliers
	const dirX: Record<string, number> = { left: -1, right: 1, up: 0, down: 0 };
	const dirY: Record<string, number> = { left: 0, right: 0, up: -1, down: 1 };

	// CSS fallback direction map
	const directionMap: Record<string, string> = {
		left: `translateX(-${offset}px)`,
		right: `translateX(${offset}px)`,
		up: `translateY(-${offset}px)`,
		down: `translateY(${offset}px)`
	};

	$effect(() => {
		if (browser) {
			reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
			if (!reducedMotion) {
				if (useSpring) {
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

	const isSpringDriven = $derived(browser && useSpring && !reducedMotion);

	const currentOpacity = $derived(
		isSpringDriven ? Math.max(0, Math.min(1, progress.current)) : reducedMotion || visible ? 1 : 0
	);

	const currentTransform = $derived(
		isSpringDriven
			? `translate(${offset * dirX[direction] * (1 - progress.current)}px, ${offset * dirY[direction] * (1 - progress.current)}px)`
			: reducedMotion || visible
				? 'translate(0, 0)'
				: directionMap[direction]
	);
</script>

<div
	class={className}
	use:intersect={{ onEnter, threshold, rootMargin }}
	style:opacity={currentOpacity}
	style:transform={currentTransform}
	style:transition={!useSpring && !reducedMotion
		? `opacity ${duration}ms cubic-bezier(0.16, 1, 0.3, 1) ${delay}ms, transform ${duration}ms cubic-bezier(0.16, 1, 0.3, 1) ${delay}ms`
		: 'none'}
>
	{@render children()}
</div>
