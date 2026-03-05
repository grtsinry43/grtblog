import { Resvg } from '@resvg/resvg-js';

const MAX_TITLE_LENGTH = 60;
const MAX_SUBTITLE_LENGTH = 120;
const MAX_SITE_LENGTH = 48;
const MAX_TAG_LENGTH = 24;
const OG_IMAGE_WIDTH = 1200;
const OG_IMAGE_HEIGHT = 630;
const MAX_ICON_BYTES = 2 * 1024 * 1024;
const ICON_FETCH_TIMEOUT_MS = 3000;

const SUPPORTED_ICON_MIME_TYPES = new Set([
	'image/png',
	'image/jpeg',
	'image/webp',
	'image/gif',
	'image/svg+xml',
	'image/avif'
]);

const ICON_FRAME_X = 916;
const ICON_FRAME_Y = 154;
const ICON_FRAME_SIZE = 180;
const ICON_SIZE = 132;
const ICON_INSET = (ICON_FRAME_SIZE - ICON_SIZE) / 2;
const ICON_X = ICON_FRAME_X + ICON_INSET;
const ICON_Y = ICON_FRAME_Y + ICON_INSET;

const normalizeText = (value: string): string => value.replace(/\s+/g, ' ').trim();

const clipText = (value: string, max: number): string => {
	if (value.length <= max) return value;
	return `${value.slice(0, Math.max(0, max - 1)).trimEnd()}…`;
};

const escapeXml = (value: string): string =>
	value
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll("'", '&#39;');

const glyphWidth = (char: string): number => {
	const codePoint = char.codePointAt(0);
	if (!codePoint) return 1;
	if (codePoint <= 0x007f) return 1;
	if (codePoint >= 0x4e00 && codePoint <= 0x9fff) return 2;
	return 1.2;
};

const trimToWidth = (value: string, maxWidth: number): string => {
	if (!value) return '';
	let result = '';
	let width = 0;
	for (const char of value) {
		const w = glyphWidth(char);
		if (width + w > maxWidth) break;
		result += char;
		width += w;
	}
	return result;
};

const wrapText = (value: string, maxWidth: number, maxLines: number): string[] => {
	if (!value) return [];
	const lines: string[] = [];
	let current = '';
	let currentWidth = 0;
	let overflowed = false;

	for (const char of value) {
		const width = glyphWidth(char);
		if (currentWidth + width > maxWidth && current) {
			lines.push(current);
			if (lines.length >= maxLines) {
				overflowed = true;
				current = '';
				break;
			}
			current = char;
			currentWidth = width;
			continue;
		}
		current += char;
		currentWidth += width;
	}

	if (current && lines.length < maxLines) {
		lines.push(current);
	}

	if (overflowed && lines.length > 0) {
		const last = lines[lines.length - 1] ?? '';
		lines[lines.length - 1] = `${trimToWidth(last, Math.max(1, maxWidth - 1))}…`;
	}

	return lines;
};

export type ThemeMode = 'light' | 'dark';

const palette = {
	light: {
		bgStart: '#fafaf9',
		bgEnd: '#f5f5f4',
		panel: '#ffffff',
		border: '#e7e5e4',
		title: '#1c1917',
		subtitle: '#57534e',
		brand: '#0d9488',
		mono: '#78716c',
		noise: 0.04,
		glowA: '#14b8a6',
		glowB: '#f59e0b'
	},
	dark: {
		bgStart: '#1c1917',
		bgEnd: '#0c0a09',
		panel: '#292524',
		border: '#44403c',
		title: '#f5f5f4',
		subtitle: '#d6d3d1',
		brand: '#2dd4bf',
		mono: '#a8a29e',
		noise: 0.07,
		glowA: '#0d9488',
		glowB: '#d97706'
	}
} as const;

const buildSvg = (
	title: string,
	subtitle: string,
	site: string,
	tag: string,
	theme: ThemeMode,
	iconDataUrl: string
): string => {
	const titleLines = wrapText(title, 22, 2);
	const subtitleLines = wrapText(subtitle, 46, 2);
	const color = palette[theme];

	const titleBlocks = titleLines
		.map(
			(line, index) =>
				`<text x="124" y="${228 + index * 82}" font-size="60" font-weight="700" fill="${color.title}" font-family="'Noto Serif SC','Georgia',serif" letter-spacing="1">${escapeXml(line)}</text>`
		)
		.join('');

	const subtitleBlocks = subtitleLines
		.map(
			(line, index) =>
				`<text x="124" y="${458 + index * 42}" font-size="28" font-weight="400" fill="${color.subtitle}" font-family="'Noto Serif SC','Georgia',serif" font-style="italic">${escapeXml(line)}</text>`
		)
		.join('');

	const iconClipDef = iconDataUrl
		? `
    <clipPath id="iconClip">
      <rect x="${ICON_X}" y="${ICON_Y}" width="${ICON_SIZE}" height="${ICON_SIZE}" rx="4"/>
    </clipPath>
`
		: '';

	const iconBlock = iconDataUrl
		? `
  <rect x="${ICON_FRAME_X}" y="${ICON_FRAME_Y}" width="${ICON_FRAME_SIZE}" height="${ICON_FRAME_SIZE}" rx="4" fill="${color.panel}" stroke="${color.border}" stroke-width="1.5"/>
  <image href="${escapeXml(iconDataUrl)}" x="${ICON_X}" y="${ICON_Y}" width="${ICON_SIZE}" height="${ICON_SIZE}" preserveAspectRatio="xMidYMid meet" clip-path="url(#iconClip)"/>
`
		: '';

	return `<?xml version="1.0" encoding="UTF-8"?>
<svg width="${OG_IMAGE_WIDTH}" height="${OG_IMAGE_HEIGHT}" viewBox="0 0 ${OG_IMAGE_WIDTH} ${OG_IMAGE_HEIGHT}" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="${escapeXml(title)}">
  <defs>
    <linearGradient id="bg" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="${color.bgStart}"/>
      <stop offset="100%" stop-color="${color.bgEnd}"/>
    </linearGradient>
    <radialGradient id="jadeGlow" cx="0.78" cy="0.18" r="0.72">
      <stop offset="0%" stop-color="${color.glowA}" stop-opacity="0.22"/>
      <stop offset="100%" stop-color="${color.glowA}" stop-opacity="0"/>
    </radialGradient>
    <radialGradient id="amberGlow" cx="0.14" cy="0.88" r="0.52">
      <stop offset="0%" stop-color="${color.glowB}" stop-opacity="0.14"/>
      <stop offset="100%" stop-color="${color.glowB}" stop-opacity="0"/>
    </radialGradient>
    <filter id="noise" x="-20%" y="-20%" width="140%" height="140%">
      <feTurbulence type="fractalNoise" baseFrequency="0.85" numOctaves="2" seed="7"/>
      <feColorMatrix type="saturate" values="0"/>
      <feComponentTransfer>
        <feFuncA type="table" tableValues="0 ${color.noise}"/>
      </feComponentTransfer>
    </filter>
${iconClipDef}
    <clipPath id="contentClip">
      <rect x="124" y="154" width="760" height="360"/>
    </clipPath>
  </defs>
  <rect width="${OG_IMAGE_WIDTH}" height="${OG_IMAGE_HEIGHT}" fill="url(#bg)"/>
  <rect width="${OG_IMAGE_WIDTH}" height="${OG_IMAGE_HEIGHT}" fill="url(#jadeGlow)"/>
  <rect width="${OG_IMAGE_WIDTH}" height="${OG_IMAGE_HEIGHT}" fill="url(#amberGlow)"/>
  <rect width="${OG_IMAGE_WIDTH}" height="${OG_IMAGE_HEIGHT}" filter="url(#noise)"/>
  <rect x="72" y="68" width="1056" height="494" rx="4" fill="${color.panel}" stroke="${color.border}" stroke-width="1.5"/>
  <line x1="124" y1="94" x2="124" y2="146" stroke="${color.brand}" stroke-opacity="0.55" />
  <text x="144" y="118" font-size="18" fill="${color.brand}" fill-opacity="0.85" letter-spacing="2" font-family="'Google Sans Code',system-ui,sans-serif">${escapeXml(tag)}</text>
  <circle cx="124" cy="534" r="3" fill="${color.brand}" fill-opacity="0.65"/>
  <line x1="138" y1="534" x2="312" y2="534" stroke="${color.brand}" stroke-opacity="0.35" />
  ${iconBlock}
  <g clip-path="url(#contentClip)">
    ${titleBlocks}
    ${subtitleBlocks}
  </g>
  <text x="124" y="542" font-size="22" font-weight="600" fill="${color.brand}" font-family="'Google Sans Code',system-ui,sans-serif">${escapeXml(site)}</text>
	</svg>`;
};

const inferImageType = (source: string): string => {
	const lowered = source.toLowerCase();
	if (lowered.endsWith('.svg')) return 'image/svg+xml';
	if (lowered.endsWith('.jpg') || lowered.endsWith('.jpeg')) return 'image/jpeg';
	if (lowered.endsWith('.webp')) return 'image/webp';
	if (lowered.endsWith('.gif')) return 'image/gif';
	if (lowered.endsWith('.avif')) return 'image/avif';
	return 'image/png';
};

const normalizeContentType = (value: string | null, fallbackType: string): string => {
	const baseType = (value || '').split(';')[0]?.trim()?.toLowerCase() || fallbackType;
	return SUPPORTED_ICON_MIME_TYPES.has(baseType) ? baseType : '';
};

const resolveImageFetchTarget = (source: string, requestUrl: URL): string | null => {
	if (!source) return null;
	if (source.startsWith('data:image/')) return source;
	try {
		const target = new URL(source, requestUrl.origin);
		if (!['http:', 'https:'].includes(target.protocol)) return null;
		if (target.origin === requestUrl.origin) {
			return `${target.pathname}${target.search}`;
		}
		return target.toString();
	} catch {
		return null;
	}
};

const fetchImageAsDataUrl = async (
	fetcher: typeof fetch,
	source: string,
	requestUrl: URL
): Promise<string> => {
	const normalized = normalizeText(source);
	if (!normalized) return '';
	if (normalized.startsWith('data:image/')) return normalized;

	const target = resolveImageFetchTarget(normalized, requestUrl);
	if (!target) return '';

	const controller = new AbortController();
	const timeout = setTimeout(() => controller.abort(), ICON_FETCH_TIMEOUT_MS);
	try {
		const response = await fetcher(target, {
			method: 'GET',
			headers: {
				accept: 'image/*'
			},
			signal: controller.signal
		});
		if (!response.ok) return '';

		const contentType = normalizeContentType(
			response.headers.get('content-type'),
			inferImageType(normalized)
		);
		if (!contentType) return '';

		const imageBytes = new Uint8Array(await response.arrayBuffer());
		if (imageBytes.byteLength === 0 || imageBytes.byteLength > MAX_ICON_BYTES) return '';

		return `data:${contentType};base64,${Buffer.from(imageBytes).toString('base64')}`;
	} catch {
		return '';
	} finally {
		clearTimeout(timeout);
	}
};

export type OgImageInput = {
	title: string;
	subtitle: string;
	site: string;
	tag: string;
	theme?: ThemeMode;
	iconUrl?: string;
	fallbackIconUrl?: string;
};

export async function renderOgImage(
	input: OgImageInput,
	fetcher: typeof fetch,
	requestUrl: URL
): Promise<ArrayBuffer> {
	const title = clipText(normalizeText(input.title || 'grtBlog'), MAX_TITLE_LENGTH);
	const subtitle = clipText(
		normalizeText(input.subtitle || 'A personal blog about software and life.'),
		MAX_SUBTITLE_LENGTH
	);
	const site = clipText(normalizeText(input.site || 'grtBlog'), MAX_SITE_LENGTH);
	const tag = clipText(normalizeText(input.tag || 'PREVIEW'), MAX_TAG_LENGTH);
	const theme: ThemeMode = input.theme === 'dark' ? 'dark' : 'light';

	let iconDataUrl = '';
	if (input.iconUrl) {
		iconDataUrl = await fetchImageAsDataUrl(fetcher, input.iconUrl, requestUrl);
	}
	if (!iconDataUrl && input.fallbackIconUrl) {
		iconDataUrl = await fetchImageAsDataUrl(fetcher, input.fallbackIconUrl, requestUrl);
	}

	const svg = buildSvg(title, subtitle, site, tag, theme, iconDataUrl);
	const resvg = new Resvg(svg, {
		fitTo: { mode: 'width', value: OG_IMAGE_WIDTH },
		font: { loadSystemFonts: true, defaultFontFamily: 'Noto Serif SC' }
	});
	const png = resvg.render().asPng();
	return png.buffer.slice(png.byteOffset, png.byteOffset + png.byteLength) as ArrayBuffer;
}
