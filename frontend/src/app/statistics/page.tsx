import BlogActivityGraph from "./activity-graph"
import {getArchives} from "@/api/archive";

export default async function Home() {
    const archives = await getArchives({next: {revalidate: 60}})
    return (
        <div className="w-full max-w-4xl space-y-6 mb-8 mx-auto">
            <BlogActivityGraph data={archives} colorScheme="green"/>
        </div>
    )
}

