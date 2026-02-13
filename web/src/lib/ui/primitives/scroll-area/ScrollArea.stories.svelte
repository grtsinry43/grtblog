<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ScrollArea from './ScrollArea.svelte';

	const { Story } = defineMeta({
		component: ScrollArea,
		args: {
			orientation: 'vertical',
			class: 'h-48 w-80',
			viewportClass: 'p-4',
			scrollbarClass: '',
			thumbClass: '',
			cornerClass: ''
		},
		argTypes: {
			orientation: { control: 'select', options: ['vertical', 'horizontal', 'both'] },
			class: { control: 'text' },
			viewportClass: { control: 'text' },
			scrollbarClass: { control: 'text' },
			thumbClass: { control: 'text' },
			cornerClass: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<ScrollArea
			orientation={args.orientation}
			class={args.class}
			viewportClass={args.viewportClass}
			scrollbarClass={args.scrollbarClass}
			thumbClass={args.thumbClass}
			cornerClass={args.cornerClass}
		>
			<div class="space-y-3 text-[13px] text-ink-800 dark:text-ink-200">
				{#each Array.from({ length: 18 }, (_, index) => index + 1) as row (row)}
					<div
						class="rounded-default border border-ink-100/70 bg-white/70 px-3 py-2 dark:border-ink-800/40 dark:bg-ink-900/40"
					>
						Row {row} · Scroll me
					</div>
				{/each}
			</div>
		</ScrollArea>
	{/snippet}
</Story>
