<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import { Link2 } from 'lucide-svelte';
	import FriendTimelineCard from './FriendTimelineCard.svelte';
	import { friendTimelineListCtx } from '$lib/features/friend-timeline/context';
	import { goto } from '$app/navigation';

	type PaginationData = {
		total: number;
		page: number;
		size: number;
	};

	interface Props {
		basePath?: string;
	}

	let {
		basePath = '/friends-timeline'
	}: Props = $props();

	import type { FriendTimelineListResponse } from '../types';

	const itemsStore = friendTimelineListCtx.selectModelData((state: FriendTimelineListResponse | null) => state?.items ?? []);
	const totalStore = friendTimelineListCtx.selectModelData((state: FriendTimelineListResponse | null) => state?.total ?? 0);
	const pageStore = friendTimelineListCtx.selectModelData((state: FriendTimelineListResponse | null) => state?.page ?? 1);
	const sizeStore = friendTimelineListCtx.selectModelData((state: FriendTimelineListResponse | null) => state?.size ?? 10);

	let items = itemsStore;
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
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		goto(resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`));
	};
</script>

<div class="w-full max-w-4xl mx-auto py-4 space-y-12">
	<!-- Content List -->
	{#if $items && $items.length > 0}
		<StaggerList
			class="flex flex-col max-w-2xl mx-auto w-full"
			role="list"
			aria-label="朋友圈列表"
			staggerDelay={50}
			duration={400}
			y={15}
			spring={false}
			key="friend-timeline"
		>
			{#each $items as item (item.url)}
				<FriendTimelineCard {item} />
			{/each}
		</StaggerList>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex justify-center pt-6 pb-8 sm:pt-8 sm:pb-12">
				<Pagination current={pagination.page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<!-- Empty State -->
		<div
			class="flex flex-col items-center justify-center py-16 sm:py-32 text-center space-y-4 border-2 border-dashed border-ink-100 dark:border-ink-800/50 rounded-2xl bg-ink-50/50 dark:bg-ink-900/20 max-w-2xl mx-auto"
		>
			<div class="relative">
				<div class="absolute -inset-4 bg-jade-500/10 rounded-full blur-xl animate-pulse"></div>
				<Link2 size={48} class="relative text-ink-300 dark:text-ink-700" />
			</div>
			<div class="space-y-1">
				<h3 class="font-serif text-lg font-medium text-ink-900 dark:text-ink-100">暂无动态</h3>
				<p class="text-sm text-ink-500 dark:text-ink-500 max-w-xs mx-auto">
					目前还没有收到任何好友的新鲜事，或许他们都在潜心创作呢。过段时间再来看看吧。
				</p>
			</div>
		</div>
	{/if}
</div>
