import {ofetch} from 'ofetch';
import {getToken, setToken} from '@/utils/token';

const ins = ofetch.create({
    baseURL: process.env.NEXT_PUBLIC_BASE_API_URL || '/api/v1', // Ensure it works in both browser and Node.js environments
    async onRequest({options}) {
        const token = getToken();

        if (token) {
            options.headers = new Headers({
                ...options.headers,
                Authorization: `Bearer ${token}`,
            });
        }
    },
    async onResponse({response}) {


        const token = response.headers.get('Authorization');
        if (token) {
            setToken(token);
        }
    },
});

export const request = async (url: string, options?: RequestInit) => {
    const res = await ins(url, options);


    if (res.code !== 0) {
        // if (typeof window !== 'undefined') {
        //     const notificationUtil = useNotificationUtil();
        //     notificationUtil(res.msg, "error");
        // }
        return null;
    } else {
        return res.data;
    }
};

const ins_bff = ofetch.create({
    baseURL: process.env.NEXT_PUBLIC_BFF_BASE_URL || '/api/bff', // Ensure it works in both browser and Node.js environments
    async onRequest({options}) {
        const token = getToken();
        if (token) {
            options.headers = new Headers({
                ...options.headers,
                Authorization: `Bearer ${token}`,
            });
        }
    },
    async onResponse({response}) {
        const token = response.headers.get('Authorization');
        if (token) {
            setToken(token);
        }
    },
});

export const request_bff = async (url: string, options?: RequestInit) => {
    const res = await ins_bff(url, options);


    if (res.code !== 0) {
        throw new Error(res.msg);
    } else {
        return res.data;
    }
};
