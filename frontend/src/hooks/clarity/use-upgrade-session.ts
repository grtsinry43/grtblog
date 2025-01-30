import {useCallback} from 'react';
import Clarity from '@microsoft/clarity';

const useUpgradeSession = () => {
    return useCallback((reason: string) => {
        Clarity.upgrade(reason);
    }, []);
};

export default useUpgradeSession;
