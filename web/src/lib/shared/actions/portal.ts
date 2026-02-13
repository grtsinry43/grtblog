/**
 * Svelte action: 将元素移动到指定容器（默认 document.body），
 * 使其脱离当前组件树的 CSS stacking context，保证 position:fixed 相对视口定位。
 */
export function portal(node: HTMLElement, target: HTMLElement | null = null): { destroy(): void } {
	const dest = target ?? (typeof document !== 'undefined' ? document.body : null);
	if (!dest) return { destroy() {} };

	dest.appendChild(node);

	return {
		destroy() {
			if (node.parentNode === dest) dest.removeChild(node);
		}
	};
}
