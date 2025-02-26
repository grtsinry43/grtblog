import {request} from '@/api/request';

export const addNewComment = async (notLoginComment: {
    areaId: string,
    content: string,
    userName: string,
    email: string,
    website?: string,
    parentId: string
}) => {
    return request('/comment', {
        method: 'POST',
        body: JSON.stringify(notLoginComment),
    });
};

export const addNewCommentLogin = async (LoginComment: {
    areaId: string,
    content: string,
    parentId: string
}) => {
    return request('/comment/add', {
        method: 'POST',
        body: JSON.stringify(LoginComment),
    });
};

export const getComments = async (id: string, page: number, pageSize: number) => {
    return request(`/comment/${id}?page=${page}&pageSize=${pageSize}`);
};
