<script lang="ts">
	import { Calendar, Eye, Heart, MessageCircle } from 'lucide-svelte';
	import type { MomentSummary } from '$lib/features/moment/types';

	let { moment } = $props<{ moment: MomentSummary }>();

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days < 1) return '今天';
		if (days < 30) return `大约 ${days} 天前`;
		if (days < 365) return `大约 ${Math.floor(days / 30)} 个月前`;
		return `${date.getFullYear()}年`;
	};
</script>

<a
	href="/moments/{moment.shortUrl}"
	class="group relative flex flex-col gap-3 px-6 py-6 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 block w-full outline-none"
>
	<!-- Title -->
	<h2
		class="font-serif text-lg font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-200"
	>
		{moment.title}
	</h2>

	<!-- Summary -->
	<p class="text-ink-500 dark:text-ink-400 text-sm leading-relaxed line-clamp-3 font-serif">
		{moment.summary}
	</p>

	<!-- Meta Row -->
	<div class="flex items-center gap-6 text-xs text-ink-400 dark:text-ink-500 mt-2 font-mono">
		<!-- Date -->
		<div class="flex items-center gap-1.5">
			<Calendar size={14} strokeWidth={1.5} />
			<span>{formatDate(moment.createdAt)}</span>
		</div>

		<!-- Views -->
		<div class="flex items-center gap-1.5">
			<Eye size={14} strokeWidth={1.5} />
			<span>{moment.views}</span>
		</div>

		<!-- Likes -->
		<div class="flex items-center gap-1.5">
			<Heart size={14} strokeWidth={1.5} />
			<span>{moment.likes}</span>
		</div>
        
        <!-- Comments -->
		<div class="flex items-center gap-1.5">
			<MessageCircle size={14} strokeWidth={1.5} />
			<span>{moment.comments}</span>
		</div>
	</div>
</a>
