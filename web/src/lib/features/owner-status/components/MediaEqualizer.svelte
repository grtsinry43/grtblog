<script lang="ts">
	let { playing = true }: { playing?: boolean } = $props();
</script>

<span
	class="media-equalizer inline-flex items-end gap-[2px]"
	class:is-paused={!playing}
	aria-hidden="true"
>
	<span class="eq-bar" style="--eq-delay: -0.15s; --eq-duration: 0.9s"></span>
	<span class="eq-bar" style="--eq-delay: -0.62s; --eq-duration: 1.1s"></span>
	<span class="eq-bar" style="--eq-delay: -0.34s; --eq-duration: 0.7s"></span>
	<span class="eq-bar" style="--eq-delay: -0.81s; --eq-duration: 1.25s"></span>
</span>

<style lang="postcss">
	@reference "$routes/layout.css";

	.media-equalizer {
		height: 14px;
	}

	.eq-bar {
		width: 3px;
		height: 100%;
		border-radius: 1px;
		transform: scaleY(0.3);
		transform-origin: bottom;
		@apply bg-jade-500 dark:bg-jade-400;
		animation: eq-bounce var(--eq-duration, 1s) ease-in-out var(--eq-delay, 0s) infinite;
	}

	.media-equalizer.is-paused .eq-bar {
		animation: none;
		transform: scaleY(0.3);
		@apply bg-ink-300 dark:bg-ink-600;
	}

	@keyframes eq-bounce {
		0%,
		100% {
			transform: scaleY(0.3);
		}
		50% {
			transform: scaleY(1);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.eq-bar {
			animation: none;
			transform: scaleY(0.6);
		}
	}
</style>
