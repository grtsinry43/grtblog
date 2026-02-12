<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Pagination from './Pagination.svelte';

	const { Story } = defineMeta({
		component: Pagination,
		args: {
			current: 1,
			total: 6,
			class: ''
		},
		argTypes: {
			current: { control: 'number' },
			total: { control: 'number' },
			class: { control: 'text' }
		}
	});
</script>

<Story name="Default">
	{#snippet template(args)}
		<Pagination
			current={args.current}
			total={args.total}
			class={args.class}
			onPageChange={() => {}}
		/>
	{/snippet}
</Story>
