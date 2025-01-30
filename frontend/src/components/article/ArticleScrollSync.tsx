'use client';

import React, {useRef, useEffect} from 'react';
import eventBus from '@/utils/eventBus';
import {Post} from "@/components/article/ArticleView";
import {PageVO} from "@/app/[slug]/PageView";
import {MomentView} from "@/components/moment/MomentReading";
import useSetTag from "@/hooks/clarity/use-set-tag";

export type TitleEvent = {
    type: string;
    title: string;
    categoryName: string;
};

const ArticleScrollSync = ({children, type, post}: {
    children: React.ReactNode,
    type: string,
    post: Post | MomentView | PageVO
}) => {
    const articleRef = useRef<HTMLDivElement>(null);
    const [curScrollLength, setCurScrollLength] = React.useState(0);
    const setTag = useSetTag();

    useEffect(() => {
        if (post) {
            setTag('postType', type);
            setTag('postTitle', post.title);
            if ('categoryName' in post) {
                setTag('categoryName', post.categoryName);
            }
        }
    }, [post, type, setTag]);

    useEffect(() => {
        const handleScroll = () => {
            if (articleRef.current) {
                const scrollLength = -articleRef.current.getBoundingClientRect().top;
                const elementHeight = articleRef.current.clientHeight;
                const windowHeight = window.innerHeight;
                let scrollPercentage = Math.min((scrollLength / (elementHeight - windowHeight)) * 100, 100);
                if (scrollPercentage < 0) {
                    scrollPercentage = 0;
                }
                if (scrollLength > 68 && scrollLength > curScrollLength) {
                    eventBus.emit('showTitle', {
                        type,
                        title: post.title,
                        categoryName: post.categoryName ? post.categoryName : ''
                    } as TitleEvent);
                } else {
                    eventBus.emit('hideTitle');
                }

                setCurScrollLength(scrollLength);
                eventBus.emit('scroll', articleRef.current);
                eventBus.emit('readingProgress', scrollPercentage);
            }
        };

        const articleElement = articleRef.current;
        if (articleElement) {
            document.addEventListener('scroll', handleScroll);
        }

        return () => {
            if (articleElement) {
                document.removeEventListener('scroll', handleScroll);
            }
        };
    }, [curScrollLength, post.title, post.categoryName]);

    return (
        <div ref={articleRef}>
            {children}
        </div>
    );
};

export default ArticleScrollSync;
