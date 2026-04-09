<script lang="ts">
	import { X, ArrowRight } from 'lucide-svelte';
	import { fly, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import { onMount } from 'svelte';

	let { isOpen = $bindable(false) } = $props<{ isOpen?: boolean }>();

	const STORAGE_KEY = 'grtblog_timeline_intro_shown';

	onMount(() => {
		// Only auto-open if explicitly requested by redirect AND not shown before
		const isRedirect =
			new URLSearchParams(window.location.search).get('redirect_from') === 'statistics';
		const hasShown = localStorage.getItem(STORAGE_KEY);

		if (isRedirect && !hasShown) {
			isOpen = true;
		}
	});

	function handleClose() {
		isOpen = false;
		localStorage.setItem(STORAGE_KEY, 'true');
	}
</script>

{#if isOpen}
	<!-- Backdrop -->
	<button
		type="button"
		class="fixed inset-0 z-[100] bg-ink-950/20 dark:bg-black/40 backdrop-blur-sm"
		transition:fade={{ duration: 300 }}
		aria-label="关闭提示"
		onclick={handleClose}
	></button>

	<!-- Modal -->
	<div
		class="fixed left-1/2 top-1/2 z-[101] w-[calc(100%-2rem)] max-w-sm -translate-x-1/2 -translate-y-1/2"
		transition:fly={{ y: 20, duration: 500, easing: cubicOut }}
	>
		<div
			class="overflow-hidden rounded-default border border-ink-200/60 bg-white shadow-deep dark:border-ink-800 dark:bg-ink-900 noise-surface"
		>
			<div class="relative p-8 text-center">
				<button
					onclick={handleClose}
					class="absolute right-4 top-4 rounded-full p-2 text-ink-400 hover:bg-ink-100 dark:hover:bg-ink-800 transition-colors"
				>
					<X size={18} />
				</button>

				<div class="mb-10 flex justify-center">
					<div class="evolution-container relative flex h-24 w-24 items-center justify-center">
						<!-- Morphing Blobs -->
						<div class="blob blob-1"></div>
						<div class="blob blob-2"></div>
						<div class="blob blob-3"></div>

						<!-- Transitioning Data Lines -->
						<div class="data-lines">
							<!-- Horizontal Axis Line (Static guide) -->
							<div class="timeline-axis"></div>

							<div class="line l1"></div>
							<div class="line l2"></div>
							<div class="line l3"></div>

							<!-- Pulse Nodes -->
							<div class="pulse-node n1"></div>
							<div class="pulse-node n2"></div>
						</div>

						<!-- Core Light -->
						<div class="absolute h-4 w-4 rounded-full bg-white blur-sm"></div>
					</div>
				</div>

				<h3
					class="mb-3 font-serif text-2xl font-bold tracking-tight text-ink-900 dark:text-ink-100"
				>
					数据，升维进化
				</h3>

				<p class="mb-8 text-sm leading-relaxed text-ink-500 dark:text-ink-400">
					原本的统计页面现已进化为全新的<b>交互式叙事时间轴</b
					>。从离散的数据点到连续的创作流，邀你一同回溯这段充满逻辑与感性的旅程。
				</p>

				<button
					onclick={handleClose}
					class="flex w-full items-center justify-center gap-2 rounded-default bg-ink-900 px-6 py-3.5 text-sm font-medium text-white transition-all hover:bg-jade-600 dark:bg-jade-600 dark:hover:bg-jade-500 shadow-jade-glow"
				>
					<span>开启时空回溯</span>
					<ArrowRight size={16} />
				</button>
			</div>
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.evolution-container {
		filter: drop-shadow(0 0 15px rgba(20, 184, 166, 0.2));
	}

	.blob {
		position: absolute;
		width: 100%;
		height: 100%;
		mix-blend-mode: multiply;
		transition: all 1s ease-in-out;
	}

	:global(.dark) .blob {
		mix-blend-mode: screen;
	}

	.blob-1 {
		background: rgba(20, 184, 166, 0.4);
		border-radius: 60% 40% 30% 70% / 60% 30% 70% 40%;
		animation: morph 8s ease-in-out infinite;
	}

	.blob-2 {
		background: rgba(245, 158, 11, 0.3);
		border-radius: 30% 60% 70% 40% / 50% 60% 30% 60%;
		animation: morph 10s ease-in-out infinite reverse;
	}

	.blob-3 {
		background: rgba(20, 184, 166, 0.2);
		border-radius: 50% 50% 20% 80% / 25% 80% 20% 75%;
		animation: morph 12s ease-in-out infinite;
	}

	@keyframes morph {
		0%,
		100% {
			border-radius: 60% 40% 30% 70% / 60% 30% 70% 40%;
			transform: rotate(0deg) scale(1);
		}
		33% {
			border-radius: 30% 60% 70% 40% / 50% 60% 30% 60%;
			transform: rotate(120deg) scale(1.1);
		}
		66% {
			border-radius: 50% 50% 20% 80% / 25% 80% 20% 75%;
			transform: rotate(240deg) scale(0.9);
		}
	}

	.data-lines {
		position: relative;
		z-index: 2;
		display: flex;
		gap: 8px;
		align-items: center;
		justify-content: center;
		width: 60px;
		height: 40px;
	}

	.timeline-axis {
		position: absolute;
		left: -10px;
		right: -10px;
		height: 0.5px;
		background: rgba(255, 255, 255, 0.4);
		top: 50%;
		transform: translateY(-50%);
	}

	.pulse-node {
		position: absolute;
		width: 4px;
		height: 4px;
		background: white;
		border-radius: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		box-shadow: 0 0 8px white;
		animation: node-ping 4s ease-in-out infinite;
	}

	.n1 {
		left: 15%;
		animation-delay: 0.5s;
	}
	.n2 {
		left: 85%;
		animation-delay: 1.5s;
	}

	@keyframes node-ping {
		0%,
		100% {
			transform: translate(-50%, -50%) scale(1);
			opacity: 0.5;
		}
		50% {
			transform: translate(-50%, -50%) scale(1.5);
			opacity: 1;
		}
	}

	.line {
		width: 2.5px;
		background: white;
		border-radius: 1px;
		opacity: 0.8;
		animation: line-evolve 4s ease-in-out infinite;
		position: relative;
		z-index: 3;
	}

	.l1 {
		height: 10px;
		animation-delay: 0s;
	}
	.l2 {
		height: 20px;
		animation-delay: 0.2s;
	}
	.l3 {
		height: 14px;
		animation-delay: 0.4s;
	}

	@keyframes line-evolve {
		0%,
		100% {
			height: 12px;
			transform: translateY(0) scaleY(1);
			opacity: 0.4;
		}
		45%,
		55% {
			height: 1px;
			transform: translateY(0) scaleX(12);
			opacity: 1;
		}
	}
</style>
