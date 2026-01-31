import { mount, unmount } from 'svelte';
import type { Component } from 'svelte';

type ComponentConstructor = Component<Record<string, any>>;

const registry = new Map<string, ComponentConstructor>();
const mountedInstances = new WeakMap<HTMLElement, unknown>();

const parseProps = (raw?: string) => {
	if (!raw) {
		return {};
	}
	try {
		const parsed = JSON.parse(raw);
		return typeof parsed === 'object' && parsed ? parsed : {};
	} catch {
		return {};
	}
};

const renderUnsupported = (el: HTMLElement, name: string, props: Record<string, any>) => {
	el.innerHTML = `
		<div class="md-component-fallback">
			<span class="md-component-fallback__label">组件暂不支持</span>
		</div>
	`;
};

export const registerMarkdownComponent = (name: string, component: ComponentConstructor) => {
	registry.set(name, component);
};

export const unregisterMarkdownComponent = (name: string) => {
	registry.delete(name);
};

const collectPlaceholders = (root: HTMLElement) => {
	const placeholders = root.matches('.md-component-placeholder')
		? [root]
		: [];
	for (const el of root.querySelectorAll<HTMLElement>('.md-component-placeholder')) {
		placeholders.push(el);
	}
	return placeholders;
};

export const unmountMarkdownComponent = (el: HTMLElement) => {
	const instance = mountedInstances.get(el);
	if (!instance) return;
	unmount(instance as never);
	const fallback = el.dataset.contentHtml;
	if (typeof fallback === 'string') {
		el.innerHTML = fallback;
	}
	el.dataset.mounted = 'false';
	mountedInstances.delete(el);
};

export const unmountMarkdownComponentsIn = (root: HTMLElement) => {
	for (const el of collectPlaceholders(root)) {
		unmountMarkdownComponent(el);
	}
};

export const mountMarkdownComponents = (root: HTMLElement) => {
	const depthOf = (el: HTMLElement) => {
		let depth = 0;
		let current: HTMLElement | null = el.parentElement;
		while (current && current !== root) {
			depth += 1;
			current = current.parentElement;
		}
		return depth;
	};

	const mountIn = (container: HTMLElement) => {
		const placeholders = collectPlaceholders(container).filter(
			(el) => !mountedInstances.has(el) && el.dataset.mounted !== 'true'
		);

		placeholders.sort((a, b) => depthOf(b) - depthOf(a));

		for (const el of placeholders) {
			const name = el.dataset.component?.trim() || '';
			const Component = registry.get(name);
			const props = parseProps(el.dataset.props);
			const contentHtml = el.innerHTML;
			const mountTargetSelector = el.dataset.mountTarget;
			const mountTarget = mountTargetSelector
				? el.querySelector<HTMLElement>(mountTargetSelector)
				: null;

			if (!Component) {
				renderUnsupported(el, name, props);
				el.dataset.mounted = 'true';
				continue;
			}

			el.dataset.contentHtml = contentHtml;
			if (!mountTarget) {
				el.innerHTML = '';
			}
			const imgEl = name === 'md-image' ? el.querySelector('img') : null;
			const instance = mount(Component, {
				target: mountTarget ?? el,
				props: { ...props, contentHtml, imgEl }
			});
			mountedInstances.set(el, instance);
			el.dataset.mounted = 'true';

			mountIn(el);
		}
	};

	mountIn(root);

	return () => unmountMarkdownComponentsIn(root);
};
