<script lang="ts">
	import { ArrowLeft, Share2, ArrowUp, Home, MessageSquare } from 'lucide-svelte';
	import { fade, fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import Button from '$lib/ui/primitives/button/Button.svelte';
	import { browser } from '$app/environment';
	import { toast } from 'svelte-sonner';

	interface Props {
		title: string;
		showThreshold?: number;
		showCommentShortcut?: boolean;
	}

	let { title, showThreshold = 300, showCommentShortcut = true }: Props = $props();

	let scrollY = $state(0);
	let clientHeight = $state(0);

	let showHeader = $derived(scrollY > showThreshold);
	let progress = $derived.by(() => {
		if (!browser || !showHeader) return 0;
		const scrollHeight = document.documentElement.scrollHeight;
		const totalHeight = scrollHeight - clientHeight;
		if (totalHeight <= 0) return 0;
		return Math.min(100, Math.max(0, (scrollY / totalHeight) * 100));
	});

	const scrollToTop = () => {
		if (!browser) return;
		window.scrollTo({ top: 0, behavior: 'smooth' });
	};

	const scrollToComments = () => {
		if (!browser) return;
		const commentSection = document.querySelector('#comments');
		if (commentSection) {
			commentSection.scrollIntoView({ behavior: 'smooth' });
		} else {
			toast.info('该页面没有评论区');
		}
	};

	function goBack() {
		if (!browser) return;
		if (window.history.length > 1) {
			history.back();
		} else {
			window.location.href = '/';
		}
	}

	async function handleShare() {
		if (!browser) return;
		const shareData = {
			title: title,
			url: window.location.href
		};

		try {
			if (navigator.share) {
				await navigator.share(shareData);
			} else {
				await navigator.clipboard.writeText(window.location.href);
				toast.success('链接已复制到剪贴板');
			}
		} catch (err) {
			if ((err as Error).name !== 'AbortError') {
				console.error('Share failed:', err);
				toast.error('分享失败，请重试');
			}
		}
	}
</script>

<svelte:window bind:scrollY bind:innerHeight={clientHeight} />

{#if showHeader}
	<header
		class="fixed top-0 left-0 right-0 z-40 hidden h-14 border-b border-ink-100/50 bg-white/70 backdrop-blur-xl transition-all duration-500 dark:border-ink-800/40 dark:bg-ink-950/70 md:left-24 md:block"
		in:fly={{ y: -100, duration: 600, easing: cubicOut }}
		out:fade={{ duration: 300 }}
	>
		<!-- Progress Bar -->
		<div
			class="absolute bottom-0 left-0 h-[2px] bg-gradient-to-r from-jade-400 to-jade-600 transition-all duration-150 ease-out dark:from-jade-500 dark:to-jade-700"
			style="width: {progress}%"
		></div>

		<div class="mx-auto flex h-full max-w-[1400px] items-center justify-between px-4 lg:px-8">
			<!-- Left Actions -->
			<div class="flex flex-1 items-center gap-1 md:gap-2">
				<Button
					variant="ghost"
					size="sm"
					class="group !h-9 !w-9 !p-0 text-ink-500 hover:bg-ink-100/50 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-ink-100"
					onclick={goBack}
					title="返回"
				>
					<ArrowLeft size={18} class="transition-transform group-hover:-translate-x-0.5" />
				</Button>
				<div class="hidden h-4 w-px bg-ink-100 sm:block dark:bg-ink-800/60"></div>
				<Button
					variant="ghost"
					size="sm"
					class="group !h-9 !w-9 !p-0 text-ink-500 hover:bg-ink-100/50 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-ink-100"
					href="/"
					title="首页"
				>
					<Home size={18} />
				</Button>
			</div>

			<!-- Center Title -->
			<div class="flex flex-[2] items-center justify-center overflow-hidden px-4 md:flex-[3]">
				<button
					class="group flex max-w-full items-center gap-2 transition-all hover:opacity-80"
					onclick={scrollToTop}
				>
					<span
						class="truncate font-serif text-[13px] font-medium tracking-wide text-ink-900 dark:text-ink-50 md:text-sm"
					>
						{title}
					</span>
					<ArrowUp
						size={12}
						class="shrink-0 text-jade-500 opacity-0 transition-all group-hover:translate-y-[-2px] group-hover:opacity-100"
					/>
				</button>
			</div>

			<!-- Right Actions -->
			<div class="flex flex-1 items-center justify-end gap-1 md:gap-2">
				<div class="hidden items-center gap-1 pr-2 lg:flex">
					<span class="font-mono text-[10px] font-bold text-ink-300 dark:text-ink-600">
						{Math.round(progress)}%
					</span>
				</div>

				{#if showCommentShortcut}
					<Button
						variant="ghost"
						size="sm"
						class="!h-9 !w-9 !p-0 text-ink-500 hover:bg-ink-100/50 hover:text-ink-900 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-ink-100"
						onclick={scrollToComments}
						title="跳转评论"
					>
						<MessageSquare size={18} />
					</Button>
				{/if}

				<Button
					variant="ghost"
					size="sm"
					class="!h-9 !w-9 !p-0 text-ink-500 hover:bg-ink-100/50 hover:text-jade-600 dark:text-ink-400 dark:hover:bg-ink-800/50 dark:hover:text-jade-400"
					onclick={handleShare}
					title="分享文章"
				>
					<Share2 size={18} />
				</Button>

				<div class="h-4 w-px bg-ink-100 dark:bg-ink-800/60"></div>

				<Button
					variant="ghost"
					size="sm"
					class="!h-9 !w-9 !p-0 bg-jade-500/10 text-jade-600 hover:bg-jade-500/20 dark:bg-jade-500/20 dark:text-jade-400 dark:hover:bg-jade-500/30"
					onclick={scrollToTop}
					title="回到顶部"
				>
					<ArrowUp size={18} />
				</Button>
			</div>
		</div>
	</header>
{/if}

<style>
	/* Subtle gradient for progress bar glow */
	header::after {
		content: '';
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 1px;
		background: linear-gradient(90deg, transparent, rgba(16, 185, 129, 0.2), transparent);
		pointer-events: none;
	}
</style>
