<script lang="ts">
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import type { OAuthProvider } from '$lib/features/auth/types';
	import { AuthCtx } from '$lib/features/auth/context';
	import { ChevronDown, ChevronUp } from 'lucide-svelte';

	let { onSelect, onToggleLogin } = $props<{
		onSelect: (provider: OAuthProvider) => void;
		onToggleLogin: () => void;
	}>();

	const providersStore = AuthCtx.selectModelData((data) => data?.oauth.providers ?? []);
	const errorStore = AuthCtx.selectModelData((data) => data?.oauth.error ?? '');
	const loadingKeyStore = AuthCtx.selectModelData((data) => data?.oauth.loadingKey ?? null);
	const showPasswordLoginStore = AuthCtx.selectModelData(
		(data) => data?.showPasswordLogin ?? false
	);

	const hasProviders = $derived.by(() => $providersStore.length > 0);
</script>

{#if hasProviders}
	<div class="space-y-3">
		<p class="text-xs font-mono text-ink-500">使用 OAuth 登录</p>
		<div class="grid gap-2">
			{#each $providersStore as provider}
				<Button
					variant="secondary"
					fullWidth
					loading={$loadingKeyStore === provider.key}
					disabled={$loadingKeyStore !== null && $loadingKeyStore !== provider.key}
					onclick={() => onSelect(provider)}
				>
					{provider.displayName}
				</Button>
			{/each}
		</div>
		{#if $errorStore}
			<p class="text-sm text-cinnabar-600 dark:text-cinnabar-400">{$errorStore}</p>
		{/if}
	</div>

	<div class="flex items-center gap-3 text-xs text-ink-500">
		<span class="h-px flex-1 bg-ink-100 dark:bg-ink-800"></span>
		<Button variant="ghost" size="sm" onclick={onToggleLogin}>
			{#if $showPasswordLoginStore}
				<ChevronUp class="size-4" />
			{:else}
				<ChevronDown class="size-4" />
			{/if}
			<span>{$showPasswordLoginStore ? '使用 OAuth 登录' : '使用账号密码登录'}</span>
		</Button>
		<span class="h-px flex-1 bg-ink-100 dark:bg-ink-800"></span>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
