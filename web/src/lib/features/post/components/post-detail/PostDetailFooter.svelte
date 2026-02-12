<script lang="ts">
	import { Share2 } from 'lucide-svelte';
	import Button from '$lib/ui/primitives/button/Button.svelte';
</script>

{#snippet shareContent()}
	<Share2 size={14} />
{/snippet}

{#snippet topContent()}
	返回顶部
{/snippet}

<footer
	class="mt-16 flex flex-col items-start justify-between gap-4 border-t border-ink-50 pt-8 md:flex-row md:items-center dark:border-ink-800/30"
>
	<div class="flex items-center gap-3">
		<span class="font-mono text-[9px] tracking-widest text-ink-400 uppercase">分享此文</span>
		<Button
			variant="ghost"
			class="h-auto p-1.5 text-ink-400 hover:text-jade-600"
			title="分享"
			content={shareContent}
		/>
	</div>
	<Button
		variant="ghost"
		class="!h-auto !p-0 font-mono text-[9px] tracking-[0.2em] text-ink-400 uppercase hover:!bg-transparent hover:text-ink-900"
		onclick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
		content={topContent}
	/>
</footer>
