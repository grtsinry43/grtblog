<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Card from './Card.svelte';

	const { Story } = defineMeta({
		component: Card,
		args: {
			variant: 'glass',
			hover: true,
			class: '',
			content: 'Card content'
		},
		argTypes: {
			variant: { control: 'select', options: ['glass', 'solid', 'seamless'] },
			hover: { control: 'boolean' },
			class: { control: 'text' },
			content: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Card variant={args.variant} hover={args.hover} class={args.class}>
			<div class="p-4 text-sm text-ink-600 dark:text-ink-300">{args.content}</div>
		</Card>
	{/snippet}
</Story>
