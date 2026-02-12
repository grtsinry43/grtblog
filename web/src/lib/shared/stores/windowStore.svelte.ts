export class WindowManager {
	isOpen = $state(false);
	isMinimized = $state(false);
	title = $state('系统终端');
	position = $state({ x: 100, y: 100 });
	openVersion = $state(0);

	constructor() {}

	open(title: string) {
		this.title = title;
		this.isOpen = true;
		this.isMinimized = false;
		this.openVersion += 1;
		this.centerInViewport(this.getEstimatedWidth(), this.getEstimatedHeight());
	}

	close() {
		this.isOpen = false;
	}

	minimize() {
		this.isMinimized = true;
	}

	restore() {
		this.isMinimized = false;
		this.syncToViewport(450, 300);
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
}

export const windowStore = new WindowManager();
