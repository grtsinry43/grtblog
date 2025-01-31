import {request} from './request';

export function getLatestThinking(options = {}) {
    return request('/thinking/latest', options);
}

export function getThinkingList(options = {}) {
    return request('/thinking/all', options);
}

export interface Thinking {
    id: string;
    author: string;
    content: string;
    createdAt: string;
}
