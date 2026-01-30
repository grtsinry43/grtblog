<script lang="ts">
	const { lang = 'text', contentHtml = '' } = $props<{
		lang?: string;
		contentHtml?: string;
	}>();
</script>

<div class="md-codeblock" data-lang={lang}>
	<div class="md-codeblock__header">
		<span class="md-codeblock__lang">{lang || 'text'}</span>
	</div>
	<div class="md-codeblock__body">
		{@html contentHtml}
	</div>
</div>
