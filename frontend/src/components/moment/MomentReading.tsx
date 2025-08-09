import React from 'react'
import {Clock, Eye, Heart, MessageSquare, Award, Zap, PenTool} from 'lucide-react'
import {Separator, Badge} from "@radix-ui/themes"
import ReactMarkdown from 'react-markdown'
import {clsx} from "clsx";
import {article_font, title_font, moment_font} from "@/app/fonts/font";
import styles from '@/styles/moment/MomentPage.module.scss'
import ArticleScrollSync from "@/components/article/ArticleScrollSync";
import Toc from "@/components/article/Toc";
import ImageView from "@/components/article/ImageView";
import ArticleInlineLink from "@/components/article/ArticleInlineLink";
import TableView from "@/components/article/TableView";
import InlineCodeBlock from "@/components/InlineCodeBlock";
import CodeBlock from "@/components/CodeBlock";
import ArticleTopPaddingAnimate from "@/components/article/ArticleTopPaddingAnimate";
import remarkGfm from 'remark-gfm';
import BackgroundGrid from "@/components/moment/BackgroundGrid";
import FloatingTocMobile from "@/components/article/FloatingTocMobile";
import AiSummaryBlock from "@/components/article/AISummaryBlock";
import CopyrightNotice from "@/components/article/CopyNotice";

export interface MomentView {
    id: string;
    authorName: string;
    categoryName: string;
    summary: string;
    aiSummary: string;
    toc: string;
    comments: number;
    commentId: string;
    content: string;
    createdAt: string;
    images: string[];
    isHot: boolean;
    isOriginal: boolean;
    isTop: boolean;
    likes: number;
    shortUrl: string;
    title: string;
    updatedAt: string;
    views: number;
}

// 这里保证生成 id 按照顺序，匹配目录
const generateId = (index: number) => `article-md-title-${index + 1}`;

function MomentReadingPage({moment}: { moment: MomentView }) {
    let headingIndex = 0;
    return (
        <div className="mt-4" style={{
            width: "100%",
            maxWidth: "100%",
            scrollBehavior: "smooth",
        }}>
            <ArticleTopPaddingAnimate reverse={true} once={true}/>
            <div className={styles.articleContainer} style={{
                paddingTop: "2em",
                display: "flex",
                gap: "2em"
            }}>
                <div style={{
                    flex: "1",
                    transition: "all 1s",
                    minWidth: "0",
                }}>
                    <ArticleScrollSync post={moment} type={"记录"}>
                        {moment.toc && <FloatingTocMobile type={"page"} targetId={moment.id}
                                                          toc={JSON.parse(moment.toc)}/>}
                        <div
                            className={clsx(styles.moment, "mx-auto bg-white dark:bg-black rounded-lg overflow-hidden")}>
                            <BackgroundGrid/>
                            <article className="px-8 py-12 prose prose-slate dark:prose-invert max-w-none">
                                <h1 className={clsx(title_font.className, "text-3xl font-bold mb-4 text-gray-900 dark:text-gray-100")}>{moment.title}</h1>

                                <div
                                    className={clsx(article_font.className, "flex flex-wrap items-center gap-4 text-sm text-gray-500 dark:text-gray-400 mb-8")}>
                                    {moment.authorName && (
                                        <span className="flex items-center">
                                            <span>{moment.authorName}</span>
                                        </span>
                                    )}
                                    {moment.createdAt && (
                                        <span className="flex items-center">
                                            <Clock className="w-4 h-4 mr-1" aria-hidden="true"/>
                                            <span>{new Date(moment.createdAt).toLocaleDateString()}</span>
                                        </span>
                                    )}
                                    {moment.views !== undefined && (
                                        <span className="flex items-center">
                                            <Eye className="w-4 h-4 mr-1" aria-hidden="true"/>
                                            <span>{moment.views} views</span>
                                        </span>
                                    )}
                                    {moment.likes !== undefined && (
                                        <span className="flex items-center">
                                            <Heart className="w-4 h-4 mr-1" aria-hidden="true"/>
                                            <span>{moment.likes} likes</span>
                                        </span>
                                    )}
                                    {moment.comments !== undefined && (
                                        <span className="flex items-center">
                                            <MessageSquare className="w-4 h-4 mr-1" aria-hidden="true"/>
                                            <span>{moment.comments} comments</span>
                                        </span>
                                    )}
                                </div>

                                <div className={clsx(title_font.className, "flex flex-wrap gap-2 mb-6")}>
                                    {moment.isHot && (
                                        <Badge variant="soft" className="flex items-center gap-1">
                                            <Zap className="w-3 h-3"/>
                                            Hot
                                        </Badge>
                                    )}
                                    {moment.isOriginal && (
                                        <Badge variant="soft" className="flex items-center gap-1">
                                            <PenTool className="w-3 h-3"/>
                                            Original
                                        </Badge>
                                    )}
                                    {moment.isTop && (
                                        <Badge variant="soft" className="flex items-center gap-1">
                                            <Award className="w-3 h-3"/>
                                            Top
                                        </Badge>
                                    )}
                                </div>

                                <Separator className="my-6"/>

                                <div
                                    className={clsx(moment_font.className, "space-y-6 text-gray-800 dark:text-gray-200 leading-relaxed")}>
                                    <AiSummaryBlock aiSummary={moment.aiSummary}/>
                                    <ReactMarkdown
                                        remarkPlugins={[remarkGfm]}
                                        components={{
                                            img({...props}) {
                                                return <ImageView {...props} />;
                                            },
                                            code({inline, className, children, ...props}) {
                                                const match = /language-(\w+)/.exec(className || '');
                                                if (!match) {
                                                    return <InlineCodeBlock {...props}>{children}</InlineCodeBlock>;
                                                }
                                                return inline ? (
                                                    <InlineCodeBlock {...props}>{children}</InlineCodeBlock>
                                                ) : (
                                                    <CodeBlock language={match[1]}
                                                               value={String(children).replace(/\n$/, '')}/>
                                                );
                                            },
                                            a({...props}) {
                                                return (
                                                    <ArticleInlineLink
                                                        className={clsx(styles.underlineAnimation, styles.glowAnimation)}
                                                        {...props}
                                                        linkTitle={props.children}
                                                        linkUrl={props.href}
                                                    />
                                                );
                                            },
                                            p({...props}) {
                                                return <p className={styles.paragraph} {...props} />;
                                            },
                                            table({...props}) {
                                                return <TableView {...props} />;
                                            },
                                            h1({...props}) {
                                                return <h1 id={generateId(headingIndex++)}
                                                           className={styles.heading1} {...props} />;
                                            },
                                            h2({...props}) {
                                                return <h2 id={generateId(headingIndex++)}
                                                           className={styles.heading2} {...props} />;
                                            },
                                            h3({...props}) {
                                                return <h3 id={generateId(headingIndex++)}
                                                           className={styles.heading3} {...props} />;
                                            },
                                            h4({...props}) {
                                                return <h4 id={generateId(headingIndex++)}
                                                           className={styles.heading4} {...props} />;
                                            },
                                            h5({...props}) {
                                                return <h5 id={generateId(headingIndex++)}
                                                           className={styles.heading5} {...props} />;
                                            },
                                            h6({...props}) {
                                                return <h6 id={generateId(headingIndex++)}
                                                           className={styles.heading6} {...props} />;
                                            },
                                            strong({...props}) {
                                                return <strong className={styles.bold} {...props} />;
                                            },
                                            em({...props}) {
                                                return <em className={styles.italic} {...props} />;
                                            },
                                            blockquote({...props}) {
                                                return <blockquote className={styles.blockquote} {...props} />;
                                            },
                                        }}>{moment.content || ''}</ReactMarkdown>
                                </div>

                                {moment.updatedAt && moment.updatedAt !== moment.createdAt && (
                                    <div className="mt-8 text-sm text-gray-500 dark:text-gray-400">
                                        上次更新于: {new Date(moment.updatedAt).toLocaleString()}
                                    </div>
                                )}

                                {
                                    moment.isOriginal && (
                                        <CopyrightNotice
                                            author={moment.authorName}
                                            year={new Date(moment.createdAt).getFullYear()}
                                            additionalText={"转载请注明出处并遵循 CC BY 许可协议条款"}
                                            articleTitle={moment.title}
                                        />
                                    )
                                }
                            </article>
                        </div>
                    </ArticleScrollSync>
                </div>
                <aside style={{
                    position: 'sticky',
                    top: "250px",
                    height: "100%",
                    overflowY: "auto",
                    flex: "0 0 200px",
                    overflowX: "hidden",
                }}>
                    <Toc toc={JSON.parse(moment.toc)} commentId={moment.commentId} likes={moment.likes}
                         targetId={moment.id} type={"moment"}
                         comments={moment.likes}/>
                </aside>
            </div>
        </div>
    )
}

export default MomentReadingPage;
