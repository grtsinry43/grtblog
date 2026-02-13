<script lang="ts">
	import { resolve } from '$app/paths';
	import Hero from '$lib/features/home/Hero.svelte';
	import InspirationGrid from '$lib/features/home/InspirationGrid.svelte';
	import SubscribeSection from '$lib/features/home/SubscribeSection.svelte';
	import ActivityPulse from '$lib/features/home/ActivityPulse.svelte';
	import HomeArticleItem from '$lib/features/post/components/HomeArticleItem.svelte';
	import HomeMomentItem from '$lib/features/moment/components/HomeMomentItem.svelte';
	import { SlideIn, StaggerList } from '$lib/ui/animation';
	import { ArrowRight } from 'lucide-svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();
</script>

<div class="homepage-container">
	<Hero />

	<div class="max-w-300 mx-auto px-6 py-12 md:py-20">
		<div class="grid grid-cols-1 md:grid-cols-2 gap-12 lg:gap-24">
			<!-- Recent Articles -->
			<section>
				<SlideIn direction="left">
					<div
						class="flex items-center justify-between mb-6 border-b border-ink-100 dark:border-ink-800 pb-4"
					>
						<div class="flex items-center gap-3">
							<span class="h-px w-8 bg-jade-500/40"></span>
							<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">
								最近文章
							</h2>
						</div>
						<a
							href={resolve('/posts')}
							class="flex items-center gap-1 text-xs font-mono text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors group"
						>
							<span>查看全部</span>
							<ArrowRight size={12} class="group-hover:translate-x-1 transition-transform" />
						</a>
					</div>
				</SlideIn>

				<StaggerList staggerDelay={100} y={16} class="flex flex-col">
					{#each data.recentPosts.items as post (post.id)}
						<HomeArticleItem {post} />
					{/each}
				</StaggerList>
			</section>

			<!-- Recent Moments -->
			<section>
				<SlideIn direction="right">
					<div
						class="flex items-center justify-between mb-6 border-b border-ink-100 dark:border-ink-800 pb-4"
					>
						<div class="flex items-center gap-3">
							<span class="h-px w-8 bg-jade-500/40"></span>
							<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">
								最近手记
							</h2>
						</div>
						<a
							href={resolve('/moments')}
							class="flex items-center gap-1 text-xs font-mono text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors group"
						>
							<span>查看全部</span>
							<ArrowRight size={12} class="group-hover:translate-x-1 transition-transform" />
						</a>
					</div>
				</SlideIn>

				<StaggerList staggerDelay={100} y={16} class="flex flex-col">
					{#each data.recentMoments.items as moment (moment.id)}
						<HomeMomentItem {moment} />
					{/each}
				</StaggerList>
			</section>
		</div>

		<!-- New Inspiration Grid -->
		<InspirationGrid />

		<!-- New Activity Pulse -->
		<ActivityPulse />

		<!-- New Subscribe Section -->
		<SubscribeSection />
	</div>
</div>

<style lang="postcss">
	@reference "./layout.css";
</style>
