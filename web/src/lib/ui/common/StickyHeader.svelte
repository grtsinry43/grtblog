<script lang="ts">
	import { ArrowLeft } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import Button from '$lib/ui/ui/button/Button.svelte';

	interface Props {
		title: string;
		showThreshold?: number;
	}

	let { title, showThreshold = 300 } = $props<Props>();

	let scrollY = $state(0);
	let showHeader = $derived(scrollY > showThreshold);

	// Simple smooth scroll to top if title is clicked
	const scrollToTop = () => {
		window.scrollTo({ top: 0, behavior: 'smooth' });
	};

	function goBack() {
		history.back();
	}
</script>

<svelte:window bind:scrollY />

{#if showHeader}
	<div
		class="fixed top-0 left-0 right-0 z-10 flex items-center justify-between px-4 py-3 bg-white/80 dark:bg-ink-950/80 backdrop-blur-md border-b border-ink-100 dark:border-ink-800/50 shadow-sm transition-all duration-300"
		in:fly={{ y: -20, duration: 300 }}
		out:fly={{ y: -20, duration: 200 }}
	>
		<div class="flex items-center gap-4 max-w-6xl mx-auto w-full">
			<Button
				variant="ghost"
				size="sm"
				class="!h-8 !w-8 text-ink-500 hover:text-ink-900 dark:text-ink-400 dark:hover:text-ink-100"
				onclick={goBack}
			>
				<ArrowLeft size={18} />
			</Button>

			<div class="h-4 w-px bg-ink-200 dark:bg-ink-800"></div>

			<button
				class="text-sm font-serif font-medium text-ink-900 dark:text-ink-100 truncate flex-1 text-left hover:opacity-80 transition-opacity"
				onclick={scrollToTop}
			>
				{title}
			</button>
		</div>
	</div>
{/if}
