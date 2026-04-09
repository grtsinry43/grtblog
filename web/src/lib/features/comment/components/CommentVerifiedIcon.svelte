<script lang="ts">
	import { BadgeCheck } from 'lucide-svelte';
	import { Tooltip } from '$lib/ui/primitives';

	type VerifiedType = 'owner' | 'friend' | 'author';

	let { type, content } = $props<{
		type: VerifiedType;
		content: string;
	}>();

	const colors = {
		owner: 'text-blue-500 dark:text-blue-400',
		friend: 'text-jade-500 dark:text-jade-400',
		author: 'text-purple-500 dark:text-purple-400'
	} satisfies Record<VerifiedType, string>;
</script>

<Tooltip {content}>
	<span class="inline-flex items-center justify-center hover:opacity-80 transition-opacity">
		<BadgeCheck
			size={15}
			fill="currentColor"
			class={`${colors[type as VerifiedType]} verified-icon`}
		/>
		<span class="sr-only">{content}</span>
	</span>
</Tooltip>

<style>
	:global(.verified-icon) {
		/* The stroke defines the checkmark. White makes it look like a cutout on the solid color. */
		stroke: white;
		stroke-width: 2px;
	}

	:global(.dark .verified-icon) {
		/* Use a very dark color for the checkmark in dark mode for better contrast. */
		stroke: #121212;
	}
</style>
