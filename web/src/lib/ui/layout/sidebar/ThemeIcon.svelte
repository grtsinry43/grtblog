<script lang="ts">
	import { resolveTheme, themeManager } from '$lib/shared/theme/theme.svelte';
	import { Moon, Sun } from 'lucide-svelte';

	const theme = themeManager;
	const resolved = $derived.by(() => resolveTheme(theme.current));

	type ViewTransitionLike = { ready: Promise<void> };
	type DocumentWithViewTransition = Document & {
		startViewTransition?: (callback: () => void) => ViewTransitionLike;
	};

	const isMobile = () => window.innerWidth < 768;

	const toggleTheme = async (event: MouseEvent) => {
		const next = resolved === 'dark' ? 'light' : 'dark';
		const doc = document as DocumentWithViewTransition;
		const root = document.documentElement;

		// On mobile or without View Transitions support: instant switch, no animation
		if (!doc.startViewTransition || isMobile()) {
			theme.set(next);
			return;
		}

		root.dataset.themeTransitioning = 'true';
		try {
			const x = event.clientX;
			const y = event.clientY;
			const endRadius = Math.hypot(
				Math.max(x, window.innerWidth - x),
				Math.max(y, window.innerHeight - y)
			);

			const transition = doc.startViewTransition.call(doc, () => {
				theme.set(next);
			});

			await transition.ready;

			const reveal = document.documentElement.animate(
				{
					clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`]
				},
				{
					duration: 350,
					easing: 'ease-out',
					pseudoElement: '::view-transition-new(root)'
				}
			);
			await reveal.finished;
		} finally {
			delete root.dataset.themeTransitioning;
		}
	};
</script>

<button
	type="button"
	data-theme={resolved}
	aria-label={`Switch to ${resolved === 'dark' ? 'light' : 'dark'} theme`}
	onclick={toggleTheme}
	class="h-10 w-10 rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100 flex items-center justify-center"
>
	{#if resolved === 'dark'}
		<Sun class="w-5 h-5 relative z-10" />
	{:else}
		<Moon class="w-5 h-5 relative z-10" />
	{/if}
</button>
