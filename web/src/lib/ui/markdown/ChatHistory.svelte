<script lang="ts">
	import type { SvmdComponentNode } from 'svmarkdown';
	import { extractPlainTextFromNodes } from '$lib/shared/markdown/component-body';

	let { node } = $props<{
		node?: SvmdComponentNode;
	}>();

	const title = $derived(node?.props?.title || 'Conversation');
	const bodyText = $derived(extractPlainTextFromNodes(node?.children));

	// 更加鲁棒的解析逻辑
	const messages = $derived.by(() => {
		if (!bodyText) return [];
		const raw = String(bodyText);
		// 处理各种可能的换行符转义
		const lines = raw.replace(/\\n/g, '\n').split('\n');
		return lines
			.map((line) => {
				const index = line.indexOf('|');
				if (index === -1) return null;
				return {
					role: line.substring(0, index).trim(),
					content: line.substring(index + 1).trim()
				};
			})
			.filter(Boolean);
	});

	const isMe = (role: string) => ['me', 'user', 'admin', '我', '本人'].includes(role.toLowerCase());
</script>

<div
	class="chat-wrapper not-prose my-8 overflow-hidden rounded-default border border-ink-200/60 bg-white dark:border-ink-800/60 dark:bg-ink-950/20"
>
	<!-- 简洁的 Header -->
	<div
		class="flex items-center justify-between border-b border-ink-100/60 bg-ink-50/30 px-5 py-2.5 dark:border-ink-800/60 dark:bg-ink-900/40"
	>
		<span class="text-[10px] font-bold tracking-widest text-ink-500 uppercase">{title}</span>
		<div class="flex gap-1">
			<div class="h-1 w-1 rounded-full bg-ink-300 dark:bg-ink-700"></div>
			<div class="h-1 w-1 rounded-full bg-ink-300 dark:bg-ink-700"></div>
		</div>
	</div>

	<!-- 聊天流：去掉复杂装饰，纯粹的气泡 -->
	<div class="flex flex-col gap-4 p-5">
		{#each messages as msg, index (`${index}-${msg!.role}-${msg!.content}`)}
			<div class="flex flex-col {isMe(msg!.role) ? 'items-end' : 'items-start'}">
				<!-- 极简角色标识 -->
				<span class="mb-1 text-[9px] font-bold text-ink-400 uppercase tracking-tighter">
					{msg!.role}
				</span>

				<!-- 气泡：更加纤细的圆角和背景 -->
				<div
					class="max-w-[90%] rounded-[18px] px-4 py-2 text-[13.5px] leading-relaxed
					{isMe(msg!.role)
						? 'rounded-tr-[4px] bg-jade-500 text-white shadow-sm'
						: 'rounded-tl-[4px] bg-ink-100 text-ink-800 dark:bg-ink-800 dark:text-ink-100'}"
				>
					{msg!.content}
				</div>
			</div>
		{:else}
			<div class="py-4 text-center text-xs italic text-ink-400">Waiting for messages...</div>
		{/each}
	</div>
</div>
