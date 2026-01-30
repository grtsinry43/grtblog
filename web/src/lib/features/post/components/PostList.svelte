<script lang="ts">
	import type { PostSummary } from '$lib/features/post/types';
	import Card from '$lib/ui/ui/card/Card.svelte';
	import Pagination from '$lib/ui/ui/pagination/Pagination.svelte';
	import Tag from '$lib/ui/ui/tag/Tag.svelte';
	import { ArrowRight, FileText, Calendar, Sparkles } from 'lucide-svelte';
	import { postContext } from '$routes/posts/post-context';
	import { goto } from '$app/navigation';

	type PaginationData = {
		total: number;
		page: number;
		size: number;
	};

	const postsStore = postContext.selectModelData((state) => state?.posts ?? []);
	const totalStore = postContext.selectModelData((state) => state?.pagination?.total ?? 0);
	const pageStore = postContext.selectModelData((state) => state?.pagination?.page ?? 1);
	const sizeStore = postContext.selectModelData((state) => state?.pagination?.size ?? 10);

	let posts = postsStore;
	let total = totalStore;
	let page = pageStore;
	let size = sizeStore;

	const pagination: PaginationData = $derived({
		total: $total,
		page: $page,
		size: $size
	});

	let totalPages = $derived(
		pagination.size > 0 ? Math.max(1, Math.ceil(pagination.total / pagination.size)) : 1
	);

	const onPageChange = (page: number) => {
		const safePage = Number.isFinite(page) && page > 1 ? page : 1;
		goto(safePage === 1 ? '/posts/' : `/posts/page/${safePage}/`);
	};

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		return `${date.getFullYear()} . ${String(date.getMonth() + 1).padStart(2, '0')} . ${String(date.getDate()).padStart(2, '0')}`;
	};
</script>

<div class="w-full max-w-5xl mx-auto px-4 sm:px-6 py-10 space-y-12">
	<!-- Header Section -->
	<header
		class="space-y-4 text-center sm:text-left border-b border-ink-100 dark:border-ink-800 pb-8"
	>
		<div class="flex items-center justify-center sm:justify-start gap-3">
			<h1
				class="font-serif text-3xl sm:text-4xl font-medium tracking-tight text-ink-950 dark:text-ink-50"
			>
				文章归档
			</h1>
			<span class="hidden sm:inline-block h-px w-12 bg-jade-500/50"></span>
		</div>
		<p
			class="max-w-2xl text-sm sm:text-base leading-relaxed text-ink-500 dark:text-ink-400 font-normal"
		>
			按时间顺序排布的思考、笔记与技术沉淀。在这里，你可以找到所有历史文章的快照。
		</p>
	</header>

	<!-- Content List -->
	{#if $posts && $posts.length > 0}
		<div class="flex flex-col gap-4">
			{#each $posts as post, i}
				<!-- 使用 animate-in 实现简单的交错入场效果 (需 tailwindcss-animate 插件，若无则忽略) -->
				<a
					href="/posts/{post.shortUrl}"
					class="group block relative outline-none rounded-2xl transition-all duration-300 focus-visible:ring-2 focus-visible:ring-jade-500"
				>
					<Card
						variant="seamless"
						class="!p-0 overflow-hidden bg-white dark:bg-ink-950 border border-ink-100 dark:border-ink-900 shadow-sm hover:shadow-md hover:border-jade-200 dark:hover:border-jade-900/50 transition-all duration-300"
					>
						<div class="flex flex-col sm:flex-row sm:items-stretch">
							<!-- Left Decorator (Mobile hidden, Desktop visual anchor) -->
							<div
								class="hidden sm:block w-1.5 bg-ink-50 dark:bg-ink-900 group-hover:bg-jade-500 transition-colors duration-300"
							></div>

							<div class="flex-1 p-5 sm:p-7 flex flex-col sm:flex-row gap-5 sm:items-center">
								<!-- Main Content -->
								<div class="flex-1 space-y-3 min-w-0">
									<!-- Meta Row -->
									<div
										class="flex items-center gap-3 text-[10px] sm:text-xs font-mono tracking-widest text-ink-400 uppercase"
									>
										{#if post.createdAt}
											<div class="flex items-center gap-1.5">
												<Calendar size={12} class="opacity-70" />
												<time datetime={post.createdAt} class="mt-0.5"
													>{formatDate(post.createdAt)}</time
												>
											</div>
										{/if}

										<span class="text-ink-200 dark:text-ink-800">|</span>

										<Tag
											variant="jade"
											class="!py-0.5 !px-2 !text-[9px] !h-auto border border-jade-200 dark:border-jade-900 bg-jade-50 dark:bg-jade-950/30 text-jade-700 dark:text-jade-400"
										>
											专栏
										</Tag>
									</div>

									<!-- Title -->
									<h3
										class="font-serif text-lg sm:text-xl md:text-2xl font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-700 dark:group-hover:text-jade-400 transition-colors duration-200 truncate pr-4"
									>
										{post.title || '无标题文章'}
									</h3>

									<!-- Summary -->
									<p
										class="text-xs sm:text-sm leading-relaxed text-ink-500 dark:text-ink-400/80 line-clamp-2 sm:line-clamp-1 group-hover:text-ink-600 dark:group-hover:text-ink-300 transition-colors"
									>
										{post.summary || '这篇文章还没有摘要，点击阅读详情。'}
									</p>
								</div>

								<!-- Action Arrow (Right side) -->
								<div
									class="hidden sm:flex shrink-0 items-center justify-center pl-4 border-l border-transparent group-hover:border-ink-50 dark:group-hover:border-ink-900/50 transition-colors"
								>
									<div
										class="h-10 w-10 flex items-center justify-center rounded-full bg-ink-50 dark:bg-ink-900/50 text-ink-400 group-hover:bg-jade-500 group-hover:text-white group-hover:scale-110 group-hover:shadow-lg group-hover:shadow-jade-500/30 transition-all duration-300"
									>
										<ArrowRight
											size={18}
											class="group-hover:-rotate-45 transition-transform duration-300"
										/>
									</div>
								</div>
							</div>
						</div>
					</Card>
				</a>
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex justify-center pt-8 pb-12">
				<Pagination current={pagination.page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<!-- Empty State -->
		<div
			class="flex flex-col items-center justify-center py-32 text-center space-y-4 border-2 border-dashed border-ink-100 dark:border-ink-800/50 rounded-2xl bg-ink-50/50 dark:bg-ink-900/20"
		>
			<div class="relative">
				<div class="absolute -inset-4 bg-jade-500/10 rounded-full blur-xl animate-pulse"></div>
				<FileText size={48} class="relative text-ink-300 dark:text-ink-700" />
				<Sparkles size={20} class="absolute -top-2 -right-2 text-jade-400 animate-bounce" />
			</div>
			<div class="space-y-1">
				<h3 class="font-serif text-lg font-medium text-ink-900 dark:text-ink-100">暂无内容</h3>
				<p class="text-sm text-ink-500 dark:text-ink-500 max-w-xs mx-auto">
					似乎还没有发布任何文章，或者是筛选条件过于严格。
				</p>
			</div>
		</div>
	{/if}
</div>
