'use client';

import React, {useState, useEffect, useRef} from 'react';
import {getAllSharesByPage} from '@/api/share';
import MomentCardItem from '@/components/moment/MomentCardItem';
import {Moment} from '@/types';
import {motion} from 'framer-motion';
import MomentCardItemSkeleton from "@/components/moment/MomentCardItemSkeleton";

const AllMomentClient = ({initialMoments}: { initialMoments: Moment[] }) => {
    const [moments, setMoments] = useState<Moment[]>(initialMoments);
    const [page, setPage] = useState(1);
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(initialMoments.length > 0);
    const observer = useRef<IntersectionObserver>(null);

    const lastMomentElementRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (loading) return;
        if (observer.current) observer.current.disconnect();

        observer.current = new IntersectionObserver(entries => {
            if (entries[0].isIntersecting && hasMore) {
                setLoading(true);
                getAllSharesByPage(page + 1, 5)
                    .then(newMoments => {
                        setMoments(prevMoments => [...prevMoments, ...newMoments]);
                        setPage(prevPage => prevPage + 1);
                        setHasMore(newMoments.length > 0);
                        setLoading(false);
                    });
            }
        });

        if (lastMomentElementRef.current) {
            observer.current.observe(lastMomentElementRef.current);
        }
    }, [loading, hasMore, page]);

    return (
        <div className="min-h-screen">
            {/* 背景装饰 */}
            <div className="fixed inset-0 overflow-hidden pointer-events-none">
                <div className="absolute top-1/4 -left-4 w-72 h-72 bg-blue-300/20 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob"/>
                <div className="absolute top-1/3 -right-4 w-72 h-72 bg-purple-300/20 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-2000"/>
                <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-300/20 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-4000"/>
            </div>

            <div className="container mx-auto p-4 sm:p-6 max-w-3xl relative">
                <motion.div
                    initial={{opacity: 0, y: 20}}
                    animate={{opacity: 1, y: 0}}
                    transition={{duration: 0.6}}
                    className="relative"
                >
                    {/* 精美的时间轴线 */}
                    <div className="absolute left-1/2 top-0 bottom-0 w-px transform -translate-x-1/2">
                        <div className="w-full h-full bg-gradient-to-b from-blue-200 via-purple-200 to-pink-200 dark:from-blue-800 dark:via-purple-800 dark:to-pink-800 opacity-60"/>
                        <div className="absolute inset-0 w-full bg-gradient-to-b from-transparent via-white/50 to-transparent dark:via-gray-900/50 opacity-30"/>
                    </div>

                    {/* 动态卡片列表 */}
                    <div className="space-y-8">
                        {moments.map((moment, index) => {
                            if (moments.length === index + 1) {
                                return (
                                    <motion.div
                                        ref={lastMomentElementRef}
                                        key={moment.shortUrl}
                                        initial={{opacity: 0, y: 30, scale: 0.95}}
                                        animate={{opacity: 1, y: 0, scale: 1}}
                                        transition={{
                                            type: 'spring', 
                                            stiffness: 100, 
                                            damping: 15,
                                            delay: Math.min(index * 0.1, 1)
                                        }}
                                    >
                                        <MomentCardItem moment={moment} index={index}/>
                                    </motion.div>
                                );
                            } else {
                                return (
                                    <motion.div
                                        key={moment.shortUrl}
                                        initial={{opacity: 0, y: 30, scale: 0.95}}
                                        animate={{opacity: 1, y: 0, scale: 1}}
                                        transition={{
                                            type: 'spring', 
                                            stiffness: 120, 
                                            damping: 20,
                                            delay: Math.min(index * 0.1, 1)
                                        }}
                                    >
                                        <MomentCardItem moment={moment} index={index}/>
                                    </motion.div>
                                );
                            }
                        })}
                    </div>
                </motion.div>

                {/* 加载状态 */}
                {loading && (
                    <motion.div
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        exit={{opacity: 0, y: -20}}
                        className="mt-8"
                    >
                        <MomentCardItemSkeleton/>
                    </motion.div>
                )}

                {/* 结束提示 */}
                {!hasMore && moments.length > 0 && (
                    <motion.div 
                        initial={{opacity: 0, scale: 0.8}}
                        animate={{opacity: 1, scale: 1}}
                        transition={{delay: 0.3}}
                        className="text-center mt-12 mb-8"
                    >
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full mb-4 shadow-lg">
                            <div className="w-8 h-8 bg-white rounded-full flex items-center justify-center">
                                <span className="text-gray-600 text-lg">✨</span>
                            </div>
                        </div>
                        <p className="text-gray-500 dark:text-gray-400 text-sm max-w-xs mx-auto leading-relaxed">
                            没有更多动态啦~ <br/>
                            <span className="text-xs opacity-75">不小心让你翻到底了呢 〃•ω‹〃</span>
                        </p>
                    </motion.div>
                )}

                {/* 空状态 */}
                {!loading && moments.length === 0 && (
                    <motion.div
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        className="text-center py-20"
                    >
                        <div className="w-24 h-24 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mb-6 mx-auto">
                            <span className="text-2xl text-gray-400">📝</span>
                        </div>
                        <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">还没有动态</h3>
                        <p className="text-gray-500 dark:text-gray-400 text-sm">期待第一条分享~</p>
                    </motion.div>
                )}
            </div>
        </div>
    );
};

export default AllMomentClient;
