import type { ComponentType } from 'svelte';
import Moon from 'lucide-svelte/icons/moon';
import Sun from 'lucide-svelte/icons/sun';
import BookOpen from 'lucide-svelte/icons/book-open';
import Aperture from 'lucide-svelte/icons/aperture';
import Feather from 'lucide-svelte/icons/feather';
import Hash from 'lucide-svelte/icons/hash';
import Archive from 'lucide-svelte/icons/archive';
import Ellipsis from 'lucide-svelte/icons/ellipsis';
import { Github, Mail, Rss } from 'lucide-svelte';

export type LucideIconComponent = ComponentType<{
	size?: number;
	strokeWidth?: number;
	class?: string;
}>;

// Manual whitelist for tree-shaking in SSR/client bundles.
const lucideIcons = {
	moon: Moon,
	sun: Sun,
	'book-open': BookOpen,
	aperture: Aperture,
	feather: Feather,
	hash: Hash,
	archive: Archive,
	ellipsis: Ellipsis,
	github: Github,
	mail: Mail,
	rss: Rss
} as const;

export type LucideIconKey = keyof typeof lucideIcons;

export default lucideIcons;
