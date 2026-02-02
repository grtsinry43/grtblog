<script lang="ts">
	import { slide } from 'svelte/transition';
	import { X } from 'lucide-svelte';

	let { isOpen, onClose } = $props<{ isOpen: boolean; onClose: () => void }>();
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-[60] flex justify-end bg-black/20 backdrop-blur-sm lg:hidden"
		onclick={onClose}
		role="presentation"
	>
		<div
			class="h-full w-3/4 max-w-sm bg-white p-4 shadow-xl transition-transform dark:bg-ink-900"
			transition:slide={{ axis: 'x', duration: 300 }}
			onclick={(e) => e.stopPropagation()}
			role="presentation"
		>
			<div class="mb-4 flex items-center justify-between">
				<h2 class="text-lg font-bold">目录</h2>
				<button onclick={onClose} class="p-2">
					<X size={24} />
				</button>
			</div>
			<div class="text-ink-500">
				<p>Table of Contents placeholder.</p>
			</div>
		</div>
	</div>
{/if}
