/**
 * Calculate reading time based on content
 * @param content - The markdown or HTML content
 * @returns Reading time in minutes
 */
export function calculateReadingTime(content: string): number {
	if (!content || content.trim().length === 0) {
		return 1;
	}

	// Remove HTML tags
	const plainText = content.replace(/<[^>]+>/g, '');

	// Remove markdown syntax
	const cleanText = plainText
		.replace(/!\[.*?\]\(.*?\)/g, '') // images
		.replace(/\[.*?\]\(.*?\)/g, '') // links
		.replace(/[#*`~\-_]/g, '') // formatting
		.replace(/```[\s\S]*?```/g, '') // code blocks
		.replace(/`[^`]*`/g, ''); // inline code

	// Count Chinese characters (CJK unified ideographs)
	const chineseChars = (cleanText.match(/[\u4e00-\u9fa5]/g) || []).length;

	// Count English words (sequences of letters/numbers)
	const englishWords = (cleanText.match(/[a-zA-Z0-9]+/g) || []).length;

	// Calculate reading time
	// Chinese: ~400 chars per minute
	// English: ~200 words per minute
	const chineseMinutes = chineseChars / 400;
	const englishMinutes = englishWords / 200;

	const totalMinutes = chineseMinutes + englishMinutes;

	// Round to nearest minute, minimum 1 minute
	return Math.max(1, Math.round(totalMinutes));
}

/**
 * Format reading time as a string
 * @param minutes - Reading time in minutes
 * @returns Formatted string like "5 分钟阅读"
 */
export function formatReadingTime(minutes: number): string {
	return `${minutes} 分钟阅读`;
}
