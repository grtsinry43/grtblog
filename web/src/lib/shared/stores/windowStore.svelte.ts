export class WindowManager {
	isOpen = $state(false);
	isMinimized = $state(false);
	isExpanded = $state(false);
	title = $state('系统终端');
	kind = $state<string | null>(null);
	position = $state({ x: 100, y: 100 });
	size = $state<{ width: number | null; height: number | null }>({
		width: null,
		height: null
	});
	openVersion = $state(0);
	data = $state<Record<string, unknown> | null>(null);

	constructor() {}

	open(title: string, data: Record<string, unknown> | null = null, kind: string | null = null) {
		this.title = title;
		this.data = data;
		this.kind = kind;
		this.isOpen = true;
		this.isMinimized = false;
		this.isExpanded = this.shouldDefaultExpanded(kind);
		this.size = this.getDefaultSize();
		this.openVersion += 1;
		const width = this.size.width ?? this.getEstimatedWidth();
		const height = this.size.height ?? this.getEstimatedHeight();
		this.centerInViewport(width, height);
	}

	close() {
		this.isOpen = false;
	}

	minimize() {
		this.isMinimized = true;
	}

	restore() {
		this.isMinimized = false;
		const width = this.size.width ?? this.getEstimatedWidth();
		const height = this.size.height ?? this.getEstimatedHeight();
		this.syncToViewport(width, height);
	}

	toggleExpanded() {
		if (typeof window !== 'undefined' && window.innerWidth < 768) {
			return;
		}
		this.isExpanded = !this.isExpanded;
	}

	setSize(width: number, height: number) {
		if (typeof window === 'undefined' || window.innerWidth < 768) {
			return;
		}
		const maxWidth = this.getExpandedWidth();
		const maxHeight = this.getExpandedHeight();
		this.size = {
			width: this.clamp(Math.round(width), 450, maxWidth),
			height: this.clamp(Math.round(height), 300, maxHeight)
		};
	}

	updatePosition(dx: number, dy: number, nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const newX = this.position.x + dx;
		const newY = this.position.y + dy;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(newX, 0, bounds.maxX);
		this.position.y = this.clamp(newY, 0, bounds.maxY);
	}

	centerInViewport(nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(Math.round((window.innerWidth - nodeWidth) / 2), 0, bounds.maxX);
		this.position.y = this.clamp(Math.round((window.innerHeight - nodeHeight) / 2), 0, bounds.maxY);
	}

	syncToViewport(nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(this.position.x, 0, bounds.maxX);
		this.position.y = this.clamp(this.position.y, 0, bounds.maxY);
	}

	private getViewportBounds(nodeWidth: number, nodeHeight: number) {
		return {
			maxX: Math.max(0, window.innerWidth - nodeWidth),
			maxY: Math.max(0, window.innerHeight - nodeHeight)
		};
	}

	private clamp(value: number, min: number, max: number) {
		return Math.max(min, Math.min(value, max));
	}

	private getEstimatedWidth() {
		if (typeof window === 'undefined') return 450;
		return window.innerWidth >= 768 ? 450 : Math.round(window.innerWidth * 0.9);
	}

	private getEstimatedHeight() {
		if (typeof window === 'undefined') return 300;
		const expectedHeight = Math.round(window.innerHeight * 0.6);
		return Math.min(420, Math.max(260, expectedHeight));
	}

	private getExpandedWidth() {
		if (typeof window === 'undefined') return 1100;
		return Math.round(Math.min(window.innerWidth * 0.92, 1100));
	}

	private getExpandedHeight() {
		if (typeof window === 'undefined') return 420;
		return Math.round(window.innerHeight * 0.82);
	}

	private getDefaultSize() {
		if (typeof window === 'undefined' || window.innerWidth < 768) {
			return { width: null, height: null };
		}
		return {
			width: this.getEstimatedWidth(),
			height: this.getEstimatedHeight()
		};
	}

	private shouldDefaultExpanded(kind: string | null) {
		if (typeof window === 'undefined' || window.innerWidth < 768) {
			return false;
		}
		return kind === 'thinking-comments';
	}
}

export const windowStore = new WindowManager();
