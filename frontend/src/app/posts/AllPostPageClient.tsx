'use client';

import React, {useState, useEffect, useRef} from 'react';
import {getAllArticlesByPage, getArticleByCategory} from '@/api/article';
import {getArticlesByTag} from "@/api/tag";
import ArticlePageItem, {ArticlePreview} from '@/components/article/ArticlePageItem';
import {motion} from 'framer-motion';
import ArticlePageItemSkeleton from "@/app/posts/ArticlePostItemSkeleton";
import FloatingMenu from "@/components/menu/FloatingMenu";

const AllPostPageClient = ({initialArticles, category, tag}: {
    category?: string,
    tag?: string,
    initialArticles: ArticlePreview[]
}) => {
    const [articles, setArticles] = useState<ArticlePreview[]>(initialArticles);
    const [page, setPage] = useState(1);
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(initialArticles.length > 0);
    const observer = useRef<IntersectionObserver>(null);
    const [isSummaryShow, setIsSummaryShow] = useState(true);

    const lastArticleElementRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (loading) return;
        if (observer.current) observer.current.disconnect();

        observer.current = new IntersectionObserver(entries => {
            if (entries[0].isIntersecting && hasMore) {
                setLoading(true);
                if (category != null) {
                    getArticleByCategory(category, page + 1, 10, {next: {revalidate: 60}})
                        .then(newArticles => {
                            setArticles(prevArticles => [...prevArticles, ...newArticles]);
                            setPage(prevPage => prevPage + 1);
                            setHasMore(newArticles.length > 0);
                            setLoading(false);
                        });
                } else if (tag != null) {
                    getArticlesByTag(tag, page + 1, 10, {next: {revalidate: 60}})
                        .then(newArticles => {
                            setArticles(prevArticles => [...prevArticles, ...newArticles]);
                            setPage(prevPage => prevPage + 1);
                            setHasMore(newArticles.length > 0);
                            setLoading(false);
                        });
                } else {
                    getAllArticlesByPage(page + 1, 10, {next: {revalidate: 60}})
                        .then(newArticles => {
                            setArticles(prevArticles => [...prevArticles, ...newArticles]);
                            setPage(prevPage => prevPage + 1);
                            setHasMore(newArticles.length > 0);
                            setLoading(false);
                        });
                }
            }
        });

        if (lastArticleElementRef.current) {
            observer.current.observe(lastArticleElementRef.current);
        }
    }, [loading, hasMore, page, category, tag]);

    return (
        <div className="min-h-screen">
            {/* 文章列表容器 */}
            <motion.div 
                className="max-w-4xl mx-auto"
                initial={{opacity: 0}}
                animate={{opacity: 1}}
                transition={{duration: 0.6, ease: "easeOut"}}
            >
                {articles.map((item, index) => {
                    const delay = Math.min(index * 0.05, 0.3); // 限制最大延迟
                    if (articles.length === index + 1) {
                        return (
                            <motion.div
                                ref={lastArticleElementRef}
                                key={item.shortUrl}
                                initial={{opacity: 0, y: 20}}
                                animate={{opacity: 1, y: 0}}
                                transition={{
                                    duration: 0.4, 
                                    delay,
                                    ease: "easeOut"
                                }}
                            >
                                <ArticlePageItem post={item} isSummaryShow={isSummaryShow}/>
                            </motion.div>
                        );
                    } else {
                        return (
                            <motion.div
                                key={item.shortUrl}
                                initial={{opacity: 0, y: 20}}
                                animate={{opacity: 1, y: 0}}
                                transition={{
                                    duration: 0.4, 
                                    delay,
                                    ease: "easeOut"
                                }}
                            >
                                <ArticlePageItem post={item} isSummaryShow={isSummaryShow}/>
                            </motion.div>
                        );
                    }
                })}
            </motion.div>

            {/* 加载状态 */}
            {loading && (
                <motion.div
                    initial={{opacity: 0}}
                    animate={{opacity: 1}}
                    className="max-w-4xl mx-auto"
                >
                    <ArticlePageItemSkeleton/>
                </motion.div>
            )}

            {/* 无更多内容提示 */}
            {!hasMore && articles.length > 0 && (
                <motion.div 
                    className="text-center py-16"
                    initial={{opacity: 0, y: 20}}
                    animate={{opacity: 1, y: 0}}
                    transition={{delay: 0.3}}
                >
                    <div className="inline-flex items-center justify-center px-6 py-3 rounded-full bg-gradient-to-r from-gray-50 to-gray-100 dark:from-gray-800 dark:to-gray-700 border border-gray-200 dark:border-gray-600">
                        <span className="text-sm text-gray-500 dark:text-gray-400 font-medium">
                            🎉 没有更多啦，感谢你的耐心阅读
                        </span>
                    </div>
                </motion.div>
            )}

            {/* 浮动菜单 */}
            <FloatingMenu items={[
                {
                    type: 'switch',
                    value: isSummaryShow,
                    label: '显示摘要',
                    icon: <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                               xmlns="http://www.w3.org/2000/svg">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                    </svg>,
                    onClick: (e) => {
                        setIsSummaryShow(e ?? false);
                    }
                }
            ]}/>
        </div>
    );
};

export default AllPostPageClient;
