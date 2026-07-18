type TimelineGestureOptions = {
	enabled?: boolean;
	isActive: () => boolean;
	canConsume: (delta: number) => boolean;
	onDelta: (delta: number) => void;
};

type WheelAxis = 'x' | 'y' | null;

const AXIS_LOCK_THRESHOLD = 8;
const CLICK_SUPPRESS_DISTANCE = 6;
const CLICK_SUPPRESS_MS = 180;
const WHEEL_GESTURE_GAP_MS = 120;
const WHEEL_SIGNAL_THRESHOLD = 6;
const HORIZONTAL_INTENT_RATIO = 0.8;
const DOM_DELTA_LINE = 1;
const DOM_DELTA_PAGE = 2;

const isTouchLikePointer = (event: PointerEvent) =>
	event.pointerType === 'touch' || event.pointerType === 'pen';

export const normalizeWheelDelta = (event: Pick<WheelEvent, 'deltaMode' | 'deltaX' | 'deltaY'>) => {
	const multiplier =
		event.deltaMode === DOM_DELTA_LINE
			? 16
			: event.deltaMode === DOM_DELTA_PAGE
				? typeof window === 'undefined'
					? 800
					: window.innerHeight
				: 1;

	return {
		x: event.deltaX * multiplier,
		y: event.deltaY * multiplier
	};
};

export const resolveWheelAxis = (x: number, y: number, shiftKey = false): WheelAxis => {
	const absX = Math.abs(x);
	const absY = Math.abs(y);
	if (Math.max(absX, absY) < WHEEL_SIGNAL_THRESHOLD) return null;
	if (shiftKey || absX >= absY * HORIZONTAL_INTENT_RATIO) return 'x';
	return 'y';
};

export function timelineGesture(node: HTMLElement, initialOptions: TimelineGestureOptions) {
	let options = initialOptions;
	let pointerActive = false;
	let pointerId: number | null = null;
	let pointerStartX = 0;
	let pointerStartY = 0;
	let lastPointerX = 0;
	let dragDistance = 0;
	let lockedPointerAxis: WheelAxis = null;
	let suppressClickUntil = 0;

	let wheelAxis: WheelAxis = null;
	let wheelX = 0;
	let wheelY = 0;
	let lastWheelAt = 0;
	let pendingDelta = 0;
	let animationFrame = 0;

	const isEnabled = () => options.enabled !== false;

	const flushDelta = () => {
		animationFrame = 0;
		const delta = pendingDelta;
		pendingDelta = 0;
		if (delta !== 0) options.onDelta(delta);
	};

	const queueDelta = (delta: number) => {
		pendingDelta += delta;
		if (!animationFrame) animationFrame = requestAnimationFrame(flushDelta);
	};

	const resetWheelGesture = () => {
		wheelAxis = null;
		wheelX = 0;
		wheelY = 0;
	};

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
		lockedPointerAxis = null;
	};

	const handleWheel = (event: WheelEvent) => {
		if (!isEnabled() || !options.isActive()) return;

		const now = performance.now();
		if (now - lastWheelAt > WHEEL_GESTURE_GAP_MS) resetWheelGesture();
		lastWheelAt = now;

		const normalized = normalizeWheelDelta(event);
		const deltaX = event.shiftKey && Math.abs(normalized.x) < 0.5 ? normalized.y : normalized.x;
		const deltaY = event.shiftKey ? 0 : normalized.y;
		wheelX += deltaX;
		wheelY += deltaY;

		if (!wheelAxis) {
			wheelAxis = resolveWheelAxis(wheelX, wheelY, event.shiftKey);
		}

		if (wheelAxis !== 'x' || deltaX === 0 || !options.canConsume(deltaX)) return;
		if (event.cancelable) event.preventDefault();
		queueDelta(deltaX);
	};

	const handlePointerDown = (event: PointerEvent) => {
		if (!isEnabled() || !isTouchLikePointer(event) || !event.isPrimary) return;
		pointerActive = true;
		pointerId = event.pointerId;
		pointerStartX = lastPointerX = event.clientX;
		pointerStartY = event.clientY;
		dragDistance = 0;
		lockedPointerAxis = null;
		try {
			node.setPointerCapture(event.pointerId);
		} catch {
			// Pointer capture is optional on older touch browsers.
		}
	};

	const handlePointerMove = (event: PointerEvent) => {
		if (!pointerActive || pointerId !== event.pointerId || !isEnabled() || !options.isActive()) {
			return;
		}

		const totalDx = event.clientX - pointerStartX;
		const totalDy = event.clientY - pointerStartY;
		if (!lockedPointerAxis) {
			if (Math.abs(totalDx) < AXIS_LOCK_THRESHOLD && Math.abs(totalDy) < AXIS_LOCK_THRESHOLD) {
				return;
			}
			lockedPointerAxis = Math.abs(totalDx) > Math.abs(totalDy) ? 'x' : 'y';
		}
		if (lockedPointerAxis !== 'x') return;

		const pointerDelta = event.clientX - lastPointerX;
		lastPointerX = event.clientX;
		dragDistance += Math.abs(pointerDelta);
		const timelineDelta = -pointerDelta;
		if (timelineDelta === 0 || !options.canConsume(timelineDelta)) return;
		if (event.cancelable) event.preventDefault();
		queueDelta(timelineDelta);
	};

	node.addEventListener('wheel', handleWheel, { passive: false });
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
			if (animationFrame) cancelAnimationFrame(animationFrame);
			node.removeEventListener('wheel', handleWheel);
			node.removeEventListener('pointerdown', handlePointerDown);
			node.removeEventListener('pointermove', handlePointerMove);
			node.removeEventListener('pointerup', stopPointerGesture);
			node.removeEventListener('pointercancel', stopPointerGesture);
			window.removeEventListener('click', suppressDragEndClick, true);
		}
	};
}
