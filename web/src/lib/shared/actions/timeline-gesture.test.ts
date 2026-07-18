import { describe, expect, it } from 'vitest';
import { normalizeWheelDelta, resolveWheelAxis } from './timeline-gesture';

describe('timeline gesture helpers', () => {
	it('normalizes pixel and line wheel deltas', () => {
		expect(normalizeWheelDelta({ deltaMode: 0, deltaX: 12, deltaY: -4 })).toEqual({
			x: 12,
			y: -4
		});
		expect(normalizeWheelDelta({ deltaMode: 1, deltaX: 2, deltaY: 3 })).toEqual({
			x: 32,
			y: 48
		});
	});

	it('accepts diagonal trackpad movement with a strong horizontal signal', () => {
		expect(resolveWheelAxis(8, 10)).toBe('x');
		expect(resolveWheelAxis(1, 10)).toBe('y');
		expect(resolveWheelAxis(0, 1)).toBeNull();
		expect(resolveWheelAxis(0, 8, true)).toBe('x');
	});
});
