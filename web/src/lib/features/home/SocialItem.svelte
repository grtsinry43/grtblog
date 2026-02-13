<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';

	const { icon, name, href } = $props<{
		icon: string;
		name: string;
		href: string;
	}>();
</script>

<div class="social-item-container hover:text-jade-600 cursor-pointer">
	<a href={href.startsWith('/') ? resolve(href) : href} class="flex items-center gap-2">
		<DynamicLucideIcon name={icon} size={14} />
		<span class="font-mono hover:underline">{name}</span>
	</a>
</div>
