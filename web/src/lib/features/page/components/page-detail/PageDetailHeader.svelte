<script lang="ts">
	import type { PageDetail } from '$lib/features/page/types';
	import { Calendar, Clock } from 'lucide-svelte';
	import { formatDateCN } from '$lib/shared/utils/date';
	import { calculateReadingTime, formatReadingTime } from '$lib/shared/utils/reading-time';
	import Badge from '$lib/ui/primitives/badge/Badge.svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';

	interface Props {
		page: PageDetail;
	}

	let { page }: Props = $props();
	const readingTime = $derived(calculateReadingTime(page.content));
</script>

<header class="max-w-4xl space-y-6">
	<div class="space-y-4">
		<div class="flex items-center gap-3">
			<Badge variant="soft">页面</Badge>
			<span class="font-mono text-[9px] tracking-[0.3em] text-ink-400 uppercase">站点内容</span>
		</div>

		<h1
			class="font-serif text-2xl leading-[1.2] font-medium tracking-tight text-ink-950 md:text-3xl lg:text-4xl dark:text-ink-50"
		>
			{page.title}
		</h1>

		<div
			class="flex flex-wrap items-center gap-5 font-mono text-[9px] tracking-widest text-ink-400 uppercase"
		>
			<span class="flex items-center gap-1.5">
				<Calendar size={12} />
				{formatDateCN(page.createdAt)}
			</span>
			<span class="flex items-center gap-1.5"><Clock size={12} /> {formatReadingTime(readingTime)}</span>
			<span class="flex items-center gap-1.5">浏览 {page.metrics?.views ?? 0}</span>
			<span aria-hidden="true" class="opacity-40">·</span>
			<ContentLikeButton
				contentType="page"
				contentId={page.id}
				likes={page.metrics?.likes ?? 0}
				className="inline-flex items-center gap-1.5"
			/>
			<span aria-hidden="true" class="opacity-40">·</span>
			<span class="flex items-center gap-1.5">评论 {page.metrics?.comments ?? 0}</span>
		</div>
	</div>
</header>
