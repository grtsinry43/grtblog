const pad2 = (value: number): string => String(value).padStart(2, '0');

const parseDate = (value: string | Date): Date | null => {
	const date = value instanceof Date ? value : new Date(value);
	if (Number.isNaN(date.getTime())) return null;
	return date;
};

const parseDateParts = (
	value: string | Date
): { year: string; month: string; day: string } | null => {
	if (typeof value === 'string') {
		const matched = value.match(/^(\d{4})-(\d{2})-(\d{2})/);
		if (matched) {
			const [, year, month, day] = matched;
			return { year, month, day };
		}
	}

	const date = parseDate(value);
	if (!date) return null;
	return {
		year: String(date.getUTCFullYear()),
		month: pad2(date.getUTCMonth() + 1),
		day: pad2(date.getUTCDate())
	};
};

export const buildPostPath = (slug: string): string => `/posts/${encodeURIComponent(slug)}`;

export const buildPagePath = (slug: string): string => `/${encodeURIComponent(slug)}`;

export const buildMomentPath = (slug: string, createdAt: string | Date): string => {
	const encodedSlug = encodeURIComponent(slug);
	const dateParts = parseDateParts(createdAt);
	if (!dateParts) return `/moments/0000/00/00/${encodedSlug}`;

	const { year, month, day } = dateParts;
	return `/moments/${year}/${month}/${day}/${encodedSlug}`;
};
