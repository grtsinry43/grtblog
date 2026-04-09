<script lang="ts">
	interface Props {
		content: string;
		allowNeverShow?: boolean;
		onNeverShow?: () => void;
	}

	let { content, allowNeverShow = false, onNeverShow }: Props = $props();
	let checked = $state(false);

	function handleChange(event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		checked = target.checked;
		if (!checked) return;
		onNeverShow?.();
	}
</script>

<div class="space-y-2">
	<p class="whitespace-pre-wrap text-[13px] leading-6">{content}</p>
	{#if allowNeverShow}
		<label class="inline-flex items-center gap-2 text-[12px] text-ink-600/90 dark:text-ink-300/90">
			<input
				type="checkbox"
				{checked}
				onchange={handleChange}
				class="h-3.5 w-3.5 rounded border-ink-300 text-jade-600 focus:ring-jade-500 dark:border-ink-600 dark:bg-ink-900"
			/>
			<span>不再提示</span>
		</label>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
