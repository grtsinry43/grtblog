<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
	import {
		Github,
		Mail,
		MessageSquare,
		History,
		User,
		Info,
		Archive,
		Link as LinkIcon,
		Rss,
		BarChart3,
		Activity
	} from 'lucide-svelte';

	// Mock 数据，后续可以接入后台或 Store
	const currentYear = new Date().getFullYear();
	const onlineCount = 3;

	const footerSections = [
		{
			title: '想要了解我',
			links: [
				{ name: '关于我', href: '/about', icon: User },
				{ name: '本站历史', href: '/history', icon: History },
				{ name: '关于此项目', href: '/project', icon: Info }
			]
		},
		{
			title: '你也许在找',
			links: [
				{ name: '归档', href: '/posts', icon: Archive },
				{ name: '友链', href: '/links', icon: LinkIcon },
				{ name: 'RSS', href: '/feed', icon: Rss },
				{ name: '统计', href: '/stats', icon: BarChart3 },
				{ name: '监控', href: 'https://status.grtsinry43.com', icon: Activity }
			]
		},
		{
			title: '联系我叭',
			links: [
				{ name: '写留言', href: '/moments', icon: MessageSquare },
				{ name: '发邮件', href: 'mailto:grtsinry43@outlook.com', icon: Mail },
				{ name: 'GitHub', href: 'https://github.com/grtsinry43', icon: Github }
			]
		}
	];
</script>

<footer
	class="mt-32 border-t border-jade-100/80 dark:border-ink-800 bg-jade-50/30 dark:bg-ink-950/30 backdrop-blur-sm"
>
	<div class="max-w-[1200px] mx-auto px-6 py-12 md:py-16">
		<!-- Mobile Compact Layout (Hidden on Desktop) -->
		<div class="flex flex-col gap-4 mb-12 md:hidden">
			{#each footerSections as section (section.title)}
				<div class="flex flex-col gap-2">
					<div
						class="text-sm font-serif font-bold text-ink-900 dark:text-ink-100 flex items-center justify-between"
					>
						{section.title}
						<span class="text-ink-300 dark:text-ink-700 font-mono font-normal">></span>
					</div>
					<div class="flex flex-wrap gap-x-4 gap-y-2">
						{#each section.links as link (link.name)}
							<a
								href={/^(https?:|mailto:)/i.test(link.href) ? link.href : resolve(link.href)}
								class="text-sm text-ink-500 hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
							>
								{link.name}
							</a>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<!-- Desktop Multi-column Layout (Hidden on Mobile) -->
		<div class="hidden md:grid grid-cols-4 gap-12 mb-16">
			{#each footerSections as section (section.title)}
				<div class="flex flex-col gap-6">
					<h3
						class="text-sm font-serif font-bold text-ink-900 dark:text-ink-100 flex items-center gap-2"
					>
						<span class="w-1 h-3 bg-jade-500 rounded-full"></span>
						{section.title}
					</h3>
					<ul class="flex flex-col gap-3">
						{#each section.links as link (link.name)}
							<li>
								<a
									href={/^(https?:|mailto:)/i.test(link.href) ? link.href : resolve(link.href)}
									class="text-sm text-ink-500 hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
								>
									{link.name}
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}

			<!-- Brand Info inside Desktop Grid -->
			<div class="flex flex-col gap-6 items-end text-right">
				<div class="flex flex-col items-end">
					<div class="text-xl font-mono font-bold text-ink-900 dark:text-ink-100">
						Grtblog<span class="text-jade-500">.</span>
					</div>
					<p class="text-[11px] font-mono text-ink-400 mt-1 uppercase tracking-wider">
						A blog framework for developers
					</p>
				</div>
				<div
					class="flex items-center gap-2 px-2.5 py-1 rounded-full bg-jade-500/5 border border-jade-500/10 w-fit"
				>
					<span class="relative flex h-1.5 w-1.5">
						<span
							class="animate-ping absolute inline-flex h-full w-full rounded-full bg-jade-400 opacity-75"
						></span>
						<span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-jade-500"></span>
					</span>
					<span class="text-[10px] font-mono text-jade-700/80 dark:text-jade-400/80">
						正在有 {onlineCount} 位小伙伴看着我的网站呐
					</span>
				</div>
			</div>
		</div>

		<!-- Bottom Copyright (Universal) -->
		<div
			class="flex flex-col md:flex-row justify-between items-center pt-8 border-t border-ink-100 dark:border-ink-800/50 gap-4"
		>
			<div class="text-[10px] md:text-[11px] font-mono text-ink-400 text-center md:text-left">
				<p>Copyright © 2022 - {currentYear} grtsinry43. All rights reserved.</p>
				<div class="flex flex-wrap justify-center md:justify-start gap-x-3 mt-1">
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
					<a
						href="https://beian.miit.gov.cn/"
						target="_blank"
						rel="noreferrer"
						class="hover:text-jade-600 transition-colors">湘ICP备2023033970号-1</a
					>
					<span class="hidden md:inline text-ink-200 dark:text-ink-800">|</span>
					<span class="hidden md:inline">Powered by Svelte 5</span>
				</div>
			</div>

			<div class="hidden md:flex items-center gap-4 text-[11px] font-mono text-ink-300">
				<span>Designed with ❤️</span>
			</div>
		</div>
	</div>
</footer>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
