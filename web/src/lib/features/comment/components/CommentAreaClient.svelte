<script lang="ts">
	import { setContext } from 'svelte';
	import CommentArea from './CommentArea.svelte';
	import CommentList from './CommentList.svelte';
	import type { CommentNode } from '$lib/features/comment/types';

	let { areaId, commentsCount = 0 } = $props<{ areaId: number; commentsCount?: number }>();

	let replyingTo = $state<CommentNode | null>(null);

	setContext('COMMENT_CONTEXT', {
		get replyingTo() {
			return replyingTo;
		},
		setReplyingTo: (node: CommentNode | null) => {
			replyingTo = node;
		}
	});
</script>

<div class="mt-16 pt-10 border-t border-ink-100 dark:border-ink-800/50">
	<CommentArea {areaId} {commentsCount} />
	<CommentList {areaId} />
</div>
