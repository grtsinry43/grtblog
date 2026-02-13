<script lang="ts">
	import { Heart } from 'lucide-svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { trackContentLike } from '$lib/features/analytics/api';
	import type { TrackLikeContentType } from '$lib/features/analytics/types';
	import { getOrCreateVisitorId, syncVisitorId } from '$lib/shared/visitor/visitor-id';

	interface Props {
		contentType: TrackLikeContentType;
		contentId: number;
		initialLikes?: number;
		className?: string;
	}

	let { contentType, contentId, initialLikes = 0, className = '' }: Props = $props();

	const queryClient = useQueryClient();
	let likes = $state(0);
	let liked = $state(false);

	$effect(() => {
		likes = initialLikes;
		liked = Boolean(queryClient.getQueryData(['analytics', 'liked', contentType, contentId]));
	});

	const mutation = createMutation(() => ({
		mutationFn: async () => {
			const visitorId = getOrCreateVisitorId();
			return trackContentLike(undefined, {
				contentType,
				contentId,
				visitorId: visitorId || undefined
			});
		},
		retry: false,
		onSuccess: (result) => {
			syncVisitorId(result?.visitorId);
			queryClient.setQueryData(['analytics', 'liked', contentType, contentId], true);
			liked = true;
			if (result?.affected) {
				likes += 1;
				toast.success('点赞成功');
				return;
			}
			toast.info('你已经点过赞了');
		},
		onError: () => {
			toast.error('点赞失败，请稍后重试');
		}
	}));

	const handleClick = () => {
		if (contentId <= 0 || liked || mutation.isPending) return;
		mutation.mutate();
	};

	const buttonClass = $derived(
		[
			'inline-flex items-center gap-1.5 text-[inherit] transition-colors outline-none',
			liked
				? 'text-cinnabar-500 dark:text-cinnabar-400'
				: 'text-ink-400 hover:text-cinnabar-500 dark:text-ink-400 dark:hover:text-cinnabar-400',
			className
		]
			.filter(Boolean)
			.join(' ')
	);
</script>

<button
	type="button"
	class={buttonClass}
	onclick={handleClick}
	disabled={liked || mutation.isPending || contentId <= 0}
	aria-label={liked ? '已点赞' : '点赞'}
>
	<Heart size={12} class={liked ? 'fill-current' : ''} />
	<span>喜欢 {likes}</span>
</button>
