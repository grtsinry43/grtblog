import { browser } from '$app/environment';

export const scrollToElementById = (
	id: string,
	options: ScrollIntoViewOptions = { behavior: 'smooth', block: 'center' }
): HTMLElement | null => {
	if (!browser) return null;
	const element = document.getElementById(id);
	element?.scrollIntoView(options);
	return element;
};
