import {request} from './request';

export const getPluginList = (options = {}) => {
    return request(`/plugins/api/all-endpoints`, options);
}

export interface PluginFetchItem {
    endpoint: string;
    name: string;
}
