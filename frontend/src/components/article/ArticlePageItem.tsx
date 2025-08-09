import React from 'react';
import {article_font} from '@/app/fonts/font';
import {HashtagIcon, TagIcon} from '@heroicons/react/24/outline';
import Link from 'next/link';
import {formatDistanceToNow, parseISO} from 'date-fns';
import {zhCN} from 'date-fns/locale';
import {Calendar, Eye, ThumbsUpIcon} from 'lucide-react';
import {AiOutlineComment} from 'react-icons/ai';
import {motion, AnimatePresence} from 'framer-motion';
import useIsMobile from '@/hooks/useIsMobile';
import {PinTopIcon} from '@radix-ui/react-icons';

export type ArticlePreview = {
    authorName: string,
    categoryShortUrl: string,
    comments: number,
    cover: string | null,
    createdAt: string,
    updatedAt: string,
    categoryName: string,
    shortUrl: string,
    tags: string,
    isTop: boolean,
    likes: number,
    summary: string,
    title: string,
    views: number
}

const ArticlePageItem = ({post, isSummaryShow}: { post: ArticlePreview, isSummaryShow: boolean }) => {
    const formattedCreatedDate = formatDistanceToNow(parseISO(post.createdAt), {addSuffix: true, locale: zhCN});
    const formattedUpdatedDate = formatDistanceToNow(parseISO(post.updatedAt), {addSuffix: true, locale: zhCN});
    const isMobile = useIsMobile();

    return (
        <div className={article_font.className}>
            <motion.article
                className="group relative py-8 transition-all duration-500 ease-out"
                initial={{opacity: 0, y: 20}}
                animate={{opacity: 1, y: 0}}
                whileHover={{
                    x: 8,
                    transition: { duration: 0.3, ease: "easeOut" }
                }}
                style={{
                    background: 'linear-gradient(90deg, transparent 0%, rgba(var(--primary), 0.015) 100%)',
                    borderLeft: '1px solid transparent',
                }}
                onHoverStart={() => {}}
                onHoverEnd={() => {}}
            >
                {/* 悬浮时的左侧装饰线 */}
                <motion.div
                    className="absolute left-0 top-0 h-full w-0.5 bg-gradient-to-b from-transparent via-blue-500/60 to-transparent opacity-0 group-hover:opacity-100"
                    initial={{scaleY: 0}}
                    whileHover={{scaleY: 1}}
                    transition={{duration: 0.4, ease: "easeOut"}}
                />
                
                {/* 置顶标识 */}
                {post.isTop && (
                    <motion.div
                        className="absolute -top-2 right-0 flex items-center text-amber-500 dark:text-amber-400"
                        initial={{opacity: 0, scale: 0.8}}
                        animate={{opacity: 1, scale: 1}}
                        transition={{delay: 0.2}}
                    >
                        <PinTopIcon className="w-4 h-4 mr-1" />
                        <span className="text-xs font-medium">置顶</span>
                    </motion.div>
                )}

                <Link href={`/posts/${post.shortUrl}`} className="block group-hover:no-underline">
                    {/* 标题 */}
                    <motion.h2 
                        className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-3 leading-tight group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors duration-300"
                        whileHover={{
                            textShadow: "0 0 8px rgba(59, 130, 246, 0.3)"
                        }}
                    >
                        {post.title}
                    </motion.h2>

                    {/* 摘要 */}
                    <AnimatePresence>
                        {isSummaryShow && (
                            <motion.div
                                className="text-gray-600 dark:text-gray-400 leading-relaxed mb-4 text-sm"
                                initial={{opacity: 0, height: 0}}
                                animate={{opacity: 1, height: 'auto'}}
                                exit={{opacity: 0, height: 0}}
                                transition={{duration: 0.3, ease: "easeInOut"}}
                            >
                                <span className="line-clamp-2">
                                    {isMobile ? post.summary.length > 100 ? post.summary.slice(0, 100) + '...' : post.summary : post.summary}
                                </span>
                            </motion.div>
                        )}
                    </AnimatePresence>

                    {/* 元数据 */}
                    <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                        {/* 分类 */}
                        <motion.div 
                            className="flex items-center group/category"
                            whileHover={{scale: 1.05}}
                        >
                            {post.categoryShortUrl ? (
                                <Link href={`/categories/${post.categoryShortUrl}`}
                                      className="flex items-center hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                                    <HashtagIcon className="w-3 h-3 mr-1" />
                                    <span className="font-medium">{post.categoryName}</span>
                                </Link>
                            ) : (
                                <div className="flex items-center">
                                    <HashtagIcon className="w-3 h-3 mr-1" />
                                    <span className="font-medium">{post.categoryName}</span>
                                </div>
                            )}
                        </motion.div>

                        {/* 时间 */}
                        <div className="flex items-center">
                            <Calendar className="w-3 h-3 mr-1" />
                            <span>{formattedCreatedDate}</span>
                            {post.createdAt !== post.updatedAt && (
                                <span className="ml-1 text-gray-400 dark:text-gray-500">
                                    (更新于 {formattedUpdatedDate})
                                </span>
                            )}
                        </div>

                        {/* 标签 */}
                        {post.tags && post.tags.split(',').map((tag, index) => (
                            <motion.div key={index} whileHover={{scale: 1.05}}>
                                <Link href={`/tags/${tag}`} 
                                      className="flex items-center hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                                    <TagIcon className="w-3 h-3 mr-1" />
                                    <span>{tag}</span>
                                </Link>
                            </motion.div>
                        ))}

                        {/* 统计信息 */}
                        <div className="flex items-center gap-3 ml-auto">
                            <div className="flex items-center">
                                <Eye className="w-3 h-3 mr-1" />
                                <span>{post.views}</span>
                            </div>
                            <div className="flex items-center">
                                <AiOutlineComment className="w-3 h-3 mr-1" />
                                <span>{post.comments}</span>
                            </div>
                            <div className="flex items-center">
                                <ThumbsUpIcon className="w-3 h-3 mr-1" />
                                <span>{post.likes}</span>
                            </div>
                        </div>
                    </div>
                </Link>

                {/* 底部分隔线 - 更subtle的效果 */}
                <motion.div 
                    className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-gray-200/30 dark:via-gray-700/20 to-transparent"
                    initial={{scaleX: 0}}
                    animate={{scaleX: 1}}
                    transition={{delay: 0.2, duration: 0.5}}
                />
            </motion.article>
        </div>
    );
};

export default ArticlePageItem;
