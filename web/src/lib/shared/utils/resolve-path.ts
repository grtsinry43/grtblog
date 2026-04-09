import { resolve } from '$app/paths';
import type { PathnameWithSearchOrHash, ResolvedPathname } from '$app/types';

export const isPathnameWithSearchOrHash = (value: string): value is PathnameWithSearchOrHash =>
	value.startsWith('/');

const resolvePathname: (path: PathnameWithSearchOrHash) => ResolvedPathname = resolve;

export const resolvePath = <T extends PathnameWithSearchOrHash>(path: T): ResolvedPathname =>
	resolvePathname(path);

export const resolveHref = (href: string): string =>
	isPathnameWithSearchOrHash(href) ? resolvePath(href) : href;
