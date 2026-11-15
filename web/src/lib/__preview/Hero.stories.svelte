<script module>
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import '@fontsource/google-sans/400.css';
	import '@fontsource/google-sans/500.css';
	import '@fontsource/google-sans/700.css';
	import '@fontsource/noto-serif-sc/400.css';
	import '@fontsource/noto-serif-sc/700.css';
	import '@fontsource-variable/victor-mono/index.css';
	import Hero from '$lib/features/home/Hero.svelte';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';

	const avatar =
		"data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='200' height='200'><defs><linearGradient id='g' x1='0' y1='0' x2='1' y2='1'><stop offset='0' stop-color='%2399f6e4'/><stop offset='1' stop-color='%23115e59'/></linearGradient></defs><rect width='200' height='200' fill='url(%23g)'/></svg>";

	const { Story } = defineMeta({ title: 'Preview/Home', component: Hero });
</script>

<Story name="Hero">
	<div class="min-h-screen bg-ink-50 px-8 pt-8">
		<Hero config={{ avatarUrl: avatar }} />
	</div>
</Story>

<Story name="HeroDark">
	<div class="dark">
		<div class="min-h-screen bg-ink-900 px-8 pt-8">
			<Hero config={{ avatarUrl: avatar }} />
		</div>
	</div>
</Story>

<Story name="Header">
	<div class="min-h-[400px] bg-ink-50 px-8 py-12">
		<div class="mx-auto max-w-5xl">
			<PageHeader
				title="文章归档"
				tag="Archive"
				meta="128 篇"
				subtitle="文字是思考的锚点"
				description="按时间顺序排布的思考、笔记与技术沉淀。在这里，你可以找到所有历史文章的快照。"
			/>
		</div>
	</div>
</Story>
