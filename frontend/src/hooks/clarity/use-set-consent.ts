import {useCallback} from 'react';
import Clarity from '@microsoft/clarity';

const useSetConsent = () => {
    return useCallback((consent: boolean = true) => {
        Clarity.consent(consent);
    }, []);
};

export default useSetConsent;
