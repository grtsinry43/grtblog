import QRCode from 'qrcode';

export interface SharePosterContent {
	description: string;
	imageUrl?: string;
	siteIconUrl?: string;
	siteName: string;
	title: string;
	url: string;
}

export type LinkShareResult = 'shared' | 'copied';

const POSTER_WIDTH = 1080;
const POSTER_HEIGHT = 1440;

export function readSharePosterContent(): SharePosterContent {
	const canonicalUrl = document.querySelector<HTMLLinkElement>('link[rel="canonical"]')?.href;
	return {
		description:
			readMeta('meta[property="og:description"]') || readMeta('meta[name="description"]'),
		imageUrl: readMeta('meta[property="og:image"]') || undefined,
		siteIconUrl: document.querySelector<HTMLLinkElement>('link[rel="icon"]')?.href || undefined,
		siteName: readMeta('meta[property="og:site_name"]') || window.location.hostname,
		title: readMeta('meta[property="og:title"]') || document.title,
		url: canonicalUrl || window.location.href
	};
}

export async function sharePageLink(content: SharePosterContent): Promise<LinkShareResult> {
	if (navigator.share) {
		await navigator.share({
			text: content.description || undefined,
			title: content.title,
			url: content.url
		});
		return 'shared';
	}

	await navigator.clipboard.writeText(content.url);
	return 'copied';
}

export async function generateSharePoster(content: SharePosterContent): Promise<Blob> {
	const [, cover, siteIcon, qrCode] = await Promise.all([
		loadPosterFonts(),
		content.imageUrl ? loadImage(content.imageUrl) : Promise.resolve(null),
		content.siteIconUrl ? loadImage(content.siteIconUrl) : Promise.resolve(null),
		createQrCode(content.url)
	]);

	const canvas = document.createElement('canvas');
	canvas.width = POSTER_WIDTH;
	canvas.height = POSTER_HEIGHT;
	const context = canvas.getContext('2d');
	if (!context) throw new Error('Canvas is unavailable');

	drawPaper(context);
	if (cover) drawCoverAccent(context, cover);
	drawEditorialMarks(context, content.siteName, siteIcon);

	const titleSize = content.title.length > 34 ? 58 : content.title.length > 20 ? 66 : 76;
	context.fillStyle = '#1f2925';
	context.font = `700 ${titleSize}px "Noto Serif SC", "Songti SC", serif`;
	const titleBottom = drawWrappedText(context, content.title, 104, 340, 820, titleSize * 1.42, 5);

	context.fillStyle = '#708078';
	context.font = '400 30px "Noto Serif SC", "Songti SC", serif';
	const description = normalizeText(content.description);
	const descriptionTop = Math.max(titleBottom + 68, 670);
	if (description) {
		drawWrappedText(context, description, 108, descriptionTop, 790, 52, 4);
	}

	drawFooter(context, content, qrCode);
	return canvasToBlob(canvas);
}

export async function sharePosterImage(blob: Blob, title: string): Promise<boolean> {
	const file = new File([blob], `${posterFilename(title)}.png`, { type: 'image/png' });
	const data: ShareData = { files: [file], title };
	if (!navigator.share || !navigator.canShare?.(data)) return false;

	await navigator.share(data);
	return true;
}

export function downloadSharePoster(blob: Blob, title: string): void {
	const url = URL.createObjectURL(blob);
	const anchor = document.createElement('a');
	anchor.href = url;
	anchor.download = `${posterFilename(title)}.png`;
	document.body.appendChild(anchor);
	anchor.click();
	anchor.remove();
	setTimeout(() => URL.revokeObjectURL(url), 0);
}

function readMeta(selector: string): string {
	return document.querySelector<HTMLMetaElement>(selector)?.content.trim() ?? '';
}

function drawPaper(context: CanvasRenderingContext2D): void {
	context.fillStyle = '#f5f1e8';
	context.fillRect(0, 0, POSTER_WIDTH, POSTER_HEIGHT);

	context.fillStyle = 'rgba(36, 48, 43, 0.035)';
	for (let index = 0; index < 420; index += 1) {
		const x = (index * 83) % POSTER_WIDTH;
		const y = (index * 149) % POSTER_HEIGHT;
		context.fillRect(x, y, index % 5 === 0 ? 2 : 1, 1);
	}

	const glow = context.createRadialGradient(890, 180, 20, 890, 180, 420);
	glow.addColorStop(0, 'rgba(83, 143, 111, 0.16)');
	glow.addColorStop(1, 'rgba(83, 143, 111, 0)');
	context.fillStyle = glow;
	context.fillRect(480, 0, 600, 620);
}

function drawCoverAccent(context: CanvasRenderingContext2D, image: HTMLImageElement): void {
	context.save();
	context.globalAlpha = 0.18;
	context.beginPath();
	context.arc(930, 165, 235, 0, Math.PI * 2);
	context.clip();
	context.filter = 'grayscale(0.25) saturate(0.75) contrast(0.9)';
	const scale = Math.max(470 / image.naturalWidth, 470 / image.naturalHeight);
	const width = image.naturalWidth * scale;
	const height = image.naturalHeight * scale;
	context.drawImage(image, 930 - width / 2, 165 - height / 2, width, height);
	context.restore();
}

function drawEditorialMarks(
	context: CanvasRenderingContext2D,
	siteName: string,
	siteIcon: HTMLImageElement | null
): void {
	context.fillStyle = '#527b67';
	context.fillRect(104, 112, 58, 6);
	context.fillStyle = '#b95b4d';
	context.fillRect(170, 112, 18, 6);

	let siteNameX = 104;
	if (siteIcon) {
		context.save();
		context.beginPath();
		context.arc(122, 163, 18, 0, Math.PI * 2);
		context.clip();
		drawImageCover(context, siteIcon, 104, 145, 36, 36);
		context.restore();
		siteNameX = 154;
	}

	context.fillStyle = '#526159';
	context.font = '400 24px "Google Sans", sans-serif';
	context.letterSpacing = '2px';
	context.fillText(siteName, siteNameX, 172);
}

function drawFooter(
	context: CanvasRenderingContext2D,
	content: SharePosterContent,
	qrCode: HTMLCanvasElement
): void {
	context.strokeStyle = 'rgba(45, 61, 53, 0.18)';
	context.lineWidth = 2;
	context.beginPath();
	context.moveTo(104, 1212);
	context.lineTo(976, 1212);
	context.stroke();

	context.fillStyle = '#53635b';
	context.letterSpacing = '0px';
	drawFullUrl(context, content.url, 104, 1258, 650);

	context.fillStyle = '#8a958f';
	context.font = '400 18px "Noto Serif SC", serif';
	context.fillText('循着链接，读完这一页', 104, 1388);

	context.drawImage(qrCode, 826, 1230, 144, 144);
	context.fillStyle = '#7f8b85';
	context.font = '400 16px "Noto Serif SC", serif';
	context.textAlign = 'center';
	context.fillText('扫码阅读', 898, 1400);
	context.textAlign = 'start';
}

function drawWrappedText(
	context: CanvasRenderingContext2D,
	text: string,
	x: number,
	y: number,
	maxWidth: number,
	lineHeight: number,
	maxLines: number
): number {
	const characters = Array.from(normalizeText(text));
	const lines: string[] = [];
	let line = '';

	for (const character of characters) {
		const candidate = line + character;
		if (line && context.measureText(candidate).width > maxWidth) {
			lines.push(line.trimEnd());
			line = character.trimStart();
		} else {
			line = candidate;
		}
	}
	if (line) lines.push(line.trimEnd());

	const visibleLines = lines.slice(0, maxLines);
	if (lines.length > maxLines && visibleLines.length) {
		let last = visibleLines[visibleLines.length - 1] ?? '';
		while (last && context.measureText(`${last}…`).width > maxWidth) last = last.slice(0, -1);
		visibleLines[visibleLines.length - 1] = `${last}…`;
	}

	visibleLines.forEach((value, index) => context.fillText(value, x, y + index * lineHeight));
	return y + Math.max(visibleLines.length - 1, 0) * lineHeight;
}

function normalizeText(value: string): string {
	return value.replace(/\s+/g, ' ').trim();
}

function drawFullUrl(
	context: CanvasRenderingContext2D,
	value: string,
	x: number,
	y: number,
	maxWidth: number
): void {
	let fontSize = 21;
	let lines: string[];

	while (true) {
		context.font = `500 ${fontSize}px "Victor Mono Variable", monospace`;
		lines = wrapText(context, value, maxWidth);
		if (lines.length <= 4 || fontSize <= 12) break;
		fontSize -= 1;
	}

	const lineHeight = Math.max(22, fontSize + 8);
	lines.forEach((line, index) => context.fillText(line, x, y + index * lineHeight));
}

function wrapText(context: CanvasRenderingContext2D, value: string, maxWidth: number): string[] {
	const lines: string[] = [];
	let line = '';

	for (const character of Array.from(value)) {
		const candidate = line + character;
		if (line && context.measureText(candidate).width > maxWidth) {
			lines.push(line);
			line = character;
		} else {
			line = candidate;
		}
	}
	if (line) lines.push(line);
	return lines;
}

function posterFilename(title: string): string {
	const normalized = title
		.replace(/[\\/:*?"<>|]/g, '')
		.replace(/\s+/g, '-')
		.slice(0, 48);
	return normalized || 'share-poster';
}

async function loadPosterFonts(): Promise<void> {
	if (!document.fonts) return;

	const loadedFonts = await Promise.all([
		document.fonts.load('700 76px "Noto Serif SC"', '分享卡片'),
		document.fonts.load('400 30px "Noto Serif SC"', '根据本页内容生成'),
		document.fonts.load('400 24px "Google Sans"', 'GRtBlog'),
		document.fonts.load('500 22px "Victor Mono Variable"', 'example.com/article')
	]);
	await document.fonts.ready;

	if (loadedFonts.some((fontFaces) => fontFaces.length === 0)) {
		throw new Error('Poster fonts failed to load');
	}
}

async function createQrCode(value: string): Promise<HTMLCanvasElement> {
	const canvas = document.createElement('canvas');
	await QRCode.toCanvas(canvas, value, {
		color: {
			dark: '#26352eff',
			light: '#fffdf8ff'
		},
		errorCorrectionLevel: 'M',
		margin: 2,
		width: 288
	});
	return canvas;
}

async function loadImage(source: string): Promise<HTMLImageElement | null> {
	return new Promise((resolve) => {
		const image = new Image();
		image.crossOrigin = 'anonymous';
		image.decoding = 'async';
		image.onload = () => resolve(image);
		image.onerror = () => resolve(null);
		image.src = source;
	});
}

function drawImageCover(
	context: CanvasRenderingContext2D,
	image: HTMLImageElement,
	x: number,
	y: number,
	width: number,
	height: number
): void {
	const scale = Math.max(width / image.naturalWidth, height / image.naturalHeight);
	const sourceWidth = width / scale;
	const sourceHeight = height / scale;
	const sourceX = (image.naturalWidth - sourceWidth) / 2;
	const sourceY = (image.naturalHeight - sourceHeight) / 2;
	context.drawImage(image, sourceX, sourceY, sourceWidth, sourceHeight, x, y, width, height);
}

function canvasToBlob(canvas: HTMLCanvasElement): Promise<Blob> {
	return new Promise((resolve, reject) => {
		canvas.toBlob((blob) => {
			if (blob) resolve(blob);
			else reject(new Error('Failed to encode share poster'));
		}, 'image/png');
	});
}
