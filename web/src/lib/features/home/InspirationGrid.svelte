<script lang="ts">
	import { SlideIn, FadeIn, StaggerList } from '$lib/ui/animation';
	import { 
		Quote, 
		Code2, 
		Gamepad2, 
		Coffee, 
		Library, 
		Zap, 
		Sparkles,
		Github
	} from 'lucide-svelte';

	// Mock data for inspiration tiles
	const status = {
		learning: 'Svelte 5 & Runes',
		listening: 'Experience - Ludovico Einaudi',
		reading: 'The Design of Everyday Things',
		coding: 'grtblog-v2'
	};

	const stats = [
		{ label: 'Words', value: '124k', icon: Library, color: 'text-jade-500' },
		{ label: 'Commits', value: '856', icon: Github, color: 'text-ink-900 dark:text-ink-100' },
		{ label: 'Coffee', value: '∞', icon: Coffee, color: 'text-amber-500' }
	];
</script>

<section class="mt-20 md:mt-32">
	<SlideIn direction="up">
		<div class="flex items-center gap-3 mb-10 border-b border-ink-100 dark:border-ink-800 pb-4">
			<span class="h-px w-8 bg-jade-500/40"></span>
			<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">灵感与实验场</h2>
		</div>
	</SlideIn>

	<StaggerList staggerDelay={80} y={20} class="grid grid-cols-1 md:grid-cols-4 lg:grid-cols-6 gap-4 auto-rows-[120px]">
		<!-- Quote Tile (Large) -->
		<div class="col-span-1 md:col-span-3 lg:col-span-3 row-span-2 bento-card p-8 flex flex-col justify-center relative overflow-hidden group">
			<Quote class="absolute -top-4 -left-4 w-24 h-24 text-ink-100 dark:text-ink-800/50 -rotate-12 transition-transform group-hover:rotate-0 duration-700" />
			<div class="relative z-10">
				<p class="text-2xl md:text-3xl font-serif italic text-ink-800 dark:text-ink-100 leading-relaxed">
					“The best way to predict the future is to invent it.”
				</p>
				<p class="mt-4 font-mono text-sm text-ink-400">— Alan Kay</p>
			</div>
			<div class="absolute bottom-0 right-0 w-32 h-32 bg-jade-500/5 blur-[80px] rounded-full"></div>
		</div>

		<!-- Status Tile (Now) -->
		<div class="col-span-1 md:col-span-1 lg:col-span-2 row-span-2 bento-card p-6 flex flex-col gap-6">
			<div class="flex items-center gap-2 text-xs font-mono uppercase tracking-wider text-jade-600 dark:text-jade-400">
				<span class="relative flex h-2 w-2">
					<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-jade-400 opacity-75"></span>
					<span class="relative inline-flex rounded-full h-2 w-2 bg-jade-500"></span>
				</span>
				Now / 正在
			</div>
			
			<div class="space-y-4">
				<div class="flex items-start gap-3">
					<Code2 size={16} class="mt-0.5 text-ink-400" />
					<div>
						<div class="text-[10px] font-mono text-ink-400 uppercase">Coding</div>
						<div class="text-sm font-medium">{status.coding}</div>
					</div>
				</div>
				<div class="flex items-start gap-3">
					<Library size={16} class="mt-0.5 text-ink-400" />
					<div>
						<div class="text-[10px] font-mono text-ink-400 uppercase">Reading</div>
						<div class="text-sm font-medium">{status.reading}</div>
					</div>
				</div>
				<div class="flex items-start gap-3">
					<Zap size={16} class="mt-0.5 text-ink-400" />
					<div>
						<div class="text-[10px] font-mono text-ink-400 uppercase">Learning</div>
						<div class="text-sm font-medium">{status.learning}</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Energy/Mood Tile (Small Square) -->
		<div class="col-span-1 md:col-span-1 lg:col-span-1 row-span-1 bento-card flex items-center justify-center group">
			<div class="text-center">
				<Sparkles size={24} class="mx-auto mb-2 text-amber-400 transition-transform group-hover:scale-125 duration-500" />
				<div class="text-xs font-mono">High Energy</div>
			</div>
		</div>

		<!-- Stats Grid (Iterative) -->
		{#each stats as stat, i}
			<div class="col-span-1 md:col-span-1 lg:col-span-1 row-span-1 bento-card p-4 flex flex-col justify-between hover:border-jade-200 dark:hover:border-jade-900/50 transition-colors">
				<stat.icon size={18} class="text-ink-400" />
				<div>
					<div class="text-2xl font-serif {stat.color}">{stat.value}</div>
					<div class="text-[10px] font-mono uppercase tracking-tighter text-ink-400">{stat.label}</div>
				</div>
			</div>
		{/each}

		<!-- Tech Stack / Tooling (Wide) -->
		<div class="col-span-1 md:col-span-2 lg:col-span-3 row-span-1 bento-card px-6 flex items-center justify-between overflow-hidden relative">
			<div class="flex flex-col">
				<div class="text-[10px] font-mono text-ink-400 uppercase mb-1">Tech Stack</div>
				<div class="flex gap-4">
					<span class="text-sm font-medium hover:text-jade-500 cursor-default transition-colors">Java</span>
					<span class="text-sm font-medium hover:text-jade-500 cursor-default transition-colors">TypeScript</span>
					<span class="text-sm font-medium hover:text-jade-500 cursor-default transition-colors">Svelte</span>
					<span class="text-sm font-medium hover:text-jade-500 cursor-default transition-colors">Rust</span>
				</div>
			</div>
			<div class="flex gap-2">
				<div class="w-8 h-8 rounded-full bg-ink-100 dark:bg-ink-800 flex items-center justify-center transition-colors hover:bg-jade-100 dark:hover:bg-jade-900/30">
					<Code2 size={14} />
				</div>
				<div class="w-8 h-8 rounded-full bg-ink-100 dark:bg-ink-800 flex items-center justify-center transition-colors hover:bg-jade-100 dark:hover:bg-jade-900/30">
					<Gamepad2 size={14} />
				</div>
			</div>
		</div>
	</StaggerList>
</section>

<style lang="postcss">
	@reference "$routes/layout.css";

	.bento-card {
		@apply bg-ink-50 dark:bg-ink-900/50 border border-ink-200/80 dark:border-ink-800 rounded-default transition-all duration-500 hover:shadow-glass dark:hover:shadow-glass-dark relative z-0;
	}

	.bento-card::after {
		content: '';
		@apply absolute inset-0 -z-10 opacity-20 pointer-events-none rounded-default;
		background-image: var(--texture-noise);
		mix-blend-mode: soft-light;
	}

	:root[class~="dark"] .bento-card::after {
		@apply opacity-10;
		filter: invert(1);
	}

	.bento-card:hover {
		@apply -translate-y-1 border-ink-200 dark:border-ink-700;
	}
</style>
