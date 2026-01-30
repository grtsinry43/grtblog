<script lang="ts">
	import type { PostSummary } from '$lib/features/post/types';
	import { Calendar } from 'lucide-svelte';
	import { formatRelativeTime } from '$lib/shared/utils/date';

	let { post } = $props<{ post: PostSummary }>();
</script>

<a
	href="/posts/{post.shortUrl}"
	class="group flex flex-col gap-2 py-4 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 block w-full outline-none"
>
	<div class="flex items-start justify-between gap-4">
		<h3
			class="font-serif text-lg font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-200 line-clamp-1"
		>
			{post.title}
		</h3>

		<div
			class="flex-shrink-0 flex items-center gap-1.5 text-xs text-ink-400 dark:text-ink-500 font-mono mt-1.5"
		>
			<Calendar size={12} strokeWidth={1.5} />
			<span>{formatRelativeTime(post.createdAt)}</span>
		</div>
	</div>

	{#if post.summary}
		<p
			class="text-ink-500 dark:text-ink-400 text-sm leading-relaxed line-clamp-1 opacity-80 group-hover:opacity-100 transition-opacity"
		>
			{post.summary}
		</p>
	{/if}
</a>
