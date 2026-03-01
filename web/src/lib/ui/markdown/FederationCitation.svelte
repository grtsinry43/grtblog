<script lang="ts">
	let {
		instance = '',
		'post-id': postId = '',
		title = '',
		summary = '',
		url = '',
		'cover-image': coverImage = '',
		'author-name': authorName = '',
		status = 'pending'
	} = $props<{
		instance?: string;
		'post-id'?: string;
		title?: string;
		summary?: string;
		url?: string;
		'cover-image'?: string;
		'author-name'?: string;
		status?: string;
	}>();

	const isApproved = $derived(status === 'approved');
	const resolvedUrl = $derived(url || `https://${instance}/posts/${postId}`);
	const displayTitle = $derived(title || resolvedUrl);
</script>

{#if isApproved}
	<!-- Render as article card -->
	<a
		class="group not-prose relative my-5 flex items-stretch overflow-hidden rounded-default border border-ink-200/60 bg-white/40 transition-all duration-300 hover:border-jade-400/50 hover:bg-white hover:shadow-subtle dark:border-ink-800/60 dark:bg-ink-900/30 dark:hover:border-jade-800/80 dark:hover:bg-ink-900/10"
		href={resolvedUrl}
		target="_blank"
		rel="noreferrer"
	>
		<!-- Left accent bar -->
		<div class="w-[3px] bg-jade-400 transition-colors duration-300 group-hover:bg-jade-500 dark:bg-jade-600"></div>

		<!-- Content -->
		<div class="flex flex-1 flex-col justify-center px-4 py-3">
			<div class="mb-1 flex items-center gap-1.5">
				<svg class="h-3 w-3 text-jade-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
					<path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
				</svg>
				<span class="text-[9px] font-bold tracking-[0.1em] text-jade-600 uppercase dark:text-jade-400">
					引用自 {instance}
				</span>
			</div>

			<h4 class="truncate text-[13px] font-bold text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-400">
				{displayTitle}
			</h4>

			{#if summary}
				<div class="mt-1.5 line-clamp-2 text-[11px] text-ink-500 dark:text-ink-400">
					{summary}
				</div>
			{/if}

			{#if authorName}
				<div class="mt-1 text-[10px] text-ink-400 dark:text-ink-500">
					{authorName}
				</div>
			{/if}
		</div>

		<!-- Cover image -->
		{#if coverImage}
			<div class="relative w-24 shrink-0 overflow-hidden border-l border-ink-100/50 dark:border-ink-800/50 sm:w-32">
				<img src={coverImage} alt="" class="h-full w-full object-cover grayscale-[0.3] transition-all duration-500 group-hover:scale-105 group-hover:grayscale-0" loading="lazy" />
				<div class="absolute inset-0 bg-gradient-to-r from-white/20 via-transparent to-transparent dark:from-black/20"></div>
			</div>
		{:else}
			<div class="flex w-12 items-center justify-center opacity-0 transition-all duration-300 -translate-x-2 group-hover:translate-x-0 group-hover:opacity-100">
				<svg class="h-4 w-4 text-jade-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
					<path d="M9 18l6-6-6-6" />
				</svg>
			</div>
		{/if}
	</a>
{:else}
	<!-- Fallback: plain link -->
	<a
		class="text-jade-600 underline decoration-jade-300 underline-offset-2 hover:text-jade-700 dark:text-jade-400 dark:decoration-jade-700 dark:hover:text-jade-300"
		href={resolvedUrl}
		target="_blank"
		rel="noreferrer"
	>
		{displayTitle}
	</a>
{/if}
