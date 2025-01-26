import { request } from './request';

export function getAuthorStatus(options = {}) {
    return request('/onlineStatus', options);
}
