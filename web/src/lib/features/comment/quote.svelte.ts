/**
 * Hand-off from "quote selection" in article bodies to the root comment form.
 * The comment area is lazy-loaded, so the quote waits here until the form mounts and takes it.
 */
let pending = $state<string | null>(null);

export const commentQuote = {
	get pending() {
		return pending;
	},
	request(text: string) {
		pending = text;
	},
	take() {
		const text = pending;
		pending = null;
		return text;
	}
};

/** Markdown blockquote for `text`, followed by a blank line for the reply. */
export function toQuoteBlock(text: string): string {
	return (
		text
			.trim()
			.split(/\r?\n/)
			.map((line) => `> ${line}`)
			.join('\n') + '\n\n'
	);
}
