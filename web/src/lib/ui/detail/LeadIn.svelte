<script lang="ts">
	import { Quote } from 'lucide-svelte';

	interface Props {
		content: string;
		label?: string;
	}

	let { content, label = '引入' }: Props = $props();
</script>

{#if content}
	<div class="group relative mb-10 overflow-hidden">
		<!-- 背景装饰 -->
		<div
			class="absolute -top-10 -left-10 h-32 w-32 rounded-full bg-jade-500/5 blur-3xl transition-colors group-hover:bg-jade-500/10"
		></div>
		<div
			class="absolute -bottom-10 -right-10 h-32 w-32 rounded-full bg-amber-500/5 blur-3xl transition-colors group-hover:bg-amber-500/10"
		></div>

		<div class="relative flex gap-6 px-1">
			<!-- 左侧装饰条 -->
			<div class="flex flex-col items-center gap-2 py-1">
				<div
					class="flex h-8 w-8 items-center justify-center rounded-full bg-ink-50 text-ink-300 transition-colors group-hover:bg-jade-50 group-hover:text-jade-500 dark:bg-ink-900/50 dark:text-ink-700 dark:group-hover:bg-jade-900/30 dark:group-hover:text-jade-400"
				>
					<Quote size={14} />
				</div>
				<div
					class="w-px flex-1 bg-gradient-to-b from-ink-100 via-ink-100 to-transparent dark:from-ink-800 dark:via-ink-800"
				></div>
			</div>

			<!-- 内容区 -->
			<div class="flex-1 space-y-3 py-1">
				<div class="flex items-center gap-2">
					<span
						class="font-mono text-[10px] font-bold tracking-[0.2em] text-ink-400 uppercase transition-colors group-hover:text-jade-600 dark:text-ink-500 dark:group-hover:text-jade-400"
						>{label}</span
					>
					<div
						class="h-px w-8 bg-ink-100 transition-all group-hover:w-12 group-hover:bg-jade-200 dark:bg-ink-800 dark:group-hover:bg-jade-800"
					></div>
				</div>

				<p
					class="font-serif text-[14px] leading-relaxed text-ink-700 italic opacity-90 transition-opacity group-hover:opacity-100 md:text-[15px] dark:text-ink-300"
				>
					{content}
				</p>
			</div>
		</div>
	</div>
{/if}
