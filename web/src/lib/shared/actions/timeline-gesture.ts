type TimelineGestureOptions = {
	enabled?: boolean;
	isActive: () => boolean;
	onDeltaX: (deltaX: number) => void;
};

const AXIS_LOCK_THRESHOLD = 8;
const CLICK_SUPPRESS_DISTANCE = 6;
const CLICK_SUPPRESS_MS = 180;

const isTouchLikePointer = (event: PointerEvent) =>
	event.pointerType === 'touch' || event.pointerType === 'pen';

export function timelineGesture(node: HTMLElement, initialOptions: TimelineGestureOptions) {
	let options = initialOptions;
	let pointerActive = false;
	let pointerId: number | null = null;
	let pointerStartX = 0;
	let pointerStartY = 0;
	let lastPointerX = 0;
	let dragDistance = 0;
	let lockedAxis: 'x' | 'y' | null = null;
	let suppressClickUntil = 0;

	const isEnabled = () => options.enabled !== false;

	const suppressDragEndClick = (event: MouseEvent) => {
		if (Date.now() > suppressClickUntil) return;
		event.preventDefault();
		event.stopPropagation();
		event.stopImmediatePropagation();
	};

	const stopPointerGesture = (event?: PointerEvent) => {
		if (event && pointerId !== event.pointerId) return;
		if (dragDistance > CLICK_SUPPRESS_DISTANCE) {
			suppressClickUntil = Date.now() + CLICK_SUPPRESS_MS;
		}
		if (pointerId != null && node.hasPointerCapture(pointerId)) {
			node.releasePointerCapture(pointerId);
		}
		pointerActive = false;
		pointerId = null;
		dragDistance = 0;
		lockedAxis = null;
	};

	const handleWheel = (event: WheelEvent) => {
		if (!isEnabled() || !options.isActive()) return;
		const absDeltaX = Math.abs(event.deltaX);
		const absDeltaY = Math.abs(event.deltaY);
		const hasHorizontalSignal =
			(event.shiftKey && absDeltaY > 0.5) || (absDeltaX > 0.5 && absDeltaX >= absDeltaY);
		if (!hasHorizontalSignal) return;

		const delta = event.shiftKey && absDeltaX <= 0.5 ? event.deltaY : event.deltaX;
		if (delta === 0) return;
		event.preventDefault();
		options.onDeltaX(delta);
	};

	const handlePointerDown = (event: PointerEvent) => {
		if (!isEnabled() || !isTouchLikePointer(event) || !event.isPrimary) return;
		pointerActive = true;
		pointerId = event.pointerId;
		pointerStartX = lastPointerX = event.clientX;
		pointerStartY = event.clientY;
		dragDistance = 0;
		lockedAxis = null;
		try {
			node.setPointerCapture(event.pointerId);
		} catch {
			// Ignore platforms that do not support pointer capture for this target.
		}
	};

	const handlePointerMove = (event: PointerEvent) => {
		if (!pointerActive || pointerId !== event.pointerId || !isEnabled() || !options.isActive()) {
			return;
		}

		const totalDx = event.clientX - pointerStartX;
		const totalDy = event.clientY - pointerStartY;
		if (!lockedAxis) {
			if (Math.abs(totalDx) < AXIS_LOCK_THRESHOLD && Math.abs(totalDy) < AXIS_LOCK_THRESHOLD) {
				return;
			}
			lockedAxis = Math.abs(totalDx) > Math.abs(totalDy) ? 'x' : 'y';
		}
		if (lockedAxis !== 'x') return;

		const deltaX = event.clientX - lastPointerX;
		lastPointerX = event.clientX;
		dragDistance += Math.abs(deltaX);
		if (deltaX === 0) return;
		event.preventDefault();
		options.onDeltaX(deltaX);
	};

	window.addEventListener('wheel', handleWheel, { passive: false });
	node.addEventListener('pointerdown', handlePointerDown);
	node.addEventListener('pointermove', handlePointerMove, { passive: false });
	node.addEventListener('pointerup', stopPointerGesture);
	node.addEventListener('pointercancel', stopPointerGesture);
	window.addEventListener('click', suppressDragEndClick, true);

	return {
		update(nextOptions: TimelineGestureOptions) {
			options = nextOptions;
		},
		destroy() {
			stopPointerGesture();
			window.removeEventListener('wheel', handleWheel);
			node.removeEventListener('pointerdown', handlePointerDown);
			node.removeEventListener('pointermove', handlePointerMove);
			node.removeEventListener('pointerup', stopPointerGesture);
			node.removeEventListener('pointercancel', stopPointerGesture);
			window.removeEventListener('click', suppressDragEndClick, true);
		}
	};
}
