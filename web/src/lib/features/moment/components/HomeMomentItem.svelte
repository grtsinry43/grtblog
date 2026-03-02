<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import type { MomentSummary } from '$lib/features/moment/types';
	import { ArrowRight } from 'lucide-svelte';
	import { formatRelativeTime } from '$lib/shared/utils/date';
	import { buildMomentPath } from '$lib/shared/utils/content-path';

	let { moment } = $props<{ moment: MomentSummary }>();

	let mouseX = $state('50%');
	let mouseY = $state('50%');
	let isHovered = $state(false);

	const handleMouseMove = (event: MouseEvent) => {
		const currentTarget = event.currentTarget as HTMLElement | null;
		if (!currentTarget) return;
		const rect = currentTarget.getBoundingClientRect();
		mouseX = `${event.clientX - rect.left}px`;
		mouseY = `${event.clientY - rect.top}px`;
		isHovered = true;
	};

	const handleMouseLeave = () => {
		isHovered = false;
	};
</script>

<a
	href={resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt))}
	class="home-item-card moment-item group w-full px-4 py-4 outline-none focus-visible:ring-2 focus-visible:ring-cinnabar-500/30"
	onmousemove={handleMouseMove}
	onmouseleave={handleMouseLeave}
	data-hovered={isHovered}
	style={`--mouse-x: ${mouseX}; --mouse-y: ${mouseY};`}
>
	<div class="home-item-content flex items-center justify-between gap-3">
		<h3
			class="home-item-title line-clamp-1 font-serif text-[15px] font-medium text-ink-900 dark:text-ink-100"
		>
			<span class="title-underline">{moment.title}</span>
		</h3>

		<div
			class="flex shrink-0 items-center gap-2 text-[11px] font-mono text-ink-400 dark:text-ink-500"
		>
			<span>{formatRelativeTime(moment.createdAt)}</span>
			<span class="home-item-arrow">
				<ArrowRight size={14} strokeWidth={1.7} />
			</span>
		</div>
	</div>
</a>

<style lang="postcss">
	@reference "$routes/layout.css";

	.home-item-card {
		position: relative;
		overflow: hidden;
		isolation: isolate;
		border: 1px solid transparent;
		border-radius: var(--radius-default);
		transition:
			background-color 220ms ease,
			border-color 220ms ease,
			box-shadow 260ms cubic-bezier(0.16, 1, 0.3, 1),
			transform 260ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	.home-item-card::before {
		content: '';
		position: absolute;
		inset: 0;
		opacity: 0;
		pointer-events: none;
		background: radial-gradient(
			240px circle at var(--mouse-x, 50%) var(--mouse-y, 50%),
			rgba(239, 68, 68, 0.14),
			transparent 72%
		);
		transition: opacity 240ms ease;
	}

	.home-item-card::after {
		content: '';
		position: absolute;
		right: 0;
		top: 50%;
		width: 2px;
		height: 0;
		transform: translateY(-50%);
		border-radius: 1px;
		background: linear-gradient(
			180deg,
			transparent,
			rgba(239, 68, 68, 0.6),
			rgba(239, 68, 68, 0.85),
			rgba(239, 68, 68, 0.6),
			transparent
		);
		transition: height 260ms ease;
	}

	.home-item-content {
		position: relative;
		z-index: 1;
	}

	.home-item-title {
		transition: color 220ms ease;
	}

	.home-item-arrow {
		opacity: 0.55;
		color: rgb(120 113 108 / 0.9);
		transition:
			transform 220ms ease,
			opacity 220ms ease,
			color 220ms ease;
	}

	.title-underline {
		position: relative;
		display: inline-block;
	}

	.title-underline::after {
		content: '';
		position: absolute;
		left: 0;
		bottom: -2px;
		width: 100%;
		height: 1px;
		transform: scaleX(0);
		transform-origin: bottom right;
		background-color: rgb(120 113 108 / 0.4);
		transition: transform 260ms ease;
	}

	.home-item-card:hover,
	.home-item-card[data-hovered='true'] {
		background: rgb(250 250 249 / 0.68);
		border-color: rgb(41 37 36 / 0.08);
		box-shadow:
			0 2px 8px rgb(28 25 23 / 0.05),
			0 1px 3px rgb(28 25 23 / 0.1);
		backdrop-filter: blur(22px) saturate(132%);
		transform: translateY(-1px);
	}

	.home-item-card:hover::before,
	.home-item-card[data-hovered='true']::before {
		opacity: 1;
	}

	.home-item-card:hover::after,
	.home-item-card[data-hovered='true']::after {
		height: 24px;
	}

	.home-item-card:hover .home-item-title,
	.home-item-card[data-hovered='true'] .home-item-title {
		color: rgb(220 38 38 / 1);
	}

	.home-item-card:hover .home-item-arrow,
	.home-item-card[data-hovered='true'] .home-item-arrow {
		transform: translateX(4px);
		opacity: 0.8;
		color: rgb(220 38 38 / 1);
	}

	.home-item-card:hover .title-underline::after,
	.home-item-card[data-hovered='true'] .title-underline::after {
		transform: scaleX(1);
		transform-origin: bottom left;
	}

	:global(.dark) .home-item-card::before {
		background: radial-gradient(
			240px circle at var(--mouse-x, 50%) var(--mouse-y, 50%),
			rgba(248, 113, 113, 0.2),
			transparent 72%
		);
	}

	:global(.dark) .home-item-card:hover,
	:global(.dark) .home-item-card[data-hovered='true'] {
		background: rgb(41 37 36 / 0.55);
		border-color: rgb(231 229 228 / 0.12);
		box-shadow: 0 10px 30px -12px rgb(0 0 0 / 0.45);
	}

	:global(.dark) .home-item-arrow {
		color: rgb(168 162 158 / 0.95);
	}

	:global(.dark) .home-item-card:hover .home-item-title,
	:global(.dark) .home-item-card[data-hovered='true'] .home-item-title {
		color: rgb(248 113 113 / 0.95);
	}

	:global(.dark) .home-item-card:hover .home-item-arrow,
	:global(.dark) .home-item-card[data-hovered='true'] .home-item-arrow {
		color: rgb(248 113 113 / 0.95);
	}
</style>
