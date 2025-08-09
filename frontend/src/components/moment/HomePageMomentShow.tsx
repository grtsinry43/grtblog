"use client";

import React, { useState, useEffect } from 'react';
import styles from "@/styles/moment/HomePageMomentShow.module.scss";
import {motion} from 'framer-motion';
import Image from "next/image";
import quoteFront from "@/assets/quote-front.svg";
import {Thinking} from "@/api/thinkings";

const HomePageMomentShow = ({thinking}: { thinking: Thinking }) => {
    const [isHovered, setIsHovered] = useState(false);
    const [displayText, setDisplayText] = useState('');
    const [isTyping, setIsTyping] = useState(false);

    // 打字机效果
    useEffect(() => {
        if (thinking.content) {
            setIsTyping(true);
            let index = 0;
            const timer = setInterval(() => {
                if (index < thinking.content.length) {
                    setDisplayText(thinking.content.slice(0, index + 1));
                    index++;
                } else {
                    setIsTyping(false);
                    clearInterval(timer);
                }
            }, 80);

            return () => clearInterval(timer);
        }
    }, [thinking.content]);

    return (
        <div className={styles.container}>
            {/* 背景装饰点 */}
            <motion.div 
                className={styles.backgroundDots}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.5, duration: 1 }}
            >
                {[...Array(6)].map((_, i) => (
                    <motion.div
                        key={i}
                        className={styles.dot}
                        initial={{ scale: 0, opacity: 0 }}
                        animate={{ scale: 1, opacity: 0.1 }}
                        transition={{ 
                            delay: 0.7 + i * 0.1, 
                            duration: 0.8,
                            ease: "easeOut"
                        }}
                        style={{
                            left: `${15 + i * 12}%`,
                            top: `${20 + (i % 2) * 60}%`,
                        }}
                    />
                ))}
            </motion.div>

            <motion.div 
                className={styles.content}
                onHoverStart={() => setIsHovered(true)}
                onHoverEnd={() => setIsHovered(false)}
                initial={{ opacity: 0, y: 15 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ 
                    duration: 0.8, 
                    ease: [0.25, 0.46, 0.45, 0.94],
                    delay: 0.1 
                }}
            >
                {/* 引号 */}
                <motion.div
                    className={styles.quotationContainer}
                    initial={{opacity: 0, x: -10}}
                    animate={{opacity: 1, x: 0}}
                    transition={{
                        duration: 0.6,
                        delay: 0.3,
                        ease: "easeOut"
                    }}
                >
                    <motion.div 
                        className={styles.quotationMark}
                        animate={{
                            opacity: isHovered ? 0.4 : 0.2,
                            scale: isHovered ? 1.05 : 1
                        }}
                        transition={{ duration: 0.3 }}
                    >
                        <Image src={quoteFront} alt={"quotation mark"} width={32} height={32}/>
                    </motion.div>
                </motion.div>

                {/* 内容 */}
                <motion.div
                    className={styles.textContainer}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 0.4, duration: 0.6 }}
                >
                    <div className={styles.thinkingText}>
                        <motion.span
                            className={styles.textContent}
                            animate={{
                                color: isHovered ? "rgba(59, 130, 246, 0.8)" : "inherit"
                            }}
                            transition={{ duration: 0.2 }}
                        >
                            {displayText}
                            {isTyping && <motion.span 
                                className={styles.cursor}
                                animate={{ opacity: [1, 0] }}
                                transition={{ duration: 0.8, repeat: Infinity }}
                            >|</motion.span>}
                        </motion.span>
                        
                        {/* 渐变遮罩效果 */}
                        <motion.div 
                            className={styles.textGradient}
                            initial={{ opacity: 0 }}
                            animate={{ opacity: isHovered ? 0.8 : 0 }}
                            transition={{ duration: 0.4 }}
                        />
                    </div>
                    
                    <motion.div 
                        className={styles.metadata}
                        initial={{ opacity: 0, y: 5 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: displayText.length * 0.08 + 0.5, duration: 0.5 }}
                    >
                        <motion.span 
                            className={styles.author}
                            whileHover={{ y: -1 }}
                            transition={{ type: "spring", stiffness: 300 }}
                        >
                            {thinking.author}
                        </motion.span>
                        <span className={styles.divider}>·</span>
                        <span className={styles.date}>
                            {new Date(thinking.createdAt).toLocaleDateString()}
                        </span>
                    </motion.div>
                </motion.div>
            </motion.div>
            
            {/* 滚动提示 - 更微妙 */}
            <motion.div 
                className={styles.scrollHint}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 1.5, duration: 0.6 }}
            >
                <motion.div
                    className={styles.scrollIndicator}
                    animate={{
                        y: [0, 3, 0],
                        opacity: [0.3, 0.6, 0.3]
                    }}
                    transition={{
                        duration: 2.5,
                        repeat: Infinity,
                        ease: "easeInOut"
                    }}
                >
                    <div className={styles.scrollDot} />
                    <div className={styles.scrollDot} />
                    <div className={styles.scrollDot} />
                </motion.div>
                <motion.span
                    className={styles.scrollText}
                    animate={{
                        opacity: [0.2, 0.5, 0.2],
                    }}
                    transition={{
                        duration: 3,
                        repeat: Infinity,
                        ease: "easeInOut",
                        delay: 0.3
                    }}
                >
                    继续阅读
                </motion.span>
            </motion.div>
        </div>
    );
};

export default HomePageMomentShow;
