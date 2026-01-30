<script lang="ts">
	const { href = '', title = '', contentHtml = '' } = $props<{
		href?: string;
		title?: string;
		contentHtml?: string;
	}>();

	console.log('MarkdownLink contentHtml:', contentHtml);

	const isExternal = /^https?:\/\//i.test(href) || href.startsWith('//');
	const rel = isExternal ? 'noopener noreferrer' : undefined;
	const target = isExternal ? '_blank' : undefined;
</script>

<a class="md-link" {href} {title} {rel} {target}>
	{@html contentHtml}
</a>
