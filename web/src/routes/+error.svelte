<script lang="ts">
	import { resolve } from '$app/paths';
	import { ArrowLeft, Compass, RefreshCw } from 'lucide-svelte';

	type ErrorPayload = {
		message?: string;
	};

	// eslint-disable-next-line svelte/valid-prop-names-in-kit-pages
	let { status, error }: { status: number; error: ErrorPayload } = $props();

	const isNotFound = $derived(status === 404);
	const title = $derived(isNotFound ? '页面走丢了' : '页面暂时不可用');
	const summary = $derived(
		isNotFound
			? '你访问的链接不存在，或者已经被迁移到新的地址。'
			: '服务器返回了异常响应，请稍后重试。'
	);
	const detail = $derived(error?.message || (isNotFound ? 'Not Found' : 'Unexpected Server Error'));

	function goBack() {
		history.back();
	}

	function reloadPage() {
		location.reload();
	}
</script>

<section class="mx-auto max-w-3xl px-6 py-16 md:py-24">
	<div
		class="relative overflow-hidden rounded-default border border-ink-200/70 bg-gradient-to-br from-ink-50 to-jade-50/40 p-8 shadow-subtle dark:border-ink-700/60 dark:from-ink-900 dark:to-ink-800/60 md:p-10"
	>
		<div
			class="pointer-events-none absolute -right-10 -top-10 h-32 w-32 rounded-full bg-jade-400/15 blur-2xl"
			aria-hidden="true"
		></div>

		<h1 class="font-serif text-3xl text-ink-900 dark:text-ink-100 md:text-4xl">{title}</h1>
		<p class="mt-3 text-sm leading-relaxed text-ink-600 dark:text-ink-300 md:text-base">
			{summary}
		</p>

		<div
			class="mt-6 rounded-default border border-ink-200/80 bg-white/70 px-4 py-3 text-xs text-ink-500 dark:border-ink-700/70 dark:bg-ink-900/60 dark:text-ink-400"
		>
			{detail}
		</div>

		<div class="mt-8 flex flex-wrap items-center gap-3">
			<a
				href={resolve('/')}
				class="inline-flex items-center gap-2 rounded-default border border-jade-500/30 bg-jade-500/10 px-4 py-2 text-sm text-jade-700 transition-colors hover:bg-jade-500/20 dark:text-jade-300"
			>
				<Compass size={14} />
				回到首页
			</a>
			<button
				class="inline-flex cursor-pointer items-center gap-2 rounded-default border border-ink-200/80 bg-white px-4 py-2 text-sm text-ink-700 transition-colors hover:bg-ink-100 dark:border-ink-700 dark:bg-ink-800 dark:text-ink-200 dark:hover:bg-ink-700"
				onclick={goBack}
			>
				<ArrowLeft size={14} />
				返回上一页
			</button>
			<button
				class="inline-flex cursor-pointer items-center gap-2 rounded-default border border-ink-200/80 bg-white px-4 py-2 text-sm text-ink-700 transition-colors hover:bg-ink-100 dark:border-ink-700 dark:bg-ink-800 dark:text-ink-200 dark:hover:bg-ink-700"
				onclick={reloadPage}
			>
				<RefreshCw size={14} />
				刷新页面
			</button>
		</div>
	</div>
</section>
