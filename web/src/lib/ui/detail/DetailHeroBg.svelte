<script lang="ts">
	let { src }: { src: string } = $props();

	const handleLoad = (e: Event) => {
		const img = e.target as HTMLImageElement;
		img.classList.add('hero-bg-loaded');
	};
</script>

{#if src}
	<div class="detail-hero-bg" aria-hidden="true">
		<img
			{src}
			alt=""
			class="hero-bg-img"
			loading="eager"
			onload={handleLoad}
		/>
		<!-- Progressive blur band at bottom for smooth dissolution -->
		<div class="hero-blur-band">
			<div class="hero-blur-layer hero-blur-1"></div>
			<div class="hero-blur-layer hero-blur-2"></div>
			<div class="hero-blur-layer hero-blur-3"></div>
		</div>
	</div>
{/if}

<style>
	.detail-hero-bg {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 55vh;
		overflow: hidden;
		pointer-events: none;
		z-index: 0;
		-webkit-mask-image: linear-gradient(
			to bottom,
			black 0%,
			rgba(0, 0, 0, 0.6) 40%,
			rgba(0, 0, 0, 0.25) 65%,
			rgba(0, 0, 0, 0.05) 85%,
			transparent 100%
		);
		mask-image: linear-gradient(
			to bottom,
			black 0%,
			rgba(0, 0, 0, 0.6) 40%,
			rgba(0, 0, 0, 0.25) 65%,
			rgba(0, 0, 0, 0.05) 85%,
			transparent 100%
		);
	}

	/* Image: starts very blurry + transparent, animates to decorative state on load */
	.hero-bg-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		opacity: 0;
		filter: blur(50px) saturate(1.2);
		transform: scale(1.15);
		transition:
			filter 1.8s cubic-bezier(0.22, 1, 0.36, 1),
			opacity 1.8s cubic-bezier(0.22, 1, 0.36, 1),
			transform 1.8s cubic-bezier(0.22, 1, 0.36, 1);
	}

	.hero-bg-img:global(.hero-bg-loaded) {
		opacity: 0.18;
		filter: blur(20px) saturate(1.3);
		transform: scale(1.06);
	}

	:global(.dark) .hero-bg-img:global(.hero-bg-loaded) {
		opacity: 0.14;
		filter: blur(24px) saturate(1.2);
	}

	/* Progressive blur layers at the bottom for smoother dissolution */
	.hero-blur-band {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 50%;
		pointer-events: none;
	}

	.hero-blur-layer {
		position: absolute;
		inset: 0;
	}

	.hero-blur-1 {
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
		-webkit-mask-image: linear-gradient(to bottom, transparent 0%, black 40%, transparent 70%);
		mask-image: linear-gradient(to bottom, transparent 0%, black 40%, transparent 70%);
	}

	.hero-blur-2 {
		backdrop-filter: blur(12px);
		-webkit-backdrop-filter: blur(12px);
		-webkit-mask-image: linear-gradient(to bottom, transparent 30%, black 65%, transparent 90%);
		mask-image: linear-gradient(to bottom, transparent 30%, black 65%, transparent 90%);
	}

	.hero-blur-3 {
		backdrop-filter: blur(24px);
		-webkit-backdrop-filter: blur(24px);
		background: linear-gradient(
			to bottom,
			transparent 40%,
			rgba(250, 250, 249, 0.3) 70%,
			rgba(250, 250, 249, 0.6) 100%
		);
		-webkit-mask-image: linear-gradient(to bottom, transparent 50%, black 100%);
		mask-image: linear-gradient(to bottom, transparent 50%, black 100%);
	}

	:global(.dark) .hero-blur-3 {
		background: linear-gradient(
			to bottom,
			transparent 40%,
			rgba(10, 10, 10, 0.3) 70%,
			rgba(10, 10, 10, 0.6) 100%
		);
	}
</style>
