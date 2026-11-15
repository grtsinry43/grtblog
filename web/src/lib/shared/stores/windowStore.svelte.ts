/** Minimum gap between the window and the viewport edges. */
const EDGE = 8;
const MIN_WIDTH = 360;
const MIN_HEIGHT = 220;

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
	/** Position before maximizing, so "restore" returns the window where the user left it. */
	private preExpandPosition: { x: number; y: number } | null = null;

	constructor() {}

	open(title: string, data: Record<string, unknown> | null = null, kind: string | null = null) {
		this.title = title;
		this.data = data;
		this.kind = kind;
		this.isOpen = true;
		this.isMinimized = false;
		this.isExpanded = this.shouldDefaultExpanded(kind);
		this.preExpandPosition = null;
		this.size = this.getDefaultSize();
		this.openVersion += 1;
		const width = this.isExpanded
			? this.getExpandedWidth()
			: (this.size.width ?? this.getEstimatedWidth());
		const height = this.isExpanded ? this.getExpandedHeight() : this.getEstimatedHeight();
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
		if (!this.isExpanded) {
			this.preExpandPosition = { ...this.position };
			this.isExpanded = true;
			this.centerInViewport(this.getExpandedWidth(), this.getExpandedHeight());
		} else {
			this.isExpanded = false;
			if (this.preExpandPosition) this.position = { ...this.preExpandPosition };
			this.preExpandPosition = null;
		}
	}

	/** User resize. Height `null` means "fit content" (up to the CSS max-height). */
	setSize(width: number, height: number | null) {
		if (typeof window === 'undefined' || window.innerWidth < 768) {
			return;
		}
		const maxWidth = Math.min(this.getExpandedWidth(), window.innerWidth - this.position.x - EDGE);
		const maxHeight = window.innerHeight - this.position.y - EDGE;
		this.size = {
			width: this.clamp(Math.round(width), MIN_WIDTH, Math.max(MIN_WIDTH, maxWidth)),
			height:
				height === null
					? null
					: this.clamp(Math.round(height), MIN_HEIGHT, Math.max(MIN_HEIGHT, maxHeight))
		};
	}

	updatePosition(dx: number, dy: number, nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const newX = this.position.x + dx;
		const newY = this.position.y + dy;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(newX, EDGE, bounds.maxX);
		this.position.y = this.clamp(newY, EDGE, bounds.maxY);
	}

	centerInViewport(nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(
			Math.round((window.innerWidth - nodeWidth) / 2),
			EDGE,
			bounds.maxX
		);
		this.position.y = this.clamp(
			Math.round((window.innerHeight - nodeHeight) / 2),
			EDGE,
			bounds.maxY
		);
	}

	syncToViewport(nodeWidth: number, nodeHeight: number) {
		if (typeof window === 'undefined') return;
		const bounds = this.getViewportBounds(nodeWidth, nodeHeight);
		this.position.x = this.clamp(this.position.x, EDGE, bounds.maxX);
		this.position.y = this.clamp(this.position.y, EDGE, bounds.maxY);
	}

	private getViewportBounds(nodeWidth: number, nodeHeight: number) {
		return {
			maxX: Math.max(EDGE, window.innerWidth - nodeWidth - EDGE),
			maxY: Math.max(EDGE, window.innerHeight - nodeHeight - EDGE)
		};
	}

	private clamp(value: number, min: number, max: number) {
		return Math.max(min, Math.min(value, max));
	}

	private getEstimatedWidth() {
		if (typeof window === 'undefined') return 450;
		return window.innerWidth >= 768 ? 480 : Math.round(window.innerWidth * 0.9);
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
		// Height starts content-driven; a fixed height only comes from an explicit user resize.
		return { width: this.getEstimatedWidth(), height: null };
	}

	private shouldDefaultExpanded(kind: string | null) {
		if (typeof window === 'undefined' || window.innerWidth < 768) {
			return false;
		}
		return kind === 'thinking-comments';
	}
}

export const windowStore = new WindowManager();
