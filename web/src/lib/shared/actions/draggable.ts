type DraggableOptions = {
	handle?: string;
	onMove: (dx: number, dy: number) => void;
};

const NON_DRAG_SELECTOR =
	'button, a, input, textarea, select, option, [role="button"], [contenteditable="true"], [data-no-drag]';

export function draggable(node: HTMLElement, initialOptions: DraggableOptions) {
	let options = initialOptions;
	let pointerX = 0;
	let pointerY = 0;
	let dragging = false;
	let dragDistance = 0;
	let suppressClickUntil = 0;
	let handleEl: HTMLElement = node;

	const getHandleElement = () => {
		if (!options.handle) return node;
		return (node.querySelector(options.handle) as HTMLElement) ?? node;
	};

	const cleanupDragListeners = () => {
		window.removeEventListener('pointermove', handlePointerMove);
		window.removeEventListener('pointerup', stopDragging);
		window.removeEventListener('pointercancel', stopDragging);
		document.body.style.userSelect = '';
	};

	function handlePointerMove(event: PointerEvent) {
		if (!dragging) return;
		const dx = event.clientX - pointerX;
		const dy = event.clientY - pointerY;
		dragDistance += Math.abs(dx) + Math.abs(dy);
		pointerX = event.clientX;
		pointerY = event.clientY;
		options.onMove(dx, dy);
	}

	function suppressDragEndClick(event: MouseEvent) {
		if (Date.now() > suppressClickUntil) return;
		event.preventDefault();
		event.stopPropagation();
		event.stopImmediatePropagation();
	}

	function stopDragging() {
		if (!dragging) return;
		dragging = false;
		if (dragDistance > 3) {
			suppressClickUntil = Date.now() + 120;
		}
		handleEl.style.cursor = 'grab';
		cleanupDragListeners();
	}

	function handlePointerDown(event: PointerEvent) {
		if (event.button !== 0) return;

		const target = event.target as HTMLElement | null;
		if (target?.closest(NON_DRAG_SELECTOR)) {
			return;
		}

		dragging = true;
		dragDistance = 0;
		pointerX = event.clientX;
		pointerY = event.clientY;
		handleEl.style.cursor = 'grabbing';
		document.body.style.userSelect = 'none';
		window.addEventListener('pointermove', handlePointerMove);
		window.addEventListener('pointerup', stopDragging);
		window.addEventListener('pointercancel', stopDragging);
		event.preventDefault();
	}

	const bindHandle = () => {
		handleEl = getHandleElement();
		handleEl.style.cursor = 'grab';
		handleEl.style.touchAction = 'none';
		handleEl.addEventListener('pointerdown', handlePointerDown);
		window.addEventListener('click', suppressDragEndClick, true);
	};

	const unbindHandle = () => {
		handleEl.removeEventListener('pointerdown', handlePointerDown);
		window.removeEventListener('click', suppressDragEndClick, true);
		handleEl.style.cursor = '';
		handleEl.style.touchAction = '';
	};

	bindHandle();

	return {
		update(nextOptions: DraggableOptions) {
			const prevHandle = options.handle;
			options = nextOptions;
			if (prevHandle !== options.handle) {
				unbindHandle();
				bindHandle();
			}
		},
		destroy() {
			stopDragging();
			unbindHandle();
		}
	};
}
