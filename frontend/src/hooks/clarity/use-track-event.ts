import {useCallback} from 'react';
import Clarity from '@microsoft/clarity';

const useTrackEvent = () => {
    return useCallback((eventName: string) => {
        Clarity.event(eventName);
    }, []);
};

export default useTrackEvent;
