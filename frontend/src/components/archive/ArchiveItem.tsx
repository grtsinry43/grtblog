'use client';

import React from 'react';
import {format} from 'date-fns';
import {clsx} from 'clsx';
import {article_font} from '@/app/fonts/font';
import {FileText, MessageCircle} from 'lucide-react';
import {motion} from 'framer-motion';
import {CombinedItem} from '@/app/archives/page';
import Link from 'next/link';
import "@/styles/LinkAnimate.scss";

const ArchiveItem = ({item, index}: { item: CombinedItem, index: number }) => {
    return (
        <motion.div
            key={`${item.type}-${item.shortUrl}`}
            initial={{opacity: 0, y: 10}}
            animate={{opacity: 1, y: 0}}
            transition={{
                duration: 0.4,
                delay: index * 0.05,
                type: "spring",
                stiffness: 100,
                damping: 15
            }}
            whileHover={{x: 3}}
            className="group"
        >
            <div className="relative pl-8 pb-3">
                {/* Timeline dot with subtle animation */}
                <div className="absolute left-0 top-[0.9rem]">
                    <motion.div
                        className={clsx(
                            "w-3 h-3 rounded-full z-10 transition-all duration-300 group-hover:scale-125",
                            item.type === 'article' ? 'bg-blue-500' : 'bg-green-500'
                        )}
                        whileHover={{scale: 1.2}}
                    />
                </div>

                {/* Timeline line with gradient */}
                <div
                    className="absolute left-[5px] top-7 bottom-0 w-[2px] bg-gradient-to-b from-gray-200 to-transparent dark:from-gray-700"/>

                <div className="flex flex-col sm:flex-row sm:items-center gap-2">
                    <div className="flex-1">
                        <div className="text-xs text-gray-500 dark:text-gray-400 mb-1 font-medium">
                            {format(new Date(item.createdAt), 'yyyy.MM.dd')}
                        </div>
                        <h3 className={clsx(article_font.className, 'text-base font-medium flex items-center')}>
              <span className={clsx(
                  "text-xs mr-2",
                  item.type === 'article'
                      ? 'text-blue-600 dark:text-blue-400'
                      : 'text-green-600 dark:text-green-400'
              )}>
                {item.type === 'article' ? (
                    <FileText size={14} className="inline"/>
                ) : (
                    <MessageCircle size={14} className="inline"/>
                )}
              </span>
                            <Link
                                href={item.type === 'article' ? `/posts/${item.shortUrl}` : `/moments/${item.shortUrl}`}
                                className="transition-colors duration-300 underlineAnimation glowAnimation hover:text-primary"
                            >
                                {item.title}
                            </Link>
                        </h3>
                    </div>
                    <div className={clsx(
                        'text-xs px-2 py-0.5 rounded-full font-medium whitespace-nowrap self-start sm:self-auto',
                        item.type === 'article'
                            ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-200'
                            : 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-200',
                    )}>
                        {item.type === 'article' ? '文章' : '记录'} / {item.category}
                    </div>
                </div>
            </div>
        </motion.div>
    );
};

export default ArchiveItem;
