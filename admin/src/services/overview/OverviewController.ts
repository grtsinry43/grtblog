import { Overview } from '@/services/overview/typings';
import { ApiResponse } from '@/services/typings';
import { request } from '@umijs/max';

const getOverview = async (): Promise<ApiResponse<Overview>> => {
    return request('/api/v1/overview', {
        method: 'GET',
    });
};

export default {
    getOverview,
};
