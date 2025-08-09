import React from "react"
import AuthorBanner from "@/app/home/AuthorBanner"
import HomePageMomentShow from "@/components/moment/HomePageMomentShow"
import RecentArticle from "@/app/home/recent/RecentArticle"
import {Container} from "@radix-ui/themes"
import RecentMoment from "@/app/home/recent/RecentMoment"
import styles from "@/styles/HomePage.module.scss"
import {getLastFiveArticles} from "@/api/article"
import {getLastFourShare} from "@/api/share"
import type {Article, StatusUpdate} from "@/types"
import RecommendClient from "@/app/home/recommend/RecommendClient";
import {getLatestThinking, Thinking} from "@/api/thinkings";
import FloatingMenu from "@/components/menu/FloatingMenu";

export default async function Home() {
    // 通过传递 revalidate 配置实现增量更新
    const articleList: Article[] = await getLastFiveArticles({next: {revalidate: 60}})
    const shareList: StatusUpdate[] = await getLastFourShare({next: {revalidate: 60}})
    const latestThinking: Thinking = await getLatestThinking({next: {revalidate: 60}})
    return (
        <div>
            <Container size={"4"}>
                <AuthorBanner/>
                <HomePageMomentShow thinking={latestThinking}/>
            </Container>
            
            {/* Recent Content Section */}
            <section className="relative py-12 overflow-hidden">
                {/* Background decoration */}
                <div className="absolute inset-0 bg-gradient-to-b from-background via-foreground/[0.01] to-background" />
                <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(var(--foreground),0.02),transparent_70%)]" />
                
                <div className={"flex justify-center relative"}>
                    <div className={styles.responsiveContainer}>
                        <RecentArticle articleList={articleList}/>
                        <RecentMoment shareList={shareList}/>
                    </div>
                </div>
            </section>
            
            <RecommendClient/>
            <FloatingMenu items={[]}/>
        </div>
    )
}

