export interface MarkdownComponentAttribute {
	key: string;
	label: string;
	placeholder?: string;
	defaultValue?: string;
	inputType?: 'text' | 'switch' | 'number';
}

export interface MarkdownComponentBody {
	label?: string;
	placeholder?: string;
	defaultValue?: string;
}

export interface MarkdownComponentDefinition {
	name: string;
	label: string;
	description?: string;
	attrs: MarkdownComponentAttribute[];
	body?: MarkdownComponentBody;
	insertTemplate: string;
}

export const markdownComponents: MarkdownComponentDefinition[] = [
	{
		name: 'gallery',
		label: 'Gallery',
		description: '相册组件',
		attrs: [
			{ key: 'height', label: '高度', defaultValue: '400px' },
			{ key: 'caption', label: '说明', placeholder: '相册说明' }
		],
		body: {
			label: '图片列表',
			placeholder: '每行一个图片 URL，或使用 Markdown 图片语法'
		},
		insertTemplate: '::: gallery height="400px" caption=""\nhttps://example.com/1.jpg\nhttps://example.com/2.jpg\n:::'
	},
	{
		name: 'callout',
		label: 'Callout',
		description: '提示框',
		attrs: [
			{
				key: 'type',
				label: '类型',
				defaultValue: 'info',
				placeholder: 'info, warning, error, success, quote, idea'
			},
			{ key: 'title', label: '标题', placeholder: '提示标题' }
		],
		body: {
			label: '内容',
			placeholder: '这里输入提示框正文'
		},
		insertTemplate: '::: callout type="info" title=""\n这里输入提示内容\n:::'
	},
	{
		name: 'timeline',
		label: 'Timeline',
		description: '时间轴',
		attrs: [
			{ key: 'title', label: '标题', placeholder: '时间轴标题' },
			{ key: 'sub', label: '副标题', placeholder: 'HISTORY', defaultValue: 'HISTORY' }
		],
		body: {
			label: '时间轴数据',
			placeholder: '每行格式: 时间|标题|描述'
		},
		insertTemplate:
			'::: timeline title="" sub="HISTORY"\n2024-01|项目启动|这是一个描述\n2025-02|发布上线|这是另一个描述\n:::'
	},
	{
		name: 'chat-history',
		label: 'Chat History',
		description: '聊天记录',
		attrs: [{ key: 'title', label: '标题', placeholder: 'Chat History', defaultValue: 'Chat History' }],
		body: {
			label: '聊天数据',
			placeholder: '每行格式: 角色|内容'
		},
		insertTemplate: '::: chat-history title="Chat History"\nAI|你好\nUser|你好呀\n:::'
	},
	{
		name: 'year-card',
		label: 'Year Card',
		description: '年终总结卡片',
		attrs: [
			{ key: 'url', label: '链接', placeholder: 'https://example.com' },
			{ key: 'title', label: '标题', placeholder: '2025 年终总结' },
			{ key: 'type', label: '类型', defaultValue: 'page' },
			{ key: 'cover', label: '封面图', placeholder: 'https://example.com/cover.jpg' },
			{ key: 'blur', label: '模糊度', defaultValue: '7px' }
		],
		insertTemplate: '::: year-card url="" title="" type="page" cover="" blur="7px"\n\n:::'
	},
	{
		name: 'link-card',
		label: 'Link Card',
		description: '链接卡片',
		attrs: [
			{ key: 'href', label: '链接', placeholder: '/path/to/page' },
			{ key: 'title', label: '标题', placeholder: '标题' },
			{ key: 'desc', label: '描述', placeholder: '描述' },
			{ key: 'newtab', label: '新窗口', defaultValue: 'true', inputType: 'switch' }
		],
		insertTemplate: '::: link-card href="" title="" desc="" newtab="true"\n\n:::'
	}
];

export const markdownComponentNames = new Set(markdownComponents.map((component) => component.name));

export const getMarkdownComponent = (name?: string) =>
	markdownComponents.find((component) => component.name === name);

export interface ParsedComponentInfo {
	name: string;
	attrs: Record<string, string>;
	rawAttrs: string;
}

const attrRegex = /([A-Za-z][\w-]*)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s]+))/g;

export const parseComponentAttributes = (raw: string) => {
	if (!raw) {
		return {};
	}
	const trimmed = raw.trim();
	if (trimmed.includes('{') || trimmed.includes('}')) {
		return {};
	}
	const attrs: Record<string, string> = {};
	let match: RegExpExecArray | null = null;
	attrRegex.lastIndex = 0;

	while ((match = attrRegex.exec(trimmed)) !== null) {
		const key = match[1];
		const value = match[2] ?? match[3] ?? match[4] ?? '';
		if (typeof key === 'string') {
			attrs[key] = value;
		}
	}

	return attrs;
};

export const parseComponentInfo = (info: string): ParsedComponentInfo => {
	const trimmed = info.trim();
	if (!trimmed) {
		return { name: 'unknown', attrs: {}, rawAttrs: '' };
	}

	let rest = trimmed;
	if (trimmed.startsWith('component')) {
		const match = trimmed.match(/^component\s+(.+)$/);
		rest = (match?.[1] || '').trim();
	}

	const match = rest.match(/^(?:\[([^\]\s]+)\]|([^\s\[{]+))(?:\s+(.+))?$/);
	const name = (match?.[1] || match?.[2] || 'unknown').trim();
	const rawAttrs = (match?.[3] || '').trim();

	return { name, attrs: parseComponentAttributes(rawAttrs), rawAttrs };
};

export const quoteComponentAttributeValue = (value: string) =>
	String(value)
		.replace(/\\/g, '\\\\')
		.replace(/"/g, '\\"')
		.replace(/\n/g, '\\n');

export const serializeComponentAttributes = (
	attrs: Record<string, string>,
	keyOrder?: string[]
) => {
	const keys = keyOrder?.length ? keyOrder : Object.keys(attrs);
	return keys
		.filter((key) => key in attrs)
		.map((key) => `${key}="${quoteComponentAttributeValue(attrs[key] ?? '')}"`)
		.join(' ');
};
