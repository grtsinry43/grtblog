import {useEffect} from 'react';
import Clarity from '@microsoft/clarity';

const useIdentify = (customId: string, customSessionId?: string, customPageId?: string, friendlyName?: string) => {
    useEffect(() => {
        Clarity.identify(customId, customSessionId, customPageId, friendlyName);
    }, [customId, customSessionId, customPageId, friendlyName]);
};

export default useIdentify;
