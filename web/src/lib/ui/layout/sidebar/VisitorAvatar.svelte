<script lang="ts">
	import { userStore } from '$lib/shared/stores/userStore';
	import { authLoginModal } from '$lib/shared/actions/auth-login-modal';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { UserIcon } from 'lucide-svelte';	import Button from '$lib/ui/primitives/button/Button.svelte';
</script>

{#if $userStore.isLogin}
	<Button
		variant="icon"
		onclick={() => windowStore.open('用户中心', null, 'user-center')}
		class="h-10 w-10 !p-0 rounded-default overflow-hidden ring-1 ring-ink-200 hover:ring-ink-300 dark:ring-ink-700 dark:hover:ring-ink-500"
		aria-label="用户中心"
	>
		{#if $userStore.userInfo?.avatar !== ''}
			<img src={$userStore.userInfo?.avatar} alt="User avatar" class="h-full w-full object-cover" />
		{:else}
			<div class="h-full w-full flex items-center justify-center">
				<span>{$userStore.userInfo?.nickname?.charAt(0).toUpperCase() || 'U'}</span>
			</div>
		{/if}
	</Button>
{:else}
	<Button
		variant="icon"
		class="h-10 w-10 rounded-default text-ink-400 hover:bg-ink-100 hover:text-ink-900 dark:hover:bg-ink-800 dark:hover:text-ink-100"
		aria-label="登录"
		action={authLoginModal}
		actionOptions={{ source: 'sidebar-avatar' }}
	>
		<UserIcon class="h-5 w-5" />
	</Button>
{/if}
