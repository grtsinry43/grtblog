import { mount, unmount } from 'svelte';
import type { Component } from 'svelte';

type ComponentConstructor = Component<Record<string, any>>;

const registry = new Map<string, ComponentConstructor>();

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

type MountedInstance = {
	el: HTMLElement;
	instance: unknown;
};

export const mountMarkdownComponents = (root: HTMLElement) => {
	const instances: MountedInstance[] = [];
	const mounted = new WeakSet<HTMLElement>();

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
		const placeholders = Array.from(
			container.querySelectorAll<HTMLElement>('.md-component-placeholder')
		).filter((el) => !mounted.has(el) && el.dataset.mounted !== 'true');

		placeholders.sort((a, b) => depthOf(b) - depthOf(a));

		for (const el of placeholders) {
			const name = el.dataset.component?.trim() || '';
			const Component = registry.get(name);
			const props = parseProps(el.dataset.props);
			const contentHtml = el.innerHTML;

			if (!Component) {
				renderUnsupported(el, name, props);
				mounted.add(el);
				el.dataset.mounted = 'true';
				continue;
			}

			el.dataset.contentHtml = contentHtml;
			el.innerHTML = '';
			const instance = mount(Component, { target: el, props: { ...props, contentHtml } });
			instances.push({ el, instance });
			mounted.add(el);
			el.dataset.mounted = 'true';

			mountIn(el);
		}
	};

	mountIn(root);

	return () => {
		for (const { el, instance } of instances) {
			unmount(instance as never);
			const fallback = el.dataset.contentHtml;
			if (typeof fallback === 'string') {
				el.innerHTML = fallback;
			}
			el.dataset.mounted = 'false';
		}
	};
};
