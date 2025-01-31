'use client';

import React, {useState, useEffect, useRef} from 'react';
import {getAllArticlesByPage, getArticleByCategory} from '@/api/article';
import {getArticlesByTag} from "@/api/tag";
import ArticlePageItem, {ArticlePreview} from '@/components/article/ArticlePageItem';
import {Skeleton, Text} from '@radix-ui/themes';
import {motion} from 'framer-motion';
import {HashtagIcon, TagIcon} from '@heroicons/react/24/outline';
import {Calendar, Eye, ThumbsUpIcon} from 'lucide-react';
import {AiOutlineComment} from 'react-icons/ai';

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
        <>
            {articles.map((item, index) => {
                const delay = index % 10 * 0.15;
                if (articles.length === index + 1) {
                    return (
                        <motion.div
                            ref={lastArticleElementRef}
                            key={item.shortUrl}
                            initial={{opacity: 0, y: 20}}
                            animate={{opacity: 1, y: 0}}
                            transition={{type: 'spring', stiffness: 500, damping: 100, delay, bounce: 0.7, mass: 0.5}}
                        >
                            <ArticlePageItem post={item}/>
                        </motion.div>
                    );
                } else {
                    return (
                        <motion.div
                            key={item.shortUrl}
                            initial={{opacity: 0, y: 20}}
                            animate={{opacity: 1, y: 0}}
                            transition={{type: 'spring', stiffness: 100, damping: 10, delay, bounce: 0.3}}
                        >
                            <ArticlePageItem post={item}/>
                        </motion.div>
                    );
                }
            })}
            {loading && (
                <Skeleton>
                    <Text
                        className="text-xl font-bold transition"> 六月初至七月中旬｜前端学习简要总结，生活的小回顾 </Text>
                    <Text
                        className="text-gray-500 text-sm mt-2"> 在这段时间里我通过完成多个项目，深入学习了前端和全栈开发的技术。主要使用了
                        Vue3、C++ Web 框架、jwt 鉴权、Docker、Nginx、FastAPI、微信网页开发、Vue2、Vant 组件库、TypeScript 和
                        React 等技术。
                    </Text>
                    <div className="commentMeta mt-2 text-sm text-gray-700 dark:text-gray-300">
                        <div className="flex flex-wrap">
                            <div className="flex items-center mr-4 mb-2">
                                <HashtagIcon width={12} height={12} className="inline-block mr-1"/>
                                <Text> 技术学习 </Text>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <Calendar width={12} height={12} className="inline-block mr-1"/>
                                <Text> 大约一天前 </Text>
                                <Text className="text-xs text-gray-450 dark:text-gray-550"> （更新于不到一分钟前）</Text>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <TagIcon width={12} height={12} className="inline-block mr-1"/>
                                <Text> grtsinry43</Text>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <Eye width={12} height={12} className="inline-block mr-1"/>
                                <Text> 10</Text>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <AiOutlineComment width={12} height={12} className="inline-block mr-1"/>
                                <Text> 11</Text>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <ThumbsUpIcon width={12} height={12} className="inline-block mr-1"/>
                                <Text> 12</Text>
                            </div>
                        </div>
                    </div>
                </Skeleton>
            )}
            {!hasMore && (
                <div className="text-center text-gray-500 text-sm mt-2"> 没有更多啦，不小心让你翻到底了欸 〃•ω‹〃</div>
            )}
        </>
    );
};

export default AllPostPageClient;
