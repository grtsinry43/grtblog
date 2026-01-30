import YearCard from './YearCard.svelte';
import LinkCard from './LinkCard.svelte';
import MarkdownLink from './MarkdownLink.svelte';
import MarkdownImage from './MarkdownImage.svelte';
import MarkdownCodeBlock from './MarkdownCodeBlock.svelte';
import FootnoteLinkCard from './FootnoteLinkCard.svelte';
import { registerMarkdownComponent } from './index';

registerMarkdownComponent('year-card', YearCard);
registerMarkdownComponent('link-card', LinkCard);
registerMarkdownComponent('md-link', MarkdownLink);
registerMarkdownComponent('md-image', MarkdownImage);
registerMarkdownComponent('md-codeblock', MarkdownCodeBlock);
registerMarkdownComponent('footnote-link-card', FootnoteLinkCard);
