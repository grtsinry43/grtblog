import BlogActivityGraph from "./activity-graph"
import {getArchives} from "@/api/archive";

export default async function Home() {
    const archives = await getArchives({next: {revalidate: 60}})
    return (
        <div className="w-full max-w-4xl space-y-6 mb-8 mx-auto">
            <BlogActivityGraph data={archives}
                               quote={{
                                   quote: "写作是思考的另一种体现，通过文字雕刻智慧的图腾，让思想在共鸣中点亮无界的星空。",
                                   author: "grtsinry43",
                                   authorTitle: "前端工程师 / 技术学习者 / 全栈开发者",
                                   authorAvatar: "https://dogeoss.grtsinry43.com/img/author.jpeg"
                               }}
            />
        </div>
    )
}

