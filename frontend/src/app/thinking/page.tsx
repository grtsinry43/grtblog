import React from 'react';
import {getThinkingList, Thinking} from "@/api/thinkings";
import ThinkingNote from "@/components/thinking/ThinkingNote";
import {clsx} from "clsx";
import {noto_serif_sc_bold} from "@/app/fonts/font";

const AllThinkingPage = async () => {
    const thinkingData: Thinking[] = await getThinkingList({next: {revalidate: 60}});
    return (
        <div className="min-h-screen p-8">
            <h1 className="text-4xl font-bold text-center mb-4"> 思考 </h1>
            <span
                className={clsx(noto_serif_sc_bold.className, "text-gray-600 text-md block text-center mb-8")}> 所思所想，所言所行，写下丰富的思考 </span>
            <div className="columns-1 sm:columns-2 md:columns-3 lg:columns-4 gap-6 space-y-6 transition-all">
                {thinkingData.map((thinking, index) => (
                    <ThinkingNote key={thinking.id} thinking={thinking} index={index}/>
                ))}
            </div>
        </div>
    );
};

export default AllThinkingPage;
