import {request} from './request';

export const reportRead = (articleId: string, options = {}) => {
    return request(`/user-behavior/read/${articleId}`, options);
}
