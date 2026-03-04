<script lang="ts">
	import { resolve } from '$app/paths';
	import { SlideIn } from '$lib/ui/animation';
	import { Mail, Rss } from 'lucide-svelte';
	import SubscribeModal from './SubscribeModal.svelte';

	let isModalOpen = $state(false);
</script>

<section class="mt-16 border-t border-ink-100 dark:border-ink-800/50 pt-16">
	<SlideIn direction="up">
		<div class="flex flex-col md:flex-row items-center justify-between gap-8 px-4">
			<div class="flex flex-col gap-1 text-center md:text-left">
				<h2
					class="text-base font-serif font-medium text-ink-900 dark:text-ink-100 flex items-center justify-center md:justify-start gap-2"
				>
					订阅更新
				</h2>
				<p class="text-[11px] text-ink-400 font-mono">
					欢迎通过邮件或 RSS 订阅，第一时间获取最新文章和手记等内容的更新通知。
				</p>
			</div>

			<div class="flex items-center gap-3">
				<button
					onclick={() => (isModalOpen = true)}
					class="px-5 py-2.5 text-xs font-medium bg-jade-600 dark:bg-jade-500 text-white rounded-default hover:bg-jade-700 dark:hover:bg-jade-400 transition-all shadow-sm shadow-jade-500/10 flex items-center gap-2 group"
				>
					<Mail size={13} class="group-hover:rotate-12 transition-transform" />
					<span>邮件订阅</span>
				</button>

				<a
					href={resolve('/feed', {})}
					data-sveltekit-preload-data="off"
					class="px-5 py-2.5 text-xs font-medium bg-jade-50/50 dark:bg-jade-500/5 border border-jade-200/60 dark:border-jade-800/60 text-jade-700 dark:text-jade-300 rounded-default hover:border-jade-500/50 hover:text-jade-600 transition-all flex items-center gap-2 group"
				>
					<Rss size={13} class="group-hover:text-amber-500 transition-colors" />
					<span>RSS 订阅</span>
				</a>
			</div>
		</div>
	</SlideIn>
</section>

<SubscribeModal bind:isOpen={isModalOpen} />

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
