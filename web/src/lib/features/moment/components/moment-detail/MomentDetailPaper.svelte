<script lang="ts">
	import type { MomentDetail } from '$lib/features/moment/types';
	import DetailAiSummary from '$lib/ui/detail/DetailAiSummary.svelte';
	import DetailCommentSection from '$lib/ui/detail/DetailCommentSection.svelte';
	import DetailMarkdownContent from '$lib/ui/detail/DetailMarkdownContent.svelte';
	import DetailActionBar from '$lib/ui/detail/DetailActionBar.svelte';
	import { formatDateCN, isDifferentDay } from '$lib/shared/utils/date';
	import { Sun } from 'lucide-svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import TagList from '$lib/features/tag/components/TagList.svelte';
	import { RollingNumber } from '$lib/ui/animation';

	interface Props {
		moment: MomentDetail;
		dateStr: string;
		dateNo: string;
		onActiveAnchorChange: (anchor: string | null) => void;
		onContentRootChange: (node: HTMLElement | null) => void;
	}

	let { moment, dateStr, dateNo, onActiveAnchorChange, onContentRootChange }: Props =
		$props();

	const showUpdated = $derived(isDifferentDay(moment.createdAt, moment.contentUpdatedAt));
</script>

<div
	class="
		bg-transparent border-0 shadow-none rounded-none
		md:bg-ink-50 dark:md:bg-ink-900
		md:shadow-[0_4px_30px_-8px_rgba(0,0,0,0.06)] dark:md:shadow-none
		md:border md:border-ink-200/80 dark:md:border-ink-200/10
		px-0 py-10 md:p-20 md:rounded-sm relative overflow-hidden md:min-h-[80vh]
		moment-vt
	"
	style:view-transition-name={`moment-${moment.id}`}
>
	<div class="relative z-10">
		<header class="mb-12 flex flex-col gap-6">
			<div class="flex items-start justify-between gap-3 border-b border-ink-800/10 pb-4">
				<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs font-mono text-ink-800/40 dark:text-ink-200/40">
					<span>NO. {dateNo}</span>
					<span>—</span>
					<span class="font-serif text-cinnabar-500">手记</span>
					<span>—</span>
					<span>{dateStr}</span>
					{#if showUpdated}<span class="text-ink-400/70">（更新于 {formatDateCN(moment.contentUpdatedAt)}）</span>{/if}
				</div>
				<div class="shrink-0 text-ink-800/40 dark:text-ink-200/40">
					<Sun size={18} stroke-width={1.5} />
				</div>
			</div>

			<h1
				class="text-xl md:text-3xl font-serif font-bold text-ink-900 dark:text-ink-50 leading-[1.2]"
			>
				{moment.title}
			</h1>
			<div
				class="flex flex-wrap items-center gap-3 text-[11px] font-mono tracking-[0.16em] text-ink-800/45 dark:text-ink-200/45 uppercase"
			>
				<span class="flex items-center gap-1.5">浏览 <RollingNumber value={moment.metrics?.views ?? 0} /></span>
				<span aria-hidden="true" class="opacity-40">·</span>
				<ContentLikeButton
					contentType="moment"
					contentId={moment.id}
					likes={moment.metrics?.likes ?? 0}
					className="inline-flex items-center gap-1.5"
				/>
				<span aria-hidden="true" class="opacity-40">·</span>
				<span class="flex items-center gap-1.5">评论 <RollingNumber value={moment.metrics?.comments ?? 0} /></span>
			</div>

			<TagList tags={moment.topics} />
		</header>

		{#if moment.aiSummary}
			<DetailAiSummary summary={moment.aiSummary} />
		{/if}

		<DetailMarkdownContent
			content={moment.content}
			toc={moment.toc}
			className="max-w-none text-ink-900/80 dark:text-ink-200/90 font-serif text-justify text-[15px]"
			{onContentRootChange}
			{onActiveAnchorChange}
		/>

		<DetailActionBar
			contentType="moment"
			contentId={moment.id}
			likes={moment.metrics?.likes ?? 0}
			comments={moment.metrics?.comments ?? 0}
			tone="cinnabar"
		/>

		<div class="mt-24 flex justify-center opacity-40">
			<div
				class="w-24 h-24 border-2 border-dashed border-ink-800 dark:border-ink-200 rounded-full flex items-center justify-center rotate-12"
			>
				<div class="text-center text-ink-800 dark:text-ink-200">
					<div class="text-[9px] uppercase tracking-widest mb-1">手记</div>
					<div class="font-serif font-bold text-lg">记</div>
					<div class="text-[9px] mt-1">{dateStr}</div>
				</div>
			</div>
		</div>

		<DetailCommentSection
			commentAreaId={moment.commentAreaId}
			commentsCount={moment.metrics?.comments ?? 0}
			fediverseObjectUrl={moment.activityPubObjectId}
			containerClass="mt-16 pt-10 border-t border-ink-200/50 dark:border-ink-700/30"
			fallbackText="Loading comments..."
			fallbackSize="w-6 h-6"
			fallbackContainerClass="flex justify-center py-20"
		/>
	</div>
</div>
