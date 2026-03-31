<script lang="ts">
	import { onMount } from 'svelte';
	import { getState, setState, setupFallInterceptor } from './fall-effect';

	let showPopup = $state(false);

	onMount(() => {
		const now = new Date();
		if (now.getMonth() !== 3 || now.getDate() !== 1) return;

		const state = getState();
		if (state === 'done') return;

		if (state === 'fallen') {
			showPopup = true;
			return;
		}

		return setupFallInterceptor(() => {
			setState('fallen');
		});
	});

	function dismiss() {
		showPopup = false;
		setState('done');
	}
</script>

{#if showPopup}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-[99999] flex items-center justify-center bg-black/40 backdrop-blur-sm"
		onkeydown={(e) => e.key === 'Escape' && dismiss()}
		data-april-fools-popup
	>
		<div
			class="mx-4 max-w-sm rounded-default border border-ink-200/70 bg-ink-50/90 p-8 text-center shadow-float backdrop-blur-xl dark:border-ink-700/70 dark:bg-ink-900/90"
			role="dialog"
			aria-modal="true"
			aria-label="愚人节快乐"
		>
			<p class="mb-4 text-5xl">&#x1F389;</p>
			<h2 class="mb-3 font-serif text-2xl font-bold text-jade-600 dark:text-jade-400">
				愚人节快乐！
			</h2>
			<p class="mb-2 text-ink-700 dark:text-ink-300">
				哈哈，刚才那些掉下去的按钮和链接，有没有吓到你？
			</p>
			<p class="mb-6 text-sm text-ink-500 dark:text-ink-400">
				放心，这只是一个小小的愚人节彩蛋～现在一切恢复正常啦！
			</p>
			<button
				onclick={dismiss}
				class="cursor-pointer rounded-lg bg-jade-600 px-6 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-jade-700 dark:bg-jade-500 dark:hover:bg-jade-600"
			>
				好吧，算你狠
			</button>
		</div>
	</div>
{/if}
