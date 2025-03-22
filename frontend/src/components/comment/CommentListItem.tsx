'use client';

import React, {useEffect, useState} from 'react';
import {Comment} from '@/types';
import styles from '@/styles/comment/CommentList.module.scss';
import rehypeSanitize from 'rehype-sanitize';
import remarkGfm from 'remark-gfm';
import InlineCodeBlock from '@/components/InlineCodeBlock';
import CodeBlock from '@/components/CodeBlock';
import {clsx} from 'clsx';
import TableView from '@/components/article/TableView';
import ReactMarkdown from 'react-markdown';
import {Avatar, Button, Link, Tooltip} from '@radix-ui/themes';
import {formatDistanceToNow, parseISO} from 'date-fns';
import {zhCN} from 'date-fns/locale';
import {motion} from 'framer-motion';
import CommentForm from "@/components/comment/CommentForm";
import CommentList from "@/components/comment/CommentList";
import {article_font} from "@/app/fonts/font";
import {FaWindows, FaApple, FaLinux, FaAndroid, FaEdge} from 'react-icons/fa';
import {SiSafari, SiFirefox, SiGooglechrome} from 'react-icons/si';
import {VscVerifiedFilled} from "react-icons/vsc";
import {PinTopIcon} from "@radix-ui/react-icons";
import {Reply} from "lucide-react";

const getPlatformIcon = (platform: string) => {
    if (platform.includes('Android')) return <FaAndroid size={'10'} style={{color: '#3DDC84'}}/>;
    if (platform.includes('Windows')) return <FaWindows size={'10'} style={{color: '#0078D7'}}/>;
    if (platform.includes('Mac')) return <FaApple size={'10'} style={{color: '#787878'}}/>;
    if (platform.includes('Linux')) return <FaLinux size={'10'} style={{color: 'rgba(207,144,2,0.9)'}}/>;
    return null;
};

const getBrowserIcon = (browser: string) => {
    if (browser.includes('Chrome')) return <SiGooglechrome size={'10'} style={{color: '#4285F4'}}/>;
    if (browser.includes('Firefox')) return <SiFirefox size={'10'} style={{color: '#FF7139'}}/>;
    if (browser.includes('Safari')) return <SiSafari size={'10'} style={{color: '#0078D7'}}/>;
    if (browser.includes('Edge')) return <FaEdge size={'10'} style={{color: '#0078D7'}}/>;
    return null;
};

const CommentListItem = ({comment}: { comment: Comment }) => {
    const [formattedCreatedAt, setFormattedCreatedAt] = useState('');

    useEffect(() => {
        const updateFormattedTime = () => {
            const now = new Date()
            const createdAt = parseISO(comment.createdAt)
            const diffInSeconds = (now.getTime() - createdAt.getTime()) / 1000

            if (diffInSeconds < 60) {
                if (diffInSeconds < 2) {
                    setFormattedCreatedAt('刚刚 ')
                    return
                }
                setFormattedCreatedAt(`${Math.floor(diffInSeconds)} 秒前 `)
            } else if (diffInSeconds < 3600) {
                setFormattedCreatedAt(`${Math.floor(diffInSeconds / 60)} 分钟前 `)
            } else if (diffInSeconds < 86400) {
                setFormattedCreatedAt(`${Math.floor(diffInSeconds / 3600)} 小时前 `)
            } else {
                setFormattedCreatedAt(formatDistanceToNow(createdAt, {addSuffix: true, locale: zhCN}))
            }
        }

        updateFormattedTime()

        const interval = setInterval(updateFormattedTime, 1000)

        return () => clearInterval(interval)
    }, [comment.createdAt])

    const [isReplying, setIsReplying] = React.useState(false);

    return (
        <motion.div
            initial={{opacity: 0, y: 20}}
            animate={{opacity: 1, y: 0}}
            transition={{type: 'spring', stiffness: 100}}
        >
            <div className={clsx(article_font.className)}>
                <div className="flex items-center">
                    <Avatar src={comment.avatarUrl} radius={"full"}
                            size={"2"} className="mr-2"
                            alt={comment.userName}
                            fallback={comment.userName === '' ? '' : comment.userName[0]}/>
                    {comment.website ? (
                        <>
                            {(() => {
                                if (!comment.website.startsWith('http')) {
                                    comment.website = `https://${comment.website}`;
                                }
                                return (
                                    <div className="flex items-center mr-2">
                                        <a href={comment.website} target="_blank" rel="noopener noreferrer"
                                           className="">
                                            <div className="mr-2"> {comment.userName}</div>
                                        </a>
                                        {
                                            comment.isOwner && (
                                                <Tooltip content="这个是本站的主人捏" side="top" align="center">
                                                    <VscVerifiedFilled color={'blue'}/>
                                                </Tooltip>
                                            )
                                        }
                                        {
                                            comment.isFriend && (
                                                <Tooltip content="这位小伙伴是本站的友链捏" side="top" align="center">
                                                    <VscVerifiedFilled color={'green'}/>
                                                </Tooltip>
                                            )
                                        }
                                        {
                                            comment.isAuthor && (
                                                <Tooltip content="本文作者呀" side="top" align="center">
                                                    <VscVerifiedFilled color={'orange'}/>
                                                </Tooltip>
                                            )
                                        }
                                    </div>
                                );
                            })()}
                        </>
                    ) : (
                        <div className={clsx("flex items-center mr-2")}>
                            <span className="mr-2">
                             {comment.userName}
                            </span>
                            {
                                comment.isOwner && (
                                    <Tooltip content="这个是本站的主人捏" side="top" align="center">
                                        <VscVerifiedFilled color={'blue'}/>
                                    </Tooltip>
                                )
                            }
                            {
                                comment.isFriend && (
                                    <Tooltip content="这位小伙伴是本站的友链捏" side="top" align="center">
                                        <VscVerifiedFilled color={'green'}/>
                                    </Tooltip>
                                )
                            }
                            {
                                comment.isAuthor && (
                                    <Tooltip content="本文作者呀" side="top" align="center">
                                        <VscVerifiedFilled color={'orange'}/>
                                    </Tooltip>
                                )
                            }
                        </div>
                    )}
                    {
                        comment.parentId && (
                            <div className={"flex items-center"} style={{
                                paddingLeft: '0.3rem',
                                fontSize: '0.8rem',
                                color: 'rgba(var(--foreground), 0.7)',
                                lineHeight: '2rem',
                            }}>
                                <Reply width={16} height={16} opacity={"0.5"}/>
                                <span className={"mr-4"}>{comment.parentUserName}</span>
                            </div>
                        )
                    }
                    <div className={"justify-end text-[0.75em] opacity-75"}>{formattedCreatedAt}</div>
                    {comment.isTop && (
                        <PinTopIcon width={16} height={16} className="absolute right-4 top-4"/>
                    )}
                </div>
                <div style={{
                    borderRadius: '0em 0.5em 0.5em 0.5em',
                    padding: '0 0.5rem',
                    margin: '0.5em 0 1em 0',
                    backgroundColor: 'rgba(var(--foreground), 0.05)',
                    border: '1px solid rgba(var(--foreground), 0.1)',
                    fontSize: '0.9rem',
                    width: 'fit-content',
                }}>
                    <ReactMarkdown
                        className={clsx(comment.isDeleted && 'opacity-55')}
                        rehypePlugins={[rehypeSanitize]}
                        remarkPlugins={[remarkGfm]}
                        components={{
                            code({inline, className, children, ...props}) {
                                const match = /language-(\w+)/.exec(className || '');
                                if (!match) {
                                    return <InlineCodeBlock {...props}>{children}</InlineCodeBlock>;
                                }
                                return inline ? (
                                    <InlineCodeBlock {...props}>{children}</InlineCodeBlock>
                                ) : (
                                    <CodeBlock language={match[1]} value={String(children).replace(/\n$/, '')}/>
                                );
                            },
                            a({...props}) {
                                return <Link className={clsx(styles.underlineAnimation, styles.glowAnimation)}
                                             {...props}/>;
                            },
                            p({...props}) {
                                return <p className={'mt-2 mb-2 line'} style={{lineHeight: '1.5'}} {...props} />;
                            },
                            table({...props}) {
                                return <TableView {...props} />;
                            },
                            h1({...props}) {
                                return <h1 className={'mt-4 mb-4'} {...props} />;
                            },
                            h2({...props}) {
                                return <h2 className={'mt-3 mb-3'} {...props} />;
                            },
                            h3({...props}) {
                                return <h3 className={'mt-2 mb-2'} {...props} />;
                            },
                        }}
                    >
                        {comment.content}
                    </ReactMarkdown>
                    <div className="flex justify-between">
                        <div className="flex items-center gap-2">
                            <div className="gap-1 flex">
                                {comment.platform && getPlatformIcon(comment.platform)}
                                {comment.browser && getBrowserIcon(comment.browser)}
                            </div>
                            <span className={"text-[0.75em] opacity-35"}>{comment.location}</span>
                        </div>
                        <div className="flex gap-4 ml-4">
                            {
                                !comment.isDeleted && (
                                    <div>
                                        <Button variant="ghost" size="1" className=""
                                                style={{
                                                    color: "rgb(var(--primary))",
                                                }}
                                                onClick={() => setIsReplying(true)}>
                                            <Reply width={16} height={16}/>
                                            回复 </Button>
                                    </div>
                                )
                            }
                            {isReplying && (
                                <div>
                                    <Button variant="ghost" size="1" className="ml-12"
                                            style={{
                                                color: "rgb(var(--primary))",
                                            }}
                                            onClick={() => setIsReplying(false)}> 取消 </Button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
                {isReplying && <CommentForm id={comment.areaId} parentId={comment.id}/>}
                {comment.children.length > 0 && <CommentList subComments={comment.children}/>}
            </div>
        </motion.div>
    );
};

export default CommentListItem;
