'use client';

import React from 'react';
import {motion} from 'framer-motion';
import {Heart, MessageCircle, Eye} from 'lucide-react';
import {ArticlePreview} from '@/types';
import Link from 'next/link';
import Image from "next/image";
import {useWebsiteInfo} from "@/app/website-info-provider";

export default function RecommendCard({item}: { item: ArticlePreview }) {
    const websiteInfo = useWebsiteInfo();
    return (
        <motion.div
            initial={{opacity: 0, y: 20}}
            animate={{opacity: 1, y: 0}}
            transition={{duration: 0.5}}
        >
            <div style={{
                width: '100%',
                maxWidth: '300px',
                display: 'flex',
                flexDirection: 'column',
                minWidth: '350px',
                margin: '0 auto',
                minHeight: '100%',
                borderRadius: '10px',
                border: '1px solid rgba(var(--foreground), 0.1)',
                overflow: 'hidden',
            }}
                 className="m-1 duration-300 ease-in-out">
                {item.cover ? (
                    <div className="relative h-16 overflow-hidden">
                        <Image
                            src={websiteInfo.WEBSITE_URL + item.cover}
                            alt={item.title}
                            layout="fill"
                            objectFit="cover"
                            className="transition-transform duration-300 ease-in-out transform hover:scale-105"
                        />
                    </div>
                ) : (
                    <div style={{
                        color: "rgb(var(--primary))",
                        backgroundColor: "rgba(var(--primary), 0.1)",
                    }} className="h-16 w-full"/>
                )}
                <div className="p-4">
                    <Link href={`/posts/${item.shortUrl}`}>
                        <h2 className="font-semibold mb-2 line-clamp-2">{item.title}</h2>
                    </Link>
                    <p className="text-muted-foreground mb-4 line-clamp-3 text-sm opacity-70">{item.summary.length > 50 ? item.summary.slice(0, 50) + '...' : item.summary}</p>
                    <div className="flex items-center justify-between text-sm text-muted-foreground">
                        <div className="flex items-center space-x-2">
                            <span>{item.authorName}</span>
                        </div>
                        <span>{new Date(item.createdAt).toLocaleDateString()}</span>
                    </div>
                </div>
                <div className="p-4 flex justify-center gap-6 items-center text-sm self-end">
                    <motion.div whileHover={{scale: 1.1}}>
                        <div className="flex items-center space-x-1">
                            <Eye className="w-3 h-3"/>
                            <span>{item.views}</span>
                        </div>
                    </motion.div>
                    <motion.div whileHover={{scale: 1.1}}>
                        <div className="flex items-center space-x-1">
                            <MessageCircle className="w-3 h-3"/>
                            <span>{item.comments}</span>
                        </div>
                    </motion.div>
                    <motion.div whileHover={{scale: 1.1}}>
                        <div className="flex items-center space-x-1">
                            <Heart className="w-3 h-3"/>
                            <span>{item.likes}</span>
                        </div>
                    </motion.div>
                </div>
            </div>
        </motion.div>
    );
}
