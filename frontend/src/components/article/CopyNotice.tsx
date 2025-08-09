"use client";

import {CalendarIcon, UserIcon, LinkIcon, Link2Icon as LinkIcon2, ExternalLinkIcon} from "lucide-react"
import {usePathname} from "next/navigation";
import {useWebsiteInfo} from "@/app/website-info-provider";

interface CopyrightNoticeProps {
    author?: string
    year?: number
    additionalText?: string
    articleTitle?: string
    permanentLink?: string
}

export default function CopyrightNotice({
                                            author = "作者名称",
                                            year = new Date().getFullYear(),
                                            additionalText = "版权所有，保留部分权利",
                                            articleTitle = "本文",
                                        }: CopyrightNoticeProps) {
    const pathname = usePathname();
    const websiteInfo = useWebsiteInfo()
    const permanentLink = websiteInfo.WEBSITE_URL + pathname;
    return (
        <div className="mt-8 pt-4 text-gray-500 dark:text-gray-400 text-xs">
            <div className="border-t border-gray-100 dark:border-gray-800 pt-4">
                {/* 装饰元素 - 顶部 */}
                <div className="flex items-center mb-4">
                    <div
                        className="text-gray-400 dark:text-gray-500 text-[10px] tracking-wider mr-2 font-light">COPYRIGHT
                    </div>
                    <div className="h-px flex-grow bg-gray-100 dark:bg-gray-800"></div>
                </div>

                <div className="space-y-3">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
                        <div className="space-y-2">
                            <div className="flex items-center gap-1.5">
                                <UserIcon className="h-3 w-3 text-gray-400 dark:text-gray-500"/>
                                <span className="text-gray-400 dark:text-gray-500 w-14 text-[10px]">作者</span>
                                <span className="text-gray-600 dark:text-gray-300">{author}</span>
                            </div>

                            <div className="flex items-center gap-1.5">
                                <CalendarIcon className="h-3 w-3 text-gray-400 dark:text-gray-500"/>
                                <span className="text-gray-400 dark:text-gray-500 w-14 text-[10px]">版权年份</span>
                                <span className="text-gray-600 dark:text-gray-300">© {year}</span>
                            </div>

                            <div className="flex items-center gap-1.5">
                                <LinkIcon2 className="h-3 w-3 text-gray-400 dark:text-gray-500"/>
                                <span className="text-gray-400 dark:text-gray-500 w-14 text-[10px]">永久链接</span>
                                <a
                                    href={permanentLink}
                                    className="text-gray-600 dark:text-gray-300 hover:text-gray-800 dark:hover:text-gray-100 hover:underline break-all leading-tight"
                                >
                                    {permanentLink.replace(/^https?:\/\//, "")}
                                </a>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <div className="flex items-start gap-1.5">
                                <LinkIcon className="h-3 w-3 text-gray-400 dark:text-gray-500 mt-0.5"/>
                                <span className="text-gray-400 dark:text-gray-500 w-14 text-[10px]">许可协议</span>
                                <div>
                                    <div className="flex items-center text-gray-600 dark:text-gray-300">
                                        <div className="flex">
                                            <div
                                                className="h-3.5 w-3.5 rounded-full bg-gray-500 dark:bg-gray-600 flex items-center justify-center text-white text-[7px]">
                                                CC
                                            </div>
                                            <div
                                                className="h-3.5 w-3.5 rounded-full bg-gray-500 dark:bg-gray-600 flex items-center justify-center text-white text-[7px] -ml-0.5">
                                                BY
                                            </div>
                                        </div>
                                        <a
                                            href="https://creativecommons.org/licenses/by/4.0/"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="ml-1.5 hover:underline text-[10px]"
                                        >
                                            知识共享署名 4.0
                                        </a>
                                    </div>
                                    <p className="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5 ml-0 lg:ml-[3.8rem] leading-tight">
                                        《{articleTitle}》采用
                                        <a
                                            href="https://creativecommons.org/licenses/by/4.0/"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 hover:underline inline-flex items-center"
                                        >
                                            知识共享署名 4.0 国际许可协议
                                            <ExternalLinkIcon className="h-2 w-2 ml-0.5"/>
                                        </a>
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div
                        className="text-[10px] text-gray-500 dark:text-gray-400 leading-relaxed pt-2 border-t border-gray-50 dark:border-gray-800">
                        <p className="leading-tight">{additionalText}</p>
                    </div>
                </div>

                {/* 装饰元素 - 底部 */}
                <div className="flex items-center mt-4">
                    <div className="h-px flex-grow bg-gray-100 dark:bg-gray-800"></div>
                    <div className="mx-2 text-gray-200 dark:text-gray-700 text-[10px]">❖</div>
                    <div className="h-px flex-grow bg-gray-100 dark:bg-gray-800"></div>
                </div>
            </div>
        </div>
    )
}

