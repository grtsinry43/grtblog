<script lang="ts">
	import { Popover } from 'bits-ui';
	import { ALargeSmall } from 'lucide-svelte';
	import { readingPrefs, type ReadingFont, type ReadingSize } from '../prefs.svelte';

	interface Props {
		/** `inline`: text button for the article meta row; `icon`: square button for toolbars. */
		variant?: 'inline' | 'icon';
		class?: string;
	}

	let { variant = 'inline', class: className = '' }: Props = $props();

	const sizes: { value: ReadingSize; label: string; sample: string }[] = [
		{ value: 'sm', label: '小', sample: 'text-[13px]' },
		{ value: 'md', label: '中', sample: 'text-[16px]' },
		{ value: 'lg', label: '大', sample: 'text-[19px]' }
	];
	const fonts: { value: ReadingFont; label: string; sample: string }[] = [
		{ value: 'sans', label: '无衬线', sample: 'font-sans' },
		{ value: 'serif', label: '衬线', sample: 'font-serif' }
	];

	const segment =
		'flex flex-1 items-center justify-center rounded-[2px] py-1.5 text-ink-500 transition-colors duration-200 hover:text-ink-900 dark:text-ink-400 dark:hover:text-ink-100';
	const segmentActive =
		'bg-white text-jade-700 shadow-subtle hover:text-jade-700 dark:bg-ink-700 dark:text-jade-300 dark:hover:text-jade-300';
</script>

<Popover.Root>
	<Popover.Trigger
		aria-label="阅读设置"
		title="阅读设置"
		class={variant === 'icon'
			? `inline-flex h-9 w-9 items-center justify-center rounded-default text-ink-500 transition-colors hover:bg-ink-100/50 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-ink-100 ${className}`
			: `inline-flex items-center gap-1.5 rounded-default transition-colors hover:text-jade-600 dark:hover:text-jade-400 ${className}`}
	>
		<ALargeSmall size={variant === 'icon' ? 18 : 13} />
		{#if variant === 'inline'}<span>阅读设置</span>{/if}
	</Popover.Trigger>
	<Popover.Portal>
		<Popover.Content
			sideOffset={8}
			align="end"
			class="z-(--z-popover) w-60 rounded-default border border-ink-200/70 bg-paper/95 p-3 shadow-float backdrop-blur-md animate-in fade-in zoom-in-95 dark:border-ink-700 dark:bg-ink-900/95 dark:shadow-glass-dark"
		>
			<div class="space-y-3">
				<div class="space-y-1.5">
					<p class="font-mono text-micro text-ink-500 dark:text-ink-400">字号</p>
					<div
						role="radiogroup"
						aria-label="字号"
						class="flex gap-0.5 rounded-default bg-ink-100 p-0.5 dark:bg-ink-800"
					>
						{#each sizes as option (option.value)}
							{@const active = readingPrefs.size === option.value}
							<button
								type="button"
								role="radio"
								aria-checked={active}
								aria-label={`${option.label}号字`}
								class="{segment} h-9 font-serif {option.sample} {active ? segmentActive : ''}"
								onclick={() => readingPrefs.update({ size: option.value })}
							>
								A
							</button>
						{/each}
					</div>
				</div>

				<div class="space-y-1.5">
					<p class="font-mono text-micro text-ink-500 dark:text-ink-400">正文字体</p>
					<div
						role="radiogroup"
						aria-label="正文字体"
						class="flex gap-0.5 rounded-default bg-ink-100 p-0.5 dark:bg-ink-800"
					>
						{#each fonts as option (option.value)}
							{@const active = readingPrefs.font === option.value}
							<button
								type="button"
								role="radio"
								aria-checked={active}
								class="{segment} text-[13px] {option.sample} {active ? segmentActive : ''}"
								onclick={() => readingPrefs.update({ font: option.value })}
							>
								{option.label}
							</button>
						{/each}
					</div>
				</div>
			</div>
		</Popover.Content>
	</Popover.Portal>
</Popover.Root>
