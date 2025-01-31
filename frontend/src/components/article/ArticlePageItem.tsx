import React from 'react';
import {article_font} from '@/app/fonts/font';
import {HashtagIcon, TagIcon} from '@heroicons/react/24/outline';
import Link from 'next/link';
import {formatDistanceToNow, parseISO} from 'date-fns';
import {zhCN} from 'date-fns/locale';
import {Calendar, Eye, ThumbsUpIcon} from 'lucide-react';
import {AiOutlineComment} from 'react-icons/ai';
import {motion} from 'framer-motion';
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

const ArticlePageItem = ({post}: { post: ArticlePreview }) => {
    const formattedCreatedDate = formatDistanceToNow(parseISO(post.createdAt), {addSuffix: true, locale: zhCN});
    const formattedUpdatedDate = formatDistanceToNow(parseISO(post.updatedAt), {addSuffix: true, locale: zhCN});
    const isMobile = useIsMobile();

    return (
        <div className={article_font.className}>
            <motion.div
                style={{width: '100%', margin: '2em auto', perspective: '1000px'}}
                whileTap={{scale: 0.95}}
                whileHover={{
                    scale: 1.03,
                    backgroundColor: 'rgba(var(--foreground), 0.05)',
                    borderRadius: '0.5em',
                    padding: '1em',
                    transition: {duration: 0.2},
                }}
                transition={{type: 'spring', stiffness: 300, mass: 0.3}}
            >
                <Link href={`/posts/${post.shortUrl}`}>
                    <h1 className="text-xl font-bold transition">{post.title}</h1>
                    <div
                        className="text-gray-500 text-sm mt-2">{isMobile ? post.summary.length > 100 ? post.summary.slice(0, 100) + '...' : post.summary : post.summary}</div>
                    <div className="commentMeta mt-2 text-sm text-gray-700 dark:text-gray-300">
                        <div className="flex flex-wrap">
                            <div className="flex items-center mr-4 mb-2">
                                {
                                    post.categoryShortUrl ? (
                                        <Link href={`/categories/${post.categoryShortUrl}`} className="underlineAnimation">
                                            <HashtagIcon width={12} height={12} className="inline-block mr-1"/>
                                            <span>{post.categoryName}</span>
                                        </Link>
                                    ) : (
                                        <>
                                            <HashtagIcon width={12} height={12} className="inline-block mr-1"/>
                                            <span>{post.categoryName}</span>
                                        </>
                                    )
                                }
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <Calendar width={12} height={12} className="inline-block mr-1"/>
                                <span> {formattedCreatedDate}</span>
                                {post.createdAt !== post.updatedAt &&
                                    <span
                                        className="text-xs text-gray-450 dark:text-gray-550"> （更新于 {formattedUpdatedDate}）</span>}
                            </div>
                            {post.tags && post.tags.split(',').map((tag, index) => (
                                <Link href={`/tags/${tag}`} key={index}>
                                    <div className="flex items-center mr-4 mb-2">
                                        <TagIcon width={12} height={12} className="inline-block mr-1"/>
                                        <span className="underlineAnimation"> {tag}</span>
                                    </div>
                                </Link>
                            ))}
                            <div className="flex items-center mr-4 mb-2">
                                <Eye width={12} height={12} className="inline-block mr-1"/>
                                <span> {post.views}</span>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <AiOutlineComment width={12} height={12} className="inline-block mr-1"/>
                                <span> {post.comments}</span>
                            </div>
                            <div className="flex items-center mr-4 mb-2">
                                <ThumbsUpIcon width={12} height={12} className="inline-block mr-1"/>
                                <span> {post.likes}</span>
                            </div>
                        </div>
                    </div>
                    {post.isTop && (
                        <PinTopIcon width={16} height={16} className="absolute right-4 bottom-4"/>
                    )}
                </Link>
            </motion.div>
        </div>
    );
};

export default ArticlePageItem;
