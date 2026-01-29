<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Badge from './Badge.svelte';

	const { Story } = defineMeta({
		component: Badge,
		args: {
			children: 'Badge',
			variant: 'soft',
			class: ''
		},
		argTypes: {
			children: { control: 'text' },
			variant: { control: 'select', options: ['soft', 'ghost', 'dot'] },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Badge variant={args.variant} class={args.class}>{args.children}</Badge>
	{/snippet}
</Story>
