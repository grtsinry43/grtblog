/**
 * Smooth-scroll to a heading anchor inside a container,
 * updating the URL hash via `history.replaceState`.
 */
export const scrollToAnchor = (
	container: HTMLElement | null,
	anchor: string,
	event?: MouseEvent
): void => {
	event?.preventDefault();
	const root = container ?? (typeof document !== 'undefined' ? document : null);
	if (!root) return;
	const target = root.querySelector(`#${CSS.escape(anchor)}`) as HTMLElement | null;
	if (!target) return;
	const offset = 80;
	const top = target.getBoundingClientRect().top + window.scrollY - offset;
	window.scrollTo({ top, behavior: 'smooth' });
	if (typeof history !== 'undefined') history.replaceState(null, '', `#${anchor}`);
};
