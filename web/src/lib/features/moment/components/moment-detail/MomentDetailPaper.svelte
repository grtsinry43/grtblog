<script lang="ts">
	import type { MomentDetail } from '$lib/features/moment/types';
	import DetailCommentSection from '$lib/shared/components/detail/DetailCommentSection.svelte';
	import DetailMarkdownContent from '$lib/shared/components/detail/DetailMarkdownContent.svelte';
	import { Sun } from 'lucide-svelte';

	interface Props {
		moment: MomentDetail;
		dateStr: string;
		dateNo: string;
		column: string;
		onActiveAnchorChange: (anchor: string | null) => void;
		onContentRootChange: (node: HTMLElement | null) => void;
	}

	let { moment, dateStr, dateNo, column, onActiveAnchorChange, onContentRootChange }: Props =
		$props();
</script>

<div
	class="
		bg-ink-50 md:bg-[#fbf9f4] dark:bg-ink-900 dark:md:bg-ink-900
		shadow-[0_4px_30px_-8px_rgba(0,0,0,0.06)] dark:shadow-none
		border border-ink-200/80 dark:border-ink-200/10
		px-8 py-12 md:p-20 rounded-sm relative overflow-hidden min-h-[80vh]
		transition-colors duration-500
	"
	style:view-transition-name={`moment-${moment.id}`}
>
	<div class="absolute inset-0 bg-noise opacity-30 pointer-events-none"></div>

	<div class="relative z-10">
		<header class="mb-12 flex flex-col gap-6">
			<div class="flex items-center justify-between border-b border-ink-800/10 pb-4">
				<div class="flex items-center gap-3 text-xs font-mono text-ink-800/40 dark:text-ink-200/40">
					<span>NO. {dateNo}</span>
					<span>—</span>
					<span class="font-serif text-cinnabar-500">{column}</span>
				</div>
				<div class="text-ink-800/40 dark:text-ink-200/40">
					<Sun size={18} stroke-width={1.5} />
				</div>
			</div>

			<h1
				class="text-3xl md:text-5xl font-serif font-bold text-ink-900 dark:text-ink-50 leading-[1.2]"
			>
				{moment.title}
			</h1>
		</header>

		<DetailMarkdownContent
			content={moment.content}
			toc={moment.toc}
			className="max-w-none text-ink-900/80 dark:text-ink-200/90 font-serif text-justify text-[15px]"
			{onContentRootChange}
			{onActiveAnchorChange}
		/>

		<div class="mt-24 flex justify-center opacity-40">
			<div
				class="w-24 h-24 border-2 border-dashed border-ink-800 dark:border-ink-200 rounded-full flex items-center justify-center rotate-12"
			>
				<div class="text-center text-ink-800 dark:text-ink-200">
					<div class="text-[9px] uppercase tracking-widest mb-1">审阅</div>
					<div class="font-serif font-bold text-lg">阅</div>
					<div class="text-[9px] mt-1">{dateStr}</div>
				</div>
			</div>
		</div>

		<DetailCommentSection
			commentAreaId={moment.commentAreaId}
			commentsCount={moment.metrics?.comments ?? 0}
			containerClass="mt-16 pt-10 border-t border-ink-200/50 dark:border-ink-700/30"
			fallbackText="Loading comments..."
			fallbackSize="w-6 h-6"
			fallbackContainerClass="flex justify-center py-20"
		/>
	</div>
</div>
