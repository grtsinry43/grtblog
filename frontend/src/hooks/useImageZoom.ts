'use client';

import { useState, useCallback } from 'react';

interface TriggerElementInfo {
    x: number;
    y: number;
    width: number;
    height: number;
}

interface UseImageZoomReturn {
    isOpen: boolean;
    triggerElement: TriggerElementInfo | undefined;
    openZoom: (element: HTMLElement) => void;
    closeZoom: () => void;
}

const useImageZoom = (): UseImageZoomReturn => {
    const [isOpen, setIsOpen] = useState(false);
    const [triggerElement, setTriggerElement] = useState<TriggerElementInfo | undefined>();

    const openZoom = useCallback((element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        setTriggerElement({
            x: rect.left,
            y: rect.top,
            width: rect.width,
            height: rect.height,
        });
        setIsOpen(true);
    }, []);

    const closeZoom = useCallback(() => {
        setIsOpen(false);
        // 延迟清除触发元素信息，等动画完成
        setTimeout(() => {
            setTriggerElement(undefined);
        }, 500);
    }, []);

    return {
        isOpen,
        triggerElement,
        openZoom,
        closeZoom,
    };
};

export default useImageZoom; 