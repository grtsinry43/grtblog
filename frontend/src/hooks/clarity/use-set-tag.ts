import {useCallback} from 'react';
import Clarity from '@microsoft/clarity';

const useSetTag = () => {
    return useCallback((key: string, value: string | string[]) => {
        Clarity.setTag(key, value);
    }, []);
};

export default useSetTag;
