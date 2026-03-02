// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	const __APP_VERSION__: string;
	const __BUILD_COMMIT__: string;

	interface ViewTransition {
		ready: Promise<void>;
	}

	interface Document {
		startViewTransition?: (callback: () => void) => ViewTransition;
	}

	namespace App {
		// interface Error {}
		interface Locals {
			isrDeps: Set<string>;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
