import { writable } from 'svelte/store';

/** Stores the hero background image URL for detail pages (article cover / moment image). */
export const detailHeroBgSrc = writable<string>('');
