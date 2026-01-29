<script lang="ts">
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';
	import { resolveTheme, themeManager } from '$lib/shared/theme/theme.svelte';

	const theme = themeManager;
	const resolved = $derived.by(() => resolveTheme(theme.current));
	const iconName = $derived.by(() => (resolved === 'dark' ? 'Moon' : 'Sun'));

	type ViewTransitionLike = { ready: Promise<void> };
	type DocumentWithViewTransition = Document & {
		startViewTransition?: (callback: () => void) => ViewTransitionLike;
	};

	const toggleTheme = async (event: MouseEvent) => {
		const next = resolved === 'dark' ? 'light' : 'dark';
		const doc = document as DocumentWithViewTransition;
		if (!doc.startViewTransition) {
			theme.set(next);
			return;
		}

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

		document.documentElement.animate(
			{
				clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`]
			},
			{
				duration: 500,
				easing: 'ease-in-out',
				pseudoElement: '::view-transition-new(root)'
			}
		);
	};
</script>

<button
	type="button"
	data-theme={resolved}
	aria-label={`Switch to ${resolved === 'dark' ? 'light' : 'dark'} theme`}
	onclick={toggleTheme}
	class="rounded-default hover:bg-ink-200 dark:hover:bg-ink-800 p-2"
>
	<DynamicLucideIcon name={iconName} className="theme-icon w-6 h-6" />
</button>

<style lang="postcss">
	@reference "$routes/layout.css";

	.theme-icon {
		@apply relative z-10 ;
	}
</style>
