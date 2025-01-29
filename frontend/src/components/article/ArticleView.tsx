import React from 'react';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import InlineCodeBlock from '@/components/InlineCodeBlock';
import CodeBlock from '@/components/CodeBlock';
import {clsx} from 'clsx';
import styles from '@/styles/PostPage.module.scss';
import ReactMarkdown from 'react-markdown';
import ArticleInlineLink from '@/components/article/ArticleInlineLink';
import ImageView from '@/components/article/ImageView';
import TableView from '@/components/article/TableView';
import Toc from '@/components/article/Toc';
import {article_font} from '@/app/fonts/font';
import ArticleMetadata from '@/components/article/ArticleMetaData';
import FloatingTocMobile from '@/components/article/FloatingTocMobile';
import ArticleScrollSync from '@/components/article/ArticleScrollSync';
import ArticleTopPaddingAnimate from '@/components/article/ArticleTopPaddingAnimate';
import ScrollHandler from '@/components/article/ScrollHandler';
import customCalloutRemarkPlugin from "@/plugins/markdown/Callout";
import remarkDirective from "remark-directive";
import rehypeCallouts from "rehype-callouts";
import 'rehype-callouts/theme/github'
// import rehypeRaw from "rehype-raw";
// import rehypeKatex from "rehype-katex";
import rehypeReact from "rehype-react";
import UserReadReporter from "@/components/article/UserReadReporter";

export type Post = {
    id: string;
    title: string;
    content: string;
    summary: string;
    categoryName: string;
    toc: string;
    authorName: string;
    cover: string;
    views: number;
    likes: number;
    comments: number;
    commentId: string;
    createdAt: string;
    updatedAt: string;
};

// 这里保证生成 id 按照顺序，匹配目录
const generateId = (index: number) => `article-md-title-${index + 1}`;

const ArticleView = ({post}: { post: Post }) => {
    const readingTime = Math.ceil(post.content.length / 500);
    let headingIndex = 0;

    return (
        <div className={clsx(styles.article, article_font.className)}>
            <ScrollHandler/>
            {post.toc && <FloatingTocMobile toc={JSON.parse(post.toc)}/>}
            <div className={styles.articleContainer}>
                <aside className={styles.tocContainer}>
                    {post.toc && <Toc toc={JSON.parse(post.toc)} commentId={post.commentId} likes={post.likes}
                                      comments={post.comments}/>}
                </aside>
                <main className={styles.articleContent}>
                    <ArticleScrollSync post={post} type={"文章"}>
                        <ArticleTopPaddingAnimate/>
                        <h1 className={styles.title}>{post.title}</h1>
                        <ArticleMetadata
                            authorName={post.authorName}
                            createdAt={post.createdAt}
                            updatedAt={post.updatedAt}
                            views={post.views}
                            readingTime={readingTime}
                        />
                        <UserReadReporter articleId={post.id}/>
                        <ReactMarkdown
                            className={styles.markdown}
                            rehypePlugins={[
                                rehypeCallouts,
                                [rehypeReact, {
                                    components: {
                                        InfoBlock: ({children}: { children: React.ReactNode }) => {
                                            console.log("Rendering InfoBlock");
                                            return <div className={styles.infoBlock}>45345{children}</div>;
                                        },
                                    },
                                }],
                                // rehypeRaw,
                                // rehypeSanitize,
                                // rehypeKatex,
                            ]}

                            remarkPlugins={[remarkGfm, remarkBreaks, customCalloutRemarkPlugin, remarkDirective]}
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
                                img({...props}) {
                                    return <ImageView {...props} />;
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
                                    return <h2 id={generateId(headingIndex++)}
                                               className={styles.heading4} {...props} />;
                                },
                                h5({...props}) {
                                    return <h3 id={generateId(headingIndex++)}
                                               className={styles.heading5} {...props} />;
                                },
                                h6({...props}) {
                                    return <h2 id={generateId(headingIndex++)}
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
                            }}
                        >
                            {post.content}
                        </ReactMarkdown>
                    </ArticleScrollSync>
                </main>
            </div>
        </div>
    );
};

export default ArticleView;
