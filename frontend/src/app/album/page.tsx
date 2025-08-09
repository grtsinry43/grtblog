import React from 'react';
import {TracingBeam} from "@/components/ui/tracing-beam";
import ModernAlbumFlowClient from "@/components/album/ModernAlbumFlowClient";
import {fetchPhotosByPage} from "@/api/photos";
import {noto_sans_sc, noto_serif_sc_bold} from "@/app/fonts/font";
import {clsx} from "clsx";
import FloatingMenu from "@/components/menu/FloatingMenu";

const AlbumPage = async () => {
    const initialImages = await fetchPhotosByPage(1, 12);
    
    return (
        <TracingBeam>
            <div className="relative">
                {/* 精致化标题区域 */}
                <div className="text-center mb-10">
                    <h1 
                        className={clsx(
                            noto_sans_sc.className,
                            "text-2xl md:text-3xl font-semibold text-gray-800 dark:text-gray-100 mb-3"
                        )}
                    >
                        相册
                    </h1>
                    <div className={clsx(
                        noto_serif_sc_bold.className, 
                        'text-gray-500 dark:text-gray-400 text-sm leading-relaxed'
                    )}>
                        每一个精彩的瞬间，都值得被记录
                    </div>
                </div>

                {/* 现代化相册流 */}
                <ModernAlbumFlowClient initialImages={initialImages}/>
                
                <FloatingMenu items={[]}/>
            </div>
        </TracingBeam>
    );
};

export default AlbumPage;
