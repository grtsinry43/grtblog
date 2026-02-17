import type { ComponentType } from 'svelte';
import Moon from 'lucide-svelte/icons/moon';
import Sun from 'lucide-svelte/icons/sun';
import BookOpen from 'lucide-svelte/icons/book-open';
import Aperture from 'lucide-svelte/icons/aperture';
import Feather from 'lucide-svelte/icons/feather';
import Hash from 'lucide-svelte/icons/hash';
import Archive from 'lucide-svelte/icons/archive';
import Ellipsis from 'lucide-svelte/icons/ellipsis';
import House from 'lucide-svelte/icons/house';
import PenTool from 'lucide-svelte/icons/pen-tool';
import Image from 'lucide-svelte/icons/image';
import User from 'lucide-svelte/icons/user';
import Terminal from 'lucide-svelte/icons/terminal';
import Coffee from 'lucide-svelte/icons/coffee';
import Sparkles from 'lucide-svelte/icons/sparkles';
import Code from 'lucide-svelte/icons/code';
import List from 'lucide-svelte/icons/list';
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
	house: House,
	'book-open': BookOpen,
	aperture: Aperture,
	feather: Feather,
	hash: Hash,
	'pen-tool': PenTool,
	archive: Archive,
	ellipsis: Ellipsis,
	image: Image,
	user: User,
	terminal: Terminal,
	coffee: Coffee,
	sparkles: Sparkles,
	code: Code,
	list: List,
	github: Github,
	mail: Mail,
	rss: Rss
} as const;

export type LucideIconKey = keyof typeof lucideIcons;

export default lucideIcons;
