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
            <div className={"flex justify-center"}>
                <div className={styles.responsiveContainer}>
                    <RecentArticle articleList={articleList}/>
                    <RecentMoment shareList={shareList}/>
                </div>
            </div>
            <RecommendClient/>
        </div>
    )
}

