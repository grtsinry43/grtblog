<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';

	const { icon, name, href } = $props<{
		icon: string;
		name: string;
		href: string;
	}>();

	const shouldDisablePreloadData = (value: string): boolean => {
		if (!value.startsWith('/')) return false;
		const path = value.split(/[?#]/, 1)[0];
		return path === '/feed' || path === '/rss.xml';
	};
</script>

<div class="social-item-container hover:text-jade-600 cursor-pointer">
	<a
		href={href.startsWith('/') ? resolve(href) : href}
		data-sveltekit-preload-data={shouldDisablePreloadData(href) ? 'off' : undefined}
		class="flex items-center gap-2"
		target="_blank"
		rel="noopener noreferrer"
	>
		<DynamicLucideIcon name={icon} size={14} />
		<span class="font-mono hover:underline">{name}</span>
	</a>
</div>
