/**
 * Scroll to a heading anchor inside a container,
 * updating the URL hash via `history.replaceState`.
 */
export const scrollToAnchor = (
	container: HTMLElement | null,
	anchor: string,
	event?: MouseEvent,
	behavior: ScrollBehavior = 'smooth'
): void => {
	event?.preventDefault();
	const root = container ?? (typeof document !== 'undefined' ? document : null);
	if (!root) return;
	const target = root.querySelector(`#${CSS.escape(anchor)}`) as HTMLElement | null;
	if (!target) return;

	let collapsedAncestor = target.parentElement?.closest<HTMLDetailsElement>('details') ?? null;
	let expandedDetails = false;
	while (collapsedAncestor) {
		if (!collapsedAncestor.open) {
			collapsedAncestor.open = true;
			expandedDetails = true;
		}
		collapsedAncestor =
			collapsedAncestor.parentElement?.closest<HTMLDetailsElement>('details') ?? null;
	}

	const scroll = () => {
		const offset = 80;
		const top = target.getBoundingClientRect().top + window.scrollY - offset;
		window.scrollTo({ top, behavior });
	};
	if (expandedDetails) requestAnimationFrame(scroll);
	else scroll();
	if (typeof history !== 'undefined') history.replaceState(null, '', `#${anchor}`);
};
