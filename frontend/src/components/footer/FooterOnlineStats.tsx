"use client";

import React, {useState, useEffect} from 'react';
import {useAppSelector} from "@/redux/hooks";
import {article_font} from "@/app/fonts/font";
import {clsx} from "clsx";
import {motion, AnimatePresence} from "framer-motion";
import {PaperclipIcon, UserIcon} from "lucide-react";
import {createPortal} from "react-dom";
import useIsMobile from "@/hooks/useIsMobile";

const FooterLink = () => {
    const onlineCount = useAppSelector(state => state.onlineCount.total);
    const pageViewCount = useAppSelector(state => state.onlineCount.pageView);
    const isMobile = useIsMobile();
    const [mounted, setMounted] = useState(false);

    const [isDetailVisible, setIsDetailVisible] = useState(false);

    // 挂载状态
    useEffect(() => {
        setMounted(true);
    }, []);

    const handleMouseEnter = () => {
        if (!isMobile) {
            setIsDetailVisible(true);
        }
    };

    const handleMouseLeave = () => {
        if (!isMobile) {
            setIsDetailVisible(false);
        }
    };

    const handleClick = () => {
        if (isMobile) {
            setIsDetailVisible(!isDetailVisible);
        }
    };

    return (
        <>
            {/* Portal 渲染的移动端弹框 */}
            {mounted && isMobile && createPortal(
                <AnimatePresence>
                    {isDetailVisible && (
                        <>
                            {/* 移动端遮罩层 */}
                            <motion.div
                                key="backdrop"
                                className="fixed inset-0 bg-black/20 backdrop-blur-sm z-40"
                                initial={{opacity: 0}}
                                animate={{opacity: 1}}
                                exit={{opacity: 0}}
                                onClick={() => setIsDetailVisible(false)}
                            />
                            
                            {/* 移动端弹框内容 */}
                            <motion.div
                                key="mobile-stats"
                                className="fixed bottom-4 left-4 right-4 z-50 bg-white/95 dark:bg-gray-950/95 backdrop-blur-xl border border-gray-200/80 dark:border-white/10 shadow-2xl rounded-2xl p-4"
                                initial={{opacity: 0, y: 100}}
                                animate={{opacity: 1, y: 0}}
                                exit={{opacity: 0, y: 100}}
                                transition={{type: 'spring', stiffness: 100}}
                            >
                                <div className="flex items-center justify-between mb-3">
                                    <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                        访问统计
                                    </h3>
                                    <button
                                        onClick={() => setIsDetailVisible(false)}
                                        className="w-6 h-6 rounded-full bg-gray-500/10 dark:bg-white/10 flex items-center justify-center text-gray-600 dark:text-gray-400 hover:bg-gray-500/20 dark:hover:bg-white/20 transition-colors"
                                    >
                                        ✕
                                    </button>
                                </div>
                                
                                <div className="space-y-2">
                                    {pageViewCount.map((item, index) => (
                                        <div key={index} className="flex justify-between items-center">
                                            <div className="flex items-center">
                                                <PaperclipIcon size={12} className="mr-2 opacity-35"/>
                                                <span className="text-sm text-gray-700 dark:text-gray-300">{item.name}</span>
                                            </div>
                                            <div className="flex items-center">
                                                <UserIcon size={12} className="mr-1 opacity-35"/>
                                                <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{item.count}</span>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </motion.div>
                        </>
                    )}
                </AnimatePresence>,
                document.body
            )}

            {/* 触发器 */}
            <div 
                className={clsx(article_font.className, "relative inline-block")} 
                onMouseEnter={handleMouseEnter}
                onMouseLeave={handleMouseLeave}
                onClick={handleClick}
                style={{
                    cursor: 'pointer',
                    fontSize: '0.8rem',
                }}
            >
                <span> 正在有 {onlineCount} 位小伙伴看着我的网站呐 </span>
                
                {/* 桌面端原有的弹框 */}
                {!isMobile && isDetailVisible && (
                    <motion.div
                        initial={{opacity: 0, y: 20}}
                        animate={{opacity: 1, y: 0}}
                        exit={{opacity: 0, y: 20}}
                        transition={{type: 'spring', stiffness: 100}}
                        style={{
                            position: 'absolute',
                            width: 'max-content',
                            bottom: '100%',
                            left: '50%',
                            transform: 'translateX(-50%) translateY(-2rem)',
                            backgroundColor: 'rgba(var(--background), 0.9)',
                            padding: '0.7rem',
                            borderRadius: '0.5rem',
                            backdropFilter: 'blur(10px)',
                            zIndex: 100,
                            border: '1px solid rgba(var(--foreground), 0.1)',
                        }}
                    >
                        {pageViewCount.map((item, index) => (
                            <div key={index} className="flex" style={{
                                justifyContent: 'space-between',
                                width: '100%',
                            }}>
                                <div className="mr-8">
                                    <PaperclipIcon size={12} className="inline-block mr-2 opacity-35"/>
                                    <span>{item.name}</span>
                                </div>
                                <div>
                                    <UserIcon size={12} className="inline-block mr-1 opacity-35"/>
                                    <span>{item.count}</span>
                                </div>
                            </div>
                        ))}
                    </motion.div>
                )}
            </div>
        </>
    );
};

export default FooterLink;
