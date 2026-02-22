<script lang="ts">
	import FriendLinkGrid from '$lib/features/friend-link/components/FriendLinkGrid.svelte';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { FadeIn, SlideIn } from '$lib/ui/animation';
	import { Heart, Plus } from 'lucide-svelte';

	import { userStore } from '$lib/shared/stores/userStore';
	import { authModalStore } from '$lib/shared/stores/authModalStore';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';

	let { data } = $props();

	function handleApplyClick() {
		if ($userStore.isLogin) {
			windowStore.open('申请友链');
		} else {
			authModalStore.open('apply-friend-link');
		}
	}
</script>

<div class="friends-page max-w-5xl mx-auto space-y-16">
	<PageHeader 
		title="友情链接" 
		tag="Friends" 
		subtitle="青山一道同云雨，明月何曾是两乡" 
		description="志同道合者的数字家园，感谢在这个广袤网络中的相遇。"
	/>

	<!-- Friends Grid - 直接渲染，确保服务端渲染 SEO -->
	<div class="friends-content">
		<FriendLinkGrid links={data.links} />
	</div>

	<!-- Bottom Section -->
	<FadeIn y={10} duration={1000} delay={600}>
		<div class="border-t border-ink-100 dark:border-ink-800 pt-10 flex flex-col items-center">
			<div
				class="p-8 rounded-default bg-ink-50/50 dark:bg-ink-950/30 border border-dashed border-ink-200 dark:border-ink-800 max-w-2xl w-full"
			>
				<h2
					class="text-sm font-bold text-ink-900 dark:text-ink-100 uppercase tracking-widest mb-6 font-mono text-center"
				>
					友链说明
				</h2>
				<ul class="text-xs text-ink-500 dark:text-ink-400 space-y-3 font-serif">
					<li class="flex gap-2">
						<span>•</span> 优先考虑经常更新、内容优质的技术博客或生活记录。
					</li>
					<li class="flex gap-2">
						<span>•</span> 站点需支持 HTTPS，且排版整洁，无大量广告或误导性内容。
					</li>
					<li class="flex gap-2">
						<span>•</span> 申请前请先在贵站添加本站链接，这将视作一种友好的相互确认。
					</li>
					<li class="flex gap-2">
						<span>•</span> 本站名称：<span
							class="text-jade-600 dark:text-jade-400 font-sans font-bold">grtBlog</span
						>
					</li>
					<li class="flex gap-2">
						<span>•</span> 本站地址：<span
							class="text-jade-600 dark:text-jade-400 font-sans font-bold"
							>https://grtsinry43.com</span
						>
					</li>
				</ul>

				<div class="mt-10 flex justify-center">
					<button
						onclick={handleApplyClick}
						class="flex items-center gap-2 px-4 py-2 bg-ink-900 dark:bg-ink-100 text-ink-0 dark:text-ink-950 rounded-default hover:bg-jade-600 dark:hover:bg-jade-400 transition-all duration-300 font-bold text-[11px] shadow-sm group"
					>
						<Plus size={14} class="group-hover:rotate-90 transition-transform duration-300" />
						申请加入
					</button>
				</div>
			</div>
		</div>
	</FadeIn>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
