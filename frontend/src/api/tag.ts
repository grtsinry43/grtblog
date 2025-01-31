import {request} from './request';

export function getAllTags(options = {}) {
    return request('/tag', options);
}

export function getAllTagNames(options = {}) {
    return request(`/tag/names`, options);
}

export function getArticlesByTag(tag: string, page: number, pageSize: number, options = {}) {
    return request(`/article/tag/${tag}?page=${page}&pageSize=${pageSize}`, options);
}
