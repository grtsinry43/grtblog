import React from 'react';
import {getAllTagNames, getArticlesByTag} from "@/api/tag";
import {noto_sans_sc} from '@/app/fonts/font';
import AllPostPageClient from '@/app/posts/AllPostPageClient';
import {notFound} from 'next/navigation';

const pageSize = 10;

interface TagPageProps {
    params: Promise<{ slug: string }>;
}

export async function generateStaticParams() {
    const tags: string[] = await getAllTagNames({next: {revalidate: 60}});
    return tags.map(tag => ({
        slug: tag, // 这是路由匹配的关键参数
    }));
}

export async function generateMetadata({params}: TagPageProps) {
    const {slug} = await params;
    const tags: string[] = await getAllTagNames({next: {revalidate: 60}});
    const tag = tags.find(tag => tag === slug);

    if (!tag) {
        return {
            title: '找不到标签',
            description: '所请求的标签不存在'
        };
    }

    return {
        title: `Tag - ${tag}`,
        description: ` 具有标签 ${tag} 的文章列表 `
    };
}

const ThinkingPage = async ({params}: TagPageProps) => {
    const {slug} = await params;

    // 使用 slug 加载标签名称
    const tags: string[] = await getAllTagNames({next: {revalidate: 60}});
    const tag = tags.find(t => t === slug || encodeURI(t) === slug);

    if (!tag) {
        notFound();
    }

    const initialArticles = await getArticlesByTag(slug, 1, pageSize, {next: {revalidate: 60}});

    return (
        <div style={{maxWidth: '850px', margin: '0 auto', padding: '2em'}}>
            <h1 style={{fontSize: '2em', fontWeight: 'bolder', marginBottom: '1em'}} className={noto_sans_sc.className}>
                Tag - {tag}
            </h1>
            <AllPostPageClient initialArticles={initialArticles} tag={slug}/>
        </div>
    );
};

export default ThinkingPage;
