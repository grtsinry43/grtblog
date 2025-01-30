import {useEffect} from 'react';
import Clarity from '@microsoft/clarity';

const projectId = process.env.CLARITY_PROJECT_ID || '';

const useClarityInit = () => {
    useEffect(() => {
        Clarity.init(projectId);
        console.log('Clarity initialized');
    }, []);
};

export default useClarityInit;
