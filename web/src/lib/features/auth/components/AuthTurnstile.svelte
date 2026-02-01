<script lang="ts">
	import { AuthCtx } from '$lib/features/auth/context';
	import { turnstileWidget } from '$lib/shared/actions/turnstile';

	let {
		onToken,
		onExpired,
		onError
	} = $props<{
		onToken: (token: string) => void;
		onExpired: () => void;
		onError: (error: unknown) => void;
	}>();

	const turnstileStore = AuthCtx.selectModelData(
		(data) => data?.turnstile ?? null,
		{
			equals: (a, b) =>
				a?.enabled === b?.enabled && a?.siteKey === b?.siteKey && a?.error === b?.error
		}
	);
	const widgetOptions = $derived.by(() => ({
		siteKey: $turnstileStore?.siteKey ?? '',
		onToken,
		onExpired,
		onError
	}));
</script>

{#if $turnstileStore?.enabled}
	<div class="space-y-2">
		<label class="text-xs font-mono text-ink-500">人机验证</label>
		{#if $turnstileStore.siteKey}
			<div
				class="rounded-default border border-ink-100 bg-white/70 p-3 dark:border-ink-800 dark:bg-ink-900/40"
				use:turnstileWidget={widgetOptions}
			></div>
		{:else}
			<p class="text-xs text-ink-500">Turnstile 未配置，请联系管理员。</p>
		{/if}
		{#if $turnstileStore.error}
			<p class="text-xs text-cinnabar-600 dark:text-cinnabar-400">{$turnstileStore.error}</p>
		{/if}
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
