import type { RequestHandler } from './$types';

const DEFAULT_INTERNAL_API_BASE_URL = 'http://localhost:8080/api/v2';

const HOP_BY_HOP_HEADERS = new Set([
	'connection',
	'keep-alive',
	'proxy-authenticate',
	'proxy-authorization',
	'te',
	'trailer',
	'transfer-encoding',
	'upgrade'
]);

const FORWARDED_HEADER_NAMES = [
	'accept',
	'accept-language',
	'if-none-match',
	'if-modified-since',
	'user-agent',
	'referer',
	'purpose',
	'sec-purpose',
	'x-purpose',
	'x-moz'
] as const;

function resolveInternalApiBaseURL(): string {
	if (typeof process === 'undefined' || !process.env) {
		return DEFAULT_INTERNAL_API_BASE_URL;
	}
	const raw = (process.env.INTERNAL_API_BASE_URL || '').trim();
	if (!raw) {
		return DEFAULT_INTERNAL_API_BASE_URL;
	}
	if (raw.endsWith('/api/v2')) {
		return raw;
	}
	return `${raw.replace(/\/+$/, '')}/api/v2`;
}

function buildProxyRequestHeaders(event: Parameters<RequestHandler>[0]): Headers {
	const headers = new Headers();

	for (const name of FORWARDED_HEADER_NAMES) {
		const value = event.request.headers.get(name);
		if (value) {
			headers.set(name, value);
		}
	}

	const clientAddress = event.getClientAddress();
	const xRealIP = (event.request.headers.get('x-real-ip') || '').trim() || clientAddress;
	const xForwardedFor =
		(event.request.headers.get('x-forwarded-for') || '').trim() || clientAddress;

	if (xRealIP) {
		headers.set('x-real-ip', xRealIP);
	}
	if (xForwardedFor) {
		headers.set('x-forwarded-for', xForwardedFor);
	}
	headers.set('x-forwarded-proto', event.url.protocol.replace(':', ''));

	return headers;
}

function sanitizeProxyResponseHeaders(input: Headers): Headers {
	const headers = new Headers(input);
	for (const name of HOP_BY_HOP_HEADERS) {
		headers.delete(name);
	}
	return headers;
}

async function proxyFeed(event: Parameters<RequestHandler>[0]): Promise<Response> {
	const baseURL = resolveInternalApiBaseURL().replace(/\/+$/, '');
	const upstreamURL = new URL(`${baseURL}/public/feed`);
	upstreamURL.search = event.url.search;

	const upstreamResponse = await event.fetch(upstreamURL, {
		headers: buildProxyRequestHeaders(event),
		redirect: 'manual'
	});

	const headers = sanitizeProxyResponseHeaders(upstreamResponse.headers);
	if (!headers.has('content-type')) {
		headers.set('content-type', 'application/rss+xml; charset=utf-8');
	}

	return new Response(upstreamResponse.body, {
		status: upstreamResponse.status,
		statusText: upstreamResponse.statusText,
		headers
	});
}

export const trailingSlash = 'never';
export const prerender = false;

export const GET: RequestHandler = async (event) => {
	try {
		return await proxyFeed(event);
	} catch (error) {
		console.error('Failed to proxy /feed request:', error);
		return new Response('Bad Gateway', { status: 502 });
	}
};

export const HEAD: RequestHandler = async (event) => {
	const response = await GET(event);
	return new Response(null, {
		status: response.status,
		statusText: response.statusText,
		headers: response.headers
	});
};
