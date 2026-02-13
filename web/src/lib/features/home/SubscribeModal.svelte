<script lang="ts">
	import { X, Mail, Check, BellRing, Newspaper, Coffee, Zap, Brain } from 'lucide-svelte';
	import { fly, fade, scale } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	let { isOpen = $bindable(false) } = $props<{ isOpen: boolean }>();

	let email = $state('');
	let isSubmitting = $state(false);
	let isSuccess = $state(false);

	type PreferenceKey = 'posts' | 'moments' | 'pages' | 'thinkings';

	let preferences = $state<Record<PreferenceKey, boolean>>({
		posts: true,
		moments: true,
		pages: false,
		thinkings: true
	});

	const options: Array<{
		id: PreferenceKey;
		name: string;
		desc: string;
		icon: typeof Newspaper;
		color: string;
	}> = [
		{
			id: 'posts',
			name: '文章',
			desc: '深度思考与技术分享',
			icon: Newspaper,
			color: 'text-jade-500'
		},
		{
			id: 'moments',
			name: '手记',
			desc: '生活碎片与即时感悟',
			icon: Coffee,
			color: 'text-amber-500'
		},
		{
			id: 'thinkings',
			name: '思考',
			desc: '碎片化的逻辑与灵感',
			icon: Brain,
			color: 'text-purple-500'
		},
		{ id: 'pages', name: '页面', desc: '站点重要更新通知', icon: Zap, color: 'text-blue-500' }
	];

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!email) return;

		isSubmitting = true;
		// 模拟 API 请求
		await new Promise((resolve) => setTimeout(resolve, 1500));
		isSubmitting = false;
		isSuccess = true;

		setTimeout(() => {
			isOpen = false;
			setTimeout(() => {
				isSuccess = false;
				email = '';
			}, 500);
		}, 2000);
	}

	function toggleOption(id: PreferenceKey) {
		preferences[id] = !preferences[id];
	}
</script>

{#if isOpen}
	<!-- Backdrop -->
	<button
		type="button"
		class="fixed inset-0 z-[100] bg-ink-950/20 dark:bg-black/40 backdrop-blur-sm"
		transition:fade={{ duration: 300 }}
		aria-label="关闭订阅弹窗"
		onclick={() => (isOpen = false)}
	></button>

	<!-- Modal -->
	<div
		class="fixed left-1/2 top-1/2 z-[101] w-[calc(100%-2rem)] max-w-lg -translate-x-1/2 -translate-y-1/2"
		transition:fly={{ y: 20, duration: 500, easing: cubicOut }}
	>
		<div
			class="overflow-hidden rounded-default border border-ink-200/60 bg-white shadow-deep dark:border-ink-800 dark:bg-ink-900 noise-surface"
		>
			<!-- Header -->
			<div class="relative p-8 pb-4">
				<button
					onclick={() => (isOpen = false)}
					class="absolute right-4 top-4 rounded-full p-2 text-ink-400 hover:bg-ink-100 dark:hover:bg-ink-800 transition-colors"
				>
					<X size={20} />
				</button>

				<div class="flex items-center gap-3 mb-2 text-jade-600 dark:text-jade-400">
					<BellRing size={24} />
					<h3 class="text-xl font-serif font-medium">订阅更新</h3>
				</div>
				<p class="text-sm text-ink-500">
					选择你感兴趣的内容，当有新产出时，我会第一时间发邮件通知你。
				</p>
			</div>

			{#if !isSuccess}
				<form onsubmit={handleSubmit} class="p-8 pt-4">
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-8">
						{#each options as opt (opt.id)}
							<button
								type="button"
								onclick={() => toggleOption(opt.id)}
								class="flex items-start gap-3 p-3 rounded-default border transition-all text-left group
									{preferences[opt.id as keyof typeof preferences]
									? 'border-jade-500/30 bg-jade-500/[0.03] dark:bg-jade-500/[0.05]'
									: 'border-ink-100 bg-ink-50/50 dark:border-ink-800 dark:bg-ink-950/30 opacity-60'}"
							>
								<div class="mt-0.5 {opt.color}">
									<opt.icon size={18} />
								</div>
								<div>
									<div
										class="text-sm font-medium {preferences[opt.id as keyof typeof preferences]
											? 'text-ink-900 dark:text-ink-100'
											: 'text-ink-500'}"
									>
										{opt.name}
									</div>
									<div class="text-[10px] text-ink-400 leading-tight mt-0.5">{opt.desc}</div>
								</div>
								{#if preferences[opt.id as keyof typeof preferences]}
									<div class="ml-auto text-jade-500" transition:scale>
										<Check size={14} />
									</div>
								{/if}
							</button>
						{/each}
					</div>

					<div class="relative group">
						<Mail
							size={18}
							class="absolute left-4 top-1/2 -translate-y-1/2 text-ink-400 group-focus-within:text-jade-500 transition-colors"
						/>
						<input
							type="email"
							bind:value={email}
							required
							placeholder="your-email@example.com"
							class="w-full bg-ink-50 dark:bg-ink-950 border border-ink-100 dark:border-ink-800 rounded-default py-3.5 pl-12 pr-4 text-sm focus:border-jade-500/50 transition-all outline-none"
						/>
					</div>

					<button
						type="submit"
						disabled={isSubmitting || !email}
						class="mt-6 w-full bg-ink-900 dark:bg-jade-600 text-white py-3.5 rounded-default font-medium text-sm hover:bg-jade-600 dark:hover:bg-jade-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
					>
						{#if isSubmitting}
							<div
								class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"
							></div>
							<span>正在处理...</span>
						{:else}
							<span>订阅到邮箱</span>
						{/if}
					</button>
				</form>
			{:else}
				<div class="p-12 text-center" in:fly={{ y: 10 }}>
					<div
						class="w-16 h-16 bg-jade-100 dark:bg-jade-900/30 text-jade-600 rounded-full flex items-center justify-center mx-auto mb-6"
					>
						<Check size={32} />
					</div>
					<h4 class="text-lg font-medium mb-2">订阅成功！</h4>
					<p class="text-sm text-ink-500">感谢你的关注，请留意你的收件箱（可能在垃圾箱哦）。</p>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
