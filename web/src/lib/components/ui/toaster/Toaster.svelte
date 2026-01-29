<script lang="ts">
	import { Toaster as SonnerToaster } from 'svelte-sonner';
	import { resolveTheme, themeManager } from '$lib/shared/theme.svelte';

	const theme = themeManager;
	const resolvedTheme = $derived(resolveTheme(theme.current));
</script>

<SonnerToaster
	position="top-right"
	theme={resolvedTheme}
	toastOptions={{
		unstyled: true,
		classes: {
			toast:
				'noise-surface flex items-center gap-3 rounded-default border border-ink-200/70 bg-ink-0/80 px-3.5 py-2.5 text-ink-900 shadow-subtle backdrop-blur-lg backdrop-saturate-150 dark:border-ink-800/60 dark:bg-ink-900/70 dark:text-ink-100 dark:shadow-deep',
			title: 'text-[12px] font-medium tracking-tight',
			description: 'text-[11px] text-ink-600/90 dark:text-ink-400/90',
			actionButton:
				'rounded-default bg-jade-600 px-2.5 py-1 text-[10px] font-semibold text-white hover:bg-jade-500',
			cancelButton:
				'rounded-default bg-ink-200 px-2.5 py-1 text-[10px] font-semibold text-ink-700 hover:bg-ink-300 dark:bg-ink-800 dark:text-ink-200 dark:hover:bg-ink-700',
			closeButton:
				'text-ink-500 hover:text-ink-700 dark:text-ink-400 dark:hover:text-ink-200'
		}
	}}
/>
