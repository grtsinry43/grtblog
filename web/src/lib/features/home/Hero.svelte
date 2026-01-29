<script lang="ts">
    import SocialItem from '$lib/features/home/SocialItem.svelte';
	import { ArrowDown } from 'lucide-svelte';
</script>

<div class="hero-container min-h-[100%-4em] flex flex-col w-full gap-20">
	<div class="hero-info flex justify-center mt-[200px] gap-36">
		<div class="hero-author-avatar relative z-10 w-fit">
			<img
				src="https://dogeoss.grtsinry43.com/img/author.jpeg"
				alt="Author"
				class="h-[184px] w-[184px] rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
			/>
		</div>
		<div class="hero-welcome">
			<div class="text-4xl font-mono leading-relaxed">
				<span class="italic">Hi! 👋</span><br />
				I'm grtinry43
			</div>
			<p class="hero-subtitle mt-12 font-mono">
				Java & JavaScript full-stack developer committed to crafting excellent software.
			</p>
		</div>
	</div>
    <div class="hero-motto font-serif text-2xl ml-4">
        热衷于在逻辑与感性的缝隙中构建数字花园。<br/>
        也许，代码是现代的诗歌，而文字是思想的快照。
    </div>

    <div class="social-container ml-4 flex items-center gap-6">
        <!-- Social Items -->
        <SocialItem icon="github" name="GitHub" href="https://github.com/grtinry43" />
        <SocialItem icon="mail" name="Email" href="mailto:grtsinry43@outlook.com" />
        <SocialItem icon="rss" name="RSS" href="/feed" />
    </div>

	<div class="hero-scroll-hint" aria-hidden="true">
		<ArrowDown size={20} />
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.hero-author-avatar::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-5 translate-y-5 rounded-default border border-ink-300 dark:border-ink-900/30;
	}
	.hero-author-avatar:hover::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-2 translate-y-2 rounded-default border border-ink-300 transition dark:border-ink-900/30;
	}

	.hero-container {
		@apply relative;
	}

	.hero-scroll-hint {
		@apply absolute bottom-8 right-10 flex h-12 w-12 items-center justify-center text-ink-500  dark:text-ink-300;
		animation: hero-scroll-bounce 1.6s ease-in-out infinite;
	}

	@media (prefers-reduced-motion: reduce) {
		.hero-scroll-hint {
			animation: none;
		}
	}

	@keyframes hero-scroll-bounce {
		0%, 100% {
			transform: translateY(0);
			opacity: 0.6;
		}
		50% {
			transform: translateY(8px);
			opacity: 1;
		}
	}
</style>
