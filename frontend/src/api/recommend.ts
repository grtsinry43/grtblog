import { request } from './request';

export const getRecommend = (options = {}) => {
    return request('/recommend', options);
}
