import { resolve } from '$app/paths';

/**
 * Type-safe wrapper around SvelteKit's `resolve()`.
 *
 * `resolve` uses conditional generic types (`ResolveArgs<T>`) that
 * fail inference when the argument is a runtime `string` rather than
 * a literal route-id.  This wrapper performs the cast once so every
 * call site stays clean.
 */
// @ts-expect-error — SvelteKit `resolve` overload cannot narrow a plain `string`
export const resolvePath: (path: string) => string = resolve;
