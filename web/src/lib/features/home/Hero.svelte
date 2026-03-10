<script lang="ts">
	import SocialItem from '$lib/features/home/SocialItem.svelte';
	import { FadeIn } from '$lib/ui/animation';
	import { ArrowDown } from 'lucide-svelte';
	import type { HomeHeroSocialLink, HomeHeroTemplateNode, HomeHeroThemeConfig } from './types';

	let { config }: { config?: HomeHeroThemeConfig } = $props();

	const defaultTitleTemplate: HomeHeroTemplateNode[] = [
		{ type: 'h1', text: 'Hi! 👋', variant: 'hero_h1_highlight' },
		{ type: 'br' },
		{ type: 'h1', text: "I'm grtsinry43", variant: 'hero_h1_primary' }
	];
	const defaultDescription =
		'Java & JavaScript full-stack developer committed to crafting excellent software.';
	const defaultAvatarUrl = '';
	const defaultMottoLines = [
		'热衷于在逻辑与感性的缝隙中构建数字花园。',
		'也许，代码是现代的诗歌，而文字是思想的快照。'
	];
	const defaultSocials: HomeHeroSocialLink[] = [
		{ icon: 'github', name: 'GitHub', href: 'https://github.com/grtinry43' },
		{ icon: 'mail', name: 'Email', href: 'mailto:grtsinry43@outlook.com' },
		{ icon: 'rss', name: 'RSS', href: '/feed' }
	];

	const variantClassMap: Record<string, string> = {
		hero_h1_highlight: 'italic text-jade-600 dark:text-jade-400 font-light text-4xl',
		hero_h1_primary: 'text-ink-900 dark:text-ink-100 font-medium text-4xl',
		hero_h1_light: 'font-light text-4xl text-ink-900 dark:text-ink-100',
		hero_h1_medium_gap: 'font-medium mx-2 text-4xl text-ink-900 dark:text-ink-100',
		hero_code_inline:
			'font-medium mx-2 text-3xl rounded p-1 bg-gray-200 dark:bg-gray-800/0 hover:dark:bg-gray-800/100 bg-opacity-0 hover:bg-opacity-100 transition-colors duration-200',
		hero_cursor:
			'inline-block w-[1px] h-8 -bottom-2 relative bg-gray-800/80 dark:bg-gray-200/80 opacity-0 group-hover:opacity-100 transition-opacity duration-200 animate-[hero-blink_1s_steps(1)_infinite]'
	};

	const titleTemplate = $derived(
		config?.titleTemplate && config.titleTemplate.length > 0
			? config.titleTemplate
			: defaultTitleTemplate
	);
	const description = $derived(config?.description || defaultDescription);
	const avatarUrl = $derived(config?.avatarUrl || defaultAvatarUrl);
	const mottoLines = $derived(
		config?.mottoLines && config.mottoLines.length > 0 ? config.mottoLines : defaultMottoLines
	);
	const mottoLinesAlign = $derived(config?.mottoLinesAlign ?? 'default');
	const socials = $derived(
		config?.socials && config.socials.length > 0 ? config.socials : defaultSocials
	);
	const socialsAlign = $derived(config?.socialsAlign ?? 'default');

	function resolveNodeClass(node: HomeHeroTemplateNode): string {
		const baseClass = node.type === 'h1' ? 'text-ink-900 dark:text-ink-100' : '';
		const variantClass = node.variant ? (variantClassMap[node.variant] ?? '') : '';
		const customClass = node.className ?? '';
		return `${baseClass} ${variantClass} ${customClass}`.trim();
	}
</script>

<div
	class="hero-container min-h-[calc(100svh-5rem)] md:min-h-[calc(100svh-8rem)] flex flex-col justify-center w-full"
>
	<!-- [Desktop Version] -->
	<div class="hidden md:flex flex-col gap-20">
		<div class="hero-info flex justify-center gap-36">
			{#if avatarUrl}
				<FadeIn y={24} duration={1000}>
					<div class="hero-author-avatar relative z-10 w-fit">
						<img
							src={avatarUrl}
							alt="Author"
							width="184"
							height="184"
							fetchpriority="high"
							class="h-46 w-46 rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
						/>
					</div>
				</FadeIn>
			{/if}
			<FadeIn y={20} duration={1000} delay={200}>
				<div class="hero-welcome group">
					<div class="hero-title-desktop font-mono leading-relaxed">
						{#each titleTemplate as node, idx (`${node.type}-${node.text ?? ''}-${idx}`)}
							{#if node.type === 'br'}
								<br />
								<div class="mt-2 md:mt-4"></div>
							{:else if node.type === 'code'}
								<code class={resolveNodeClass(node)}>{node.text ?? ''}</code>
							{:else if node.type === 'span'}
								<span class={resolveNodeClass(node)}>{node.text ?? ''}</span>
							{:else}
								<h1 class={resolveNodeClass(node)}>{node.text ?? ''}</h1>
							{/if}
						{/each}
					</div>
					<p class="hero-subtitle mt-12 font-mono text-ink-500">
						{description}
					</p>
				</div>
			</FadeIn>
		</div>

		<div class="flex flex-col gap-12 ml-4">
			<FadeIn y={16} duration={900} delay={400}>
				<div
					class="hero-motto font-serif text-2xl leading-relaxed text-ink-800 dark:text-ink-200"
					class:text-center={mottoLinesAlign === 'center'}
				>
					{#each mottoLines as line, lineIdx (`${line}-${lineIdx}`)}
						{line}<br />
					{/each}
				</div>
			</FadeIn>

			<FadeIn y={12} duration={800} delay={600}>
				<div
					class="social-container flex items-center gap-6"
					class:justify-center={socialsAlign === 'center'}
				>
					{#each socials as social, socialIdx (`${social.icon}-${social.href}-${socialIdx}`)}
						<SocialItem icon={social.icon} name={social.name} href={social.href} />
					{/each}
				</div>
			</FadeIn>
		</div>
	</div>

	<!-- [Mobile Version] -->
	<div class="flex md:hidden flex-col items-center pt-8">
		{#if avatarUrl}
			<FadeIn y={15} duration={1000}>
				<div class="relative mb-10">
					<div
						class="absolute inset-0 translate-x-2 translate-y-2 border border-ink-200 dark:border-ink-800 rounded-default -z-10"
					></div>
					<img
						src={avatarUrl}
						alt="Author"
						width="110"
						height="110"
						fetchpriority="high"
						class="h-[110px] w-[110px] rounded-default object-cover ring-1 ring-ink-100 dark:ring-ink-800 shadow-sm"
					/>
				</div>
			</FadeIn>
		{/if}

		<FadeIn y={10} duration={1000} delay={200}>
			<div class="text-center px-6 group">
				<div class="hero-title-mobile font-mono tracking-tight leading-relaxed">
					{#each titleTemplate as node, idx (`mobile-${node.type}-${node.text ?? ''}-${idx}`)}
						{#if node.type === 'br'}
							<br />
						{:else if node.type === 'code'}
							<code class={resolveNodeClass(node)}>{node.text ?? ''}</code>
						{:else if node.type === 'span'}
							<span class={resolveNodeClass(node)}>{node.text ?? ''}</span>
						{:else}
							<h1 class={resolveNodeClass(node)}>{node.text ?? ''}</h1>
						{/if}
					{/each}
				</div>
				<p class="text-[11px] font-mono text-ink-500 leading-relaxed mt-4">
					{description}
				</p>
			</div>
		</FadeIn>

		<FadeIn y={8} duration={800} delay={400} class="mt-12">
			<div class="flex items-center gap-5">
				{#each socials as social, socialIdx (`mobile-${social.icon}-${social.href}-${socialIdx}`)}
					<SocialItem icon={social.icon} name="" href={social.href} />
					{#if socialIdx < socials.length - 1}
						<span class="w-px h-3 bg-ink-200 dark:bg-ink-800"></span>
					{/if}
				{/each}
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

	@keyframes hero-blink {
		0%,
		49% {
			opacity: 0;
		}
		50%,
		100% {
			opacity: 1;
		}
	}

	:global(.hero-title-desktop h1) {
		display: inline;
	}

	:global(.hero-title-mobile h1) {
		display: inline;
		font-size: 1.25rem;
		font-weight: 700;
	}

	:global(.hero-title-mobile code) {
		font-size: 1rem;
	}

	/* 特别为移动端 SocialItem 去掉文字 */
	:global(.md\:hidden .social-container span) {
		display: none;
	}
</style>
