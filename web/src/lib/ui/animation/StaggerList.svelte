<script lang="ts">
	import { browser } from '$app/environment';
	import { Spring } from 'svelte/motion';
	import { intersect } from '$lib/shared/actions/intersect';
	import type { Snippet } from 'svelte';

	// Module-level set: survives SvelteKit client-side navigations,
	// so a list with the same key won't replay the entrance animation.
	const _animatedKeys = new Set<string>();

	let {
		children,
		staggerDelay = 100,
		y = 16,
		duration = 700,
		threshold = 0.1,
		rootMargin,
		spring: useSpring = true,
		class: className = '',
		itemSelector = ':scope > *',
		key = ''
	} = $props<{
		children: Snippet;
		staggerDelay?: number;
		y?: number;
		duration?: number;
		threshold?: number;
		rootMargin?: string;
		spring?: boolean;
		class?: string;
		itemSelector?: string;
		key?: string;
	}>();

	let wrapper: HTMLElement | undefined = $state();
	let revealed = $state(false);
	let reducedMotion = $state(false);
	let springs: Spring<number>[] = $state([]);
	let timers: ReturnType<typeof setTimeout>[] = [];

	// Check once at mount time — skip animation if this key was already played
	// svelte-ignore state_referenced_locally
	const skip = !!(key && _animatedKeys.has(key));

	$effect(() => {
		if (browser) {
			reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		}
	});

	const shouldAnimate = $derived(!skip && !reducedMotion);

	// Hide items on client mount; create Springs initialized at 0
	$effect(() => {
		if (!browser || !wrapper || !shouldAnimate) return;

		const items = wrapper.querySelectorAll<HTMLElement>(itemSelector);
		for (const item of items) {
			item.style.opacity = '0';
			item.style.transform = `translateY(${y}px)`;
			item.style.filter = 'blur(2px)';
		}

		if (useSpring) {
			springs = Array.from({ length: items.length }, () => {
				const sp = new Spring(1, { stiffness: 0.12, damping: 0.7 });
				sp.set(0, { instant: true });
				return sp;
			});
		}
	});

	// Spring-driven reveal
	$effect(() => {
		if (!revealed || !wrapper || !shouldAnimate || !useSpring) return;

		for (const t of timers) clearTimeout(t);
		timers = [];

		springs.forEach((sp, index) => {
			const t = setTimeout(() => {
				sp.target = 1;
			}, index * staggerDelay);
			timers.push(t);
		});

		if (key) _animatedKeys.add(key);

		return () => {
			for (const t of timers) clearTimeout(t);
		};
	});

	// Reactive style updates driven by Spring values
	$effect(() => {
		if (!wrapper || !useSpring || !shouldAnimate || springs.length === 0) return;

		const items = wrapper.querySelectorAll<HTMLElement>(itemSelector);
		springs.forEach((sp, index) => {
			const item = items[index];
			if (!item) return;
			const p = sp.current;
			item.style.opacity = String(Math.max(0, Math.min(1, p)));
			item.style.transform = `translateY(${y * (1 - p)}px)`;
			item.style.filter = `blur(${Math.max(0, 2 * (1 - p))}px)`;
		});
	});

	// CSS transition fallback reveal (when spring=false)
	$effect(() => {
		if (!revealed || !wrapper || !shouldAnimate || useSpring) return;

		const items = wrapper.querySelectorAll<HTMLElement>(itemSelector);
		items.forEach((item, index) => {
			const itemDelay = index * staggerDelay;
			item.style.transition = `opacity ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${itemDelay}ms, transform ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${itemDelay}ms, filter ${duration}ms cubic-bezier(0.23, 1, 0.32, 1) ${itemDelay}ms`;
			item.style.opacity = '1';
			item.style.transform = 'translateY(0)';
			item.style.filter = 'blur(0)';
		});

		if (key) _animatedKeys.add(key);
	});

	function onEnter() {
		if (!skip) revealed = true;
	}
</script>

<div bind:this={wrapper} class={className} use:intersect={{ onEnter, threshold, rootMargin }}>
	{@render children()}
</div>
