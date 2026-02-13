<script lang="ts">
	import SocialItem from '$lib/features/home/SocialItem.svelte';
	import { FadeIn } from '$lib/ui/animation';
	import { ArrowDown } from 'lucide-svelte';
</script>

<div
	class="hero-container min-h-[calc(100svh-5rem)] md:min-h-[calc(100svh-8rem)] flex flex-col justify-center w-full"
>
	<!-- [Desktop Version] - Original Style -->
	<div class="hidden md:flex flex-col gap-20">
		<div class="hero-info flex justify-center gap-36">
			<FadeIn y={24} duration={1000}>
				<div class="hero-author-avatar relative z-10 w-fit">
					<img
						src="https://dogeoss.grtsinry43.com/img/author.jpeg"
						alt="Author"
						class="h-[184px] w-[184px] rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
					/>
				</div>
			</FadeIn>
			<FadeIn y={20} duration={1000} delay={200}>
				<div class="hero-welcome">
					<div class="text-4xl font-mono leading-relaxed">
						<span class="italic text-jade-600 dark:text-jade-400">Hi! 👋</span><br />
						<span class="text-ink-900 dark:text-ink-100">I'm grtsinry43</span>
					</div>
					<p class="hero-subtitle mt-12 font-mono text-ink-500">
						Java & JavaScript full-stack developer committed to crafting excellent software.
					</p>
				</div>
			</FadeIn>
		</div>

		<div class="flex flex-col gap-12 ml-4">
			<FadeIn y={16} duration={900} delay={400}>
				<div class="hero-motto font-serif text-2xl leading-relaxed text-ink-800 dark:text-ink-200">
					热衷于在逻辑与感性的缝隙中构建数字花园。<br />
					也许，代码是现代的诗歌，而文字是思想的快照。
				</div>
			</FadeIn>

			<FadeIn y={12} duration={800} delay={600}>
				<div class="social-container flex items-center gap-6">
					<SocialItem icon="github" name="GitHub" href="https://github.com/grtinry43" />
					<SocialItem icon="mail" name="Email" href="mailto:grtsinry43@outlook.com" />
					<SocialItem icon="rss" name="RSS" href="/feed" />
				</div>
			</FadeIn>
		</div>
	</div>

	<!-- [Mobile Version] - Compact & Refined Style -->
	<div class="flex md:hidden flex-col items-center pt-8">
		<FadeIn y={15} duration={1000}>
			<div class="relative mb-10">
				<div
					class="absolute inset-0 translate-x-2 translate-y-2 border border-ink-200 dark:border-ink-800 rounded-default -z-10"
				></div>
				<img
					src="https://dogeoss.grtsinry43.com/img/author.jpeg"
					alt="Author"
					class="h-[110px] w-[110px] rounded-default object-cover ring-1 ring-ink-100 dark:ring-ink-800 shadow-sm"
				/>
			</div>
		</FadeIn>

		<FadeIn y={10} duration={1000} delay={200}>
			<div class="text-center px-6">
				<h1 class="text-xl font-mono tracking-tight text-ink-900 dark:text-ink-100 font-bold">
					Hi, I'm grtsinry43
				</h1>
				<p class="text-[11px] font-mono text-ink-500 leading-relaxed mt-4">
					Java & JavaScript full-stack developer committed to crafting excellent software.
				</p>
			</div>
		</FadeIn>

		<FadeIn y={8} duration={800} delay={400} class="mt-12">
			<div class="flex items-center gap-5">
				<SocialItem icon="github" name="" href="https://github.com/grtinry43" />
				<span class="w-px h-3 bg-ink-200 dark:bg-ink-800"></span>
				<SocialItem icon="mail" name="" href="mailto:grtsinry43@outlook.com" />
				<span class="w-px h-3 bg-ink-200 dark:bg-ink-800"></span>
				<SocialItem icon="rss" name="" href="/feed" />
			</div>
		</FadeIn>
	</div>

	<div class="hero-scroll-hint hidden md:flex" aria-hidden="true">
		<ArrowDown size={20} />
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.hero-container {
		@apply relative;
	}

	.hero-scroll-hint {
		@apply absolute right-10 bottom-8 flex h-12 w-12 items-center justify-center text-ink-400 opacity-40;
		animation: hero-scroll-bounce 1.6s ease-in-out infinite;
	}

	@keyframes hero-scroll-bounce {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(8px);
		}
	}

	/* 特别为移动端 SocialItem 去掉文字 */
	:global(.md\:hidden .social-container span) {
		display: none;
	}
</style>
